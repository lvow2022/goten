package goten

import (
	"encoding/binary"
	"fmt"
	"io"
	"os"
)

// WAVHeader 表示WAV文件头信息
type WAVHeader struct {
	SampleRate    uint32
	NumChannels   uint16
	BitsPerSample uint16
	DataSize      uint32
	DataOffset    int64
}

// ReadWAVFile 读取WAV文件并返回音频数据和头信息
func ReadWAVFile(filename string) ([]int16, *WAVHeader, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to open file: %v", err)
	}
	defer file.Close()

	// 读取RIFF头
	var riffHeader [12]byte
	if _, err := io.ReadFull(file, riffHeader[:]); err != nil {
		return nil, nil, fmt.Errorf("failed to read RIFF header: %v", err)
	}

	// 检查RIFF标识
	if string(riffHeader[:4]) != "RIFF" {
		return nil, nil, fmt.Errorf("not a valid WAV file (RIFF)")
	}

	if string(riffHeader[8:12]) != "WAVE" {
		return nil, nil, fmt.Errorf("not a valid WAV file (WAVE)")
	}

	header := &WAVHeader{}
	var dataOffset int64 = 12

	// 读取子块
	for {
		var chunkHeader [8]byte
		if _, err := io.ReadFull(file, chunkHeader[:]); err != nil {
			return nil, nil, fmt.Errorf("failed to read chunk header: %v", err)
		}

		chunkID := string(chunkHeader[:4])
		chunkSize := binary.LittleEndian.Uint32(chunkHeader[4:8])

		switch chunkID {
		case "fmt ":
			// 读取格式信息
			var fmtData [16]byte
			if _, err := io.ReadFull(file, fmtData[:]); err != nil {
				return nil, nil, fmt.Errorf("failed to read format data: %v", err)
			}

			header.NumChannels = binary.LittleEndian.Uint16(fmtData[2:4])
			header.SampleRate = binary.LittleEndian.Uint32(fmtData[4:8])
			header.BitsPerSample = binary.LittleEndian.Uint16(fmtData[14:16])

			// 跳过剩余数据
			if chunkSize > 16 {
				remaining := chunkSize - 16
				if _, err := file.Seek(int64(remaining), io.SeekCurrent); err != nil {
					return nil, nil, fmt.Errorf("failed to skip format data: %v", err)
				}
			}

		case "data":
			// 找到数据块
			header.DataSize = chunkSize
			header.DataOffset = dataOffset + 8
			goto readData

		default:
			// 跳过未知块
			if _, err := file.Seek(int64(chunkSize), io.SeekCurrent); err != nil {
				return nil, nil, fmt.Errorf("failed to skip chunk: %v", err)
			}
		}

		dataOffset += 8 + int64(chunkSize)
	}

readData:

	// 检查音频格式
	if header.NumChannels != 1 {
		return nil, nil, fmt.Errorf("only mono audio is supported, got %d channels", header.NumChannels)
	}

	if header.BitsPerSample != 16 {
		return nil, nil, fmt.Errorf("only 16-bit audio is supported, got %d bits", header.BitsPerSample)
	}

	if header.SampleRate != 16000 {
		return nil, nil, fmt.Errorf("only 16kHz audio is supported, got %d Hz", header.SampleRate)
	}

	// 读取音频数据
	audioBytes := make([]byte, header.DataSize)
	if _, err := io.ReadFull(file, audioBytes); err != nil {
		return nil, nil, fmt.Errorf("failed to read audio data: %v", err)
	}

	// 将字节转换为int16
	audioData := make([]int16, header.DataSize/2)
	for i := 0; i < len(audioBytes); i += 2 {
		audioData[i/2] = int16(binary.LittleEndian.Uint16(audioBytes[i : i+2]))
	}

	return audioData, header, nil
}

// ProcessWAVFile 处理WAV文件并返回VAD结果
func ProcessWAVFile(filename string, hopSize int, threshold float32) ([]*VADResult, error) {
	// 读取WAV文件
	audioData, _, err := ReadWAVFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read WAV file: %v", err)
	}

	// 创建VAD实例
	vad, err := CreateVAD(hopSize, threshold)
	if err != nil {
		return nil, fmt.Errorf("failed to create VAD: %v", err)
	}
	defer vad.Close()

	// 计算帧数
	frameCount := len(audioData) / hopSize
	results := make([]*VADResult, frameCount)

	// 处理每一帧
	for i := 0; i < frameCount; i++ {
		start := i * hopSize
		end := start + hopSize
		if end > len(audioData) {
			end = len(audioData)
		}

		frame := audioData[start:end]

		// 如果最后一帧不足hopSize，用零填充
		if len(frame) < hopSize {
			paddedFrame := make([]int16, hopSize)
			copy(paddedFrame, frame)
			frame = paddedFrame
		}

		result, err := vad.Process(frame)
		if err != nil {
			return nil, fmt.Errorf("failed to process frame %d: %v", i, err)
		}

		results[i] = result
	}

	return results, nil
}

