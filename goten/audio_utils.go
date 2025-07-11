package goten

import (
	"encoding/binary"
	"fmt"
	"io"
	"os"
)

// WAVHeader represents WAV file header information
type WAVHeader struct {
	SampleRate    uint32
	NumChannels   uint16
	BitsPerSample uint16
	DataSize      uint32
	DataOffset    int64
}

// ReadWAVFile reads WAV file and returns audio data and header information
func ReadWAVFile(filename string) ([]int16, *WAVHeader, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to open file: %v", err)
	}
	defer file.Close()

	// Read RIFF header
	var riffHeader [12]byte
	if _, err := io.ReadFull(file, riffHeader[:]); err != nil {
		return nil, nil, fmt.Errorf("failed to read RIFF header: %v", err)
	}

	// Check RIFF identifier
	if string(riffHeader[:4]) != "RIFF" {
		return nil, nil, fmt.Errorf("not a valid WAV file (RIFF)")
	}

	if string(riffHeader[8:12]) != "WAVE" {
		return nil, nil, fmt.Errorf("not a valid WAV file (WAVE)")
	}

	header := &WAVHeader{}
	var dataOffset int64 = 12

	// Read sub-chunks
	for {
		var chunkHeader [8]byte
		if _, err := io.ReadFull(file, chunkHeader[:]); err != nil {
			return nil, nil, fmt.Errorf("failed to read chunk header: %v", err)
		}

		chunkID := string(chunkHeader[:4])
		chunkSize := binary.LittleEndian.Uint32(chunkHeader[4:8])

		switch chunkID {
		case "fmt ":
			// Read format information
			var fmtData [16]byte
			if _, err := io.ReadFull(file, fmtData[:]); err != nil {
				return nil, nil, fmt.Errorf("failed to read format data: %v", err)
			}

			header.NumChannels = binary.LittleEndian.Uint16(fmtData[2:4])
			header.SampleRate = binary.LittleEndian.Uint32(fmtData[4:8])
			header.BitsPerSample = binary.LittleEndian.Uint16(fmtData[14:16])

			// Skip remaining data
			if chunkSize > 16 {
				remaining := chunkSize - 16
				if _, err := file.Seek(int64(remaining), io.SeekCurrent); err != nil {
					return nil, nil, fmt.Errorf("failed to skip format data: %v", err)
				}
			}

		case "data":
			// Found data chunk
			header.DataSize = chunkSize
			header.DataOffset = dataOffset + 8
			goto readData

		default:
			// Skip unknown chunks
			if _, err := file.Seek(int64(chunkSize), io.SeekCurrent); err != nil {
				return nil, nil, fmt.Errorf("failed to skip chunk: %v", err)
			}
		}

		dataOffset += 8 + int64(chunkSize)
	}

readData:

	// Check audio format
	if header.NumChannels != 1 {
		return nil, nil, fmt.Errorf("only mono audio is supported, got %d channels", header.NumChannels)
	}

	if header.BitsPerSample != 16 {
		return nil, nil, fmt.Errorf("only 16-bit audio is supported, got %d bits", header.BitsPerSample)
	}

	if header.SampleRate != 16000 {
		return nil, nil, fmt.Errorf("only 16kHz audio is supported, got %d Hz", header.SampleRate)
	}

	// Read audio data
	audioBytes := make([]byte, header.DataSize)
	if _, err := io.ReadFull(file, audioBytes); err != nil {
		return nil, nil, fmt.Errorf("failed to read audio data: %v", err)
	}

	// Convert bytes to int16
	audioData := make([]int16, header.DataSize/2)
	for i := 0; i < len(audioBytes); i += 2 {
		audioData[i/2] = int16(binary.LittleEndian.Uint16(audioBytes[i : i+2]))
	}

	return audioData, header, nil
}

