package vad

import (
	"encoding/binary"
	"fmt"
	"io"
	"os"
)

type WAVHeader struct {
	SampleRate    uint32
	NumChannels   uint16
	BitsPerSample uint16
	DataSize      uint32
	DataOffset    int64
}

type PCMConfig struct {
	SampleRate    uint32
	NumChannels   uint16
	BitsPerSample uint16
	ByteOrder     binary.ByteOrder
}

// ReadWAVFile reads WAV file and returns audio data and header information
func ReadWAVFile(filename string) ([]int16, *WAVHeader, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to open file: %v", err)
	}
	defer file.Close()

	var riffHeader [12]byte
	if _, err := io.ReadFull(file, riffHeader[:]); err != nil {
		return nil, nil, fmt.Errorf("failed to read RIFF header: %v", err)
	}

	if string(riffHeader[:4]) != "RIFF" {
		return nil, nil, fmt.Errorf("not a valid WAV file (RIFF)")
	}
	if string(riffHeader[8:12]) != "WAVE" {
		return nil, nil, fmt.Errorf("not a valid WAV file (WAVE)")
	}

	header := &WAVHeader{}
	var dataOffset int64 = 12

	for {
		var chunkHeader [8]byte
		if _, err := io.ReadFull(file, chunkHeader[:]); err != nil {
			return nil, nil, fmt.Errorf("failed to read chunk header: %v", err)
		}
		chunkID := string(chunkHeader[:4])
		chunkSize := binary.LittleEndian.Uint32(chunkHeader[4:8])
		switch chunkID {
		case "fmt ":
			var fmtData [16]byte
			if _, err := io.ReadFull(file, fmtData[:]); err != nil {
				return nil, nil, fmt.Errorf("failed to read format data: %v", err)
			}
			header.NumChannels = binary.LittleEndian.Uint16(fmtData[2:4])
			header.SampleRate = binary.LittleEndian.Uint32(fmtData[4:8])
			header.BitsPerSample = binary.LittleEndian.Uint16(fmtData[14:16])
			if chunkSize > 16 {
				remaining := chunkSize - 16
				if _, err := file.Seek(int64(remaining), io.SeekCurrent); err != nil {
					return nil, nil, fmt.Errorf("failed to skip format data: %v", err)
				}
			}
		case "data":
			header.DataSize = chunkSize
			header.DataOffset = dataOffset + 8
			goto readData
		default:
			if _, err := file.Seek(int64(chunkSize), io.SeekCurrent); err != nil {
				return nil, nil, fmt.Errorf("failed to skip chunk: %v", err)
			}
		}
		dataOffset += 8 + int64(chunkSize)
	}

readData:
	if header.NumChannels != 1 {
		return nil, nil, fmt.Errorf("only mono audio is supported, got %d channels", header.NumChannels)
	}
	if header.BitsPerSample != 16 {
		return nil, nil, fmt.Errorf("only 16-bit audio is supported, got %d bits", header.BitsPerSample)
	}
	if header.SampleRate != 16000 {
		return nil, nil, fmt.Errorf("only 16kHz audio is supported, got %d Hz", header.SampleRate)
	}
	audioBytes := make([]byte, header.DataSize)
	if _, err := io.ReadFull(file, audioBytes); err != nil {
		return nil, nil, fmt.Errorf("failed to read audio data: %v", err)
	}
	audioData := make([]int16, header.DataSize/2)
	for i := 0; i < len(audioBytes); i += 2 {
		audioData[i/2] = int16(binary.LittleEndian.Uint16(audioBytes[i : i+2]))
	}
	return audioData, header, nil
}

// ReadPCMFile reads PCM file and returns audio data
func ReadPCMFile(filename string, config PCMConfig) ([]int16, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open PCM file: %v", err)
	}
	defer file.Close()
	fileInfo, err := file.Stat()
	if err != nil {
		return nil, fmt.Errorf("failed to stat PCM file: %v", err)
	}
	dataSize := fileInfo.Size()
	if dataSize%2 != 0 {
		return nil, fmt.Errorf("PCM file size is not aligned to 16-bit samples")
	}
	audioBytes := make([]byte, dataSize)
	if _, err := io.ReadFull(file, audioBytes); err != nil {
		return nil, fmt.Errorf("failed to read PCM data: %v", err)
	}
	audioData := make([]int16, dataSize/2)
	for i := 0; i < len(audioBytes); i += 2 {
		sample := config.ByteOrder.Uint16(audioBytes[i : i+2])
		audioData[i/2] = int16(sample)
	}
	return audioData, nil
}