// ConvertFloat32ToInt16 将float32音频数据转换为int16
func ConvertFloat32ToInt16(floatData []float32) []int16 {
	int16Data := make([]int16, len(floatData))
	for i, sample := range floatData {
		// 将float32 [-1.0, 1.0] 转换为int16 [-32768, 32767]
		if sample > 1.0 {
			sample = 1.0
		} else if sample < -1.0 {
			sample = -1.0
		}
		int16Data[i] = int16(sample * 32767.0)
	}
	return int16Data
}

// ConvertInt16ToFloat32 将int16音频数据转换为float32
func ConvertInt16ToFloat32(int16Data []int16) []float32 {
	floatData := make([]float32, len(int16Data))
	for i, sample := range int16Data {
		// 将int16 [-32768, 32767] 转换为float32 [-1.0, 1.0]
		floatData[i] = float32(sample) / 32767.0
	}
	return floatData
}

// PCMConfig 表示PCM文件的配置信息
type PCMConfig struct {
	SampleRate    uint32
	NumChannels   uint16
	BitsPerSample uint16
	ByteOrder     binary.ByteOrder
}

// ReadPCMFile 读取PCM文件并返回音频数据
func ReadPCMFile(filename string, config PCMConfig) ([]int16, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %v", err)
	}
	defer file.Close()

	// 检查音频格式
	if config.NumChannels != 1 {
		return nil, fmt.Errorf("only mono audio is supported, got %d channels", config.NumChannels)
	}

	if config.BitsPerSample != 16 {
		return nil, fmt.Errorf("only 16-bit audio is supported, got %d bits", config.BitsPerSample)
	}

	if config.SampleRate != 16000 {
		return nil, fmt.Errorf("only 16kHz audio is supported, got %d Hz", config.SampleRate)
	}

	// 读取所有音频数据
	audioBytes, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read audio data: %v", err)
	}

	// 检查数据长度是否为偶数（16位 = 2字节）
	if len(audioBytes)%2 != 0 {
		return nil, fmt.Errorf("invalid PCM data length: %d bytes (must be even)", len(audioBytes))
	}

	// 将字节转换为int16
	audioData := make([]int16, len(audioBytes)/2)
	for i := 0; i < len(audioBytes); i += 2 {
		audioData[i/2] = int16(config.ByteOrder.Uint16(audioBytes[i : i+2]))
	}

	return audioData, nil
}

// ProcessPCMFile 处理PCM文件并返回VAD结果
func ProcessPCMFile(filename string, config PCMConfig, hopSize int, threshold float32) ([]*VADResult, error) {
	// 读取PCM文件
	audioData, err := ReadPCMFile(filename, config)
	if err != nil {
		return nil, fmt.Errorf("failed to read PCM file: %v", err)
	}

	// 创建VAD实例
	vad, err := CreateVAD(hopSize, threshold)
	if err != nil {
		return nil, fmt.Errorf("failed to create VAD: %v", err)
	}
	defer vad.Close()

	// 计算帧数
	frameCount := len(audioData) / hopSize
	results := make([]*VADResult, frameCount)

	// 处理每一帧
	for i := 0; i < frameCount; i++ {
		start := i * hopSize
		end := start + hopSize
		if end > len(audioData) {
			end = len(audioData)
		}

		frame := audioData[start:end]

		// 如果最后一帧不足hopSize，用零填充
		if len(frame) < hopSize {
			paddedFrame := make([]int16, hopSize)
			copy(paddedFrame, frame)
			frame = paddedFrame
		}

		result, err := vad.Process(frame)
		if err != nil {
			return nil, fmt.Errorf("failed to process frame %d: %v", i, err)
		}

		results[i] = result
	}

	return results, nil
}

// DetectFileType 检测文件类型（WAV或PCM）
func DetectFileType(filename string) (string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return "", fmt.Errorf("failed to open file: %v", err)
	}
	defer file.Close()

	// 获取文件大小
	fileInfo, err := file.Stat()
	if err != nil {
		return "", fmt.Errorf("failed to get file info: %v", err)
	}

	// 如果文件太小，无法确定类型，默认为PCM
	if fileInfo.Size() < 12 {
		return "pcm", nil
	}

	// 读取前12字节
	header := make([]byte, 12)
	if _, err := io.ReadFull(file, header); err != nil {
		return "", fmt.Errorf("failed to read file header: %v", err)
	}

	// 检查是否为WAV文件
	if string(header[:4]) == "RIFF" && string(header[8:12]) == "WAVE" {
		return "wav", nil
	}

	// 默认为PCM文件
	return "pcm", nil
}