// ProcessWAVFile processes WAV file and returns VAD results
func ProcessWAVFile(filename string, hopSize int, threshold float32) ([]*VADResult, error) {
	// Read WAV file
	audioData, _, err := ReadWAVFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read WAV file: %v", err)
	}

	// Create VAD instance
	vad, err := CreateVAD(hopSize, threshold)
	if err != nil {
		return nil, fmt.Errorf("failed to create VAD: %v", err)
	}
	defer vad.Close()

	frameCount := len(audioData) / hopSize
	results := make([]*VADResult, frameCount)

	for i := 0; i < frameCount; i++ {
		start := i * hopSize
		end := start + hopSize
		if end > len(audioData) {
			end = len(audioData)
		}

		frame := audioData[start:end]

		// If the last frame is smaller than hopSize, pad with zeros
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

// ConvertFloat32ToInt16 converts float32 audio data to int16
func ConvertFloat32ToInt16(floatData []float32) []int16 {
	int16Data := make([]int16, len(floatData))
	for i, sample := range floatData {
		// Convert float32 [-1.0, 1.0] to int16 [-32768, 32767]
		if sample > 1.0 {
			sample = 1.0
		} else if sample < -1.0 {
			sample = -1.0
		}
		int16Data[i] = int16(sample * 32767.0)
	}
	return int16Data
}

// ConvertInt16ToFloat32 converts int16 audio data to float32
func ConvertInt16ToFloat32(int16Data []int16) []float32 {
	floatData := make([]float32, len(int16Data))
	for i, sample := range int16Data {
		// Convert int16 [-32768, 32767] to float32 [-1.0, 1.0]
		floatData[i] = float32(sample) / 32767.0
	}
	return floatData
}

// PCMConfig represents PCM file configuration information
type PCMConfig struct {
	SampleRate    uint32
	NumChannels   uint16
	BitsPerSample uint16
	ByteOrder     binary.ByteOrder
}

// ReadPCMFile reads PCM file and returns audio data
func ReadPCMFile(filename string, config PCMConfig) ([]int16, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %v", err)
	}
	defer file.Close()

	// Check audio format
	if config.NumChannels != 1 {
		return nil, fmt.Errorf("only mono audio is supported, got %d channels", config.NumChannels)
	}

	if config.BitsPerSample != 16 {
		return nil, fmt.Errorf("only 16-bit audio is supported, got %d bits", config.BitsPerSample)
	}

	if config.SampleRate != 16000 {
		return nil, fmt.Errorf("only 16kHz audio is supported, got %d Hz", config.SampleRate)
	}

	audioBytes, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read audio data: %v", err)
	}

	// Check if data length is even (16-bit = 2 bytes)
	if len(audioBytes)%2 != 0 {
		return nil, fmt.Errorf("invalid PCM data length: %d bytes (must be even)", len(audioBytes))
	}

	// Convert bytes to int16
	audioData := make([]int16, len(audioBytes)/2)
	for i := 0; i < len(audioBytes); i += 2 {
		audioData[i/2] = int16(config.ByteOrder.Uint16(audioBytes[i : i+2]))
	}

	return audioData, nil
}

// ProcessPCMFile processes PCM file and returns VAD results
func ProcessPCMFile(filename string, config PCMConfig, hopSize int, threshold float32) ([]*VADResult, error) {
	// Read PCM file
	audioData, err := ReadPCMFile(filename, config)
	if err != nil {
		return nil, fmt.Errorf("failed to read PCM file: %v", err)
	}

	// Create VAD instance
	vad, err := CreateVAD(hopSize, threshold)
	if err != nil {
		return nil, fmt.Errorf("failed to create VAD: %v", err)
	}
	defer vad.Close()

	frameCount := len(audioData) / hopSize
	results := make([]*VADResult, frameCount)

	for i := 0; i < frameCount; i++ {
		start := i * hopSize
		end := start + hopSize
		if end > len(audioData) {
			end = len(audioData)
		}

		frame := audioData[start:end]

		// If the last frame is smaller than hopSize, pad with zeros
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

// DetectFileType detects file type (WAV or PCM)
func DetectFileType(filename string) (string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return "", fmt.Errorf("failed to open file: %v", err)
	}
	defer file.Close()

	fileInfo, err := file.Stat()
	if err != nil {
		return "", fmt.Errorf("failed to get file info: %v", err)
	}

	// If file is too small to determine type, default to PCM
	if fileInfo.Size() < 12 {
		return "pcm", nil
	}

	// Read first 12 bytes
	header := make([]byte, 12)
	if _, err := io.ReadFull(file, header); err != nil {
		return "", fmt.Errorf("failed to read file header: %v", err)
	}

	// Check if it's a WAV file
	if string(header[:4]) == "RIFF" && string(header[8:12]) == "WAVE" {
		return "wav", nil
	}

	// Default to PCM file
	return "pcm", nil
}
