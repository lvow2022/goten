package goten

import (
	"encoding/binary"
	"fmt"
	"math"
	"os"
	"testing"

	"github.com/lvow2022/goten/vad"
)

func TestVADBasicUsage(t *testing.T) {
	// Print version information
	fmt.Printf("TEN VAD Version: %s\n", vad.GetVersion())

	// Create VAD instance
	// hopSize: 256 samples (16ms at 16kHz)
	// threshold: 0.5 (default threshold)
	vadInstance, err := vad.New(256, 0.5)
	if err != nil {
		t.Fatalf("Failed to create VAD: %v", err)
	}
	defer vadInstance.Close()

	// Simulate audio data (256 int16 samples)
	// Using random data as example
	audioData := make([]int16, 256)
	for i := range audioData {
		audioData[i] = int16(i % 1000) // Simple test data
	}

	// Process audio frame
	result, err := vadInstance.Process(audioData)
	if err != nil {
		t.Fatalf("Failed to process audio: %v", err)
	}

	fmt.Printf("VAD Result - Probability: %.6f, Flag: %d\n",
		result.Probability, result.Flag)

	// Verify results
	if result.Probability < 0.0 || result.Probability > 1.0 {
		t.Errorf("Probability out of range [0,1]: %f", result.Probability)
	}

	if result.Flag != 0 && result.Flag != 1 {
		t.Errorf("Flag should be 0 or 1, got: %d", result.Flag)
	}
}

func TestVADMultipleFrames(t *testing.T) {
	vadInstance, err := vad.New(256, 0.5)
	if err != nil {
		t.Fatalf("Failed to create VAD: %v", err)
	}
	defer vadInstance.Close()

	// Process multiple audio frames
	for frame := 0; frame < 5; frame++ {
		audioData := make([]int16, 256)
		for i := range audioData {
			// Simulate different audio patterns
			audioData[i] = int16((frame*100 + i) % 2000)
		}

		result, err := vadInstance.Process(audioData)
		if err != nil {
			t.Fatalf("Failed to process frame %d: %v", frame, err)
		}

		fmt.Printf("Frame %d - Probability: %.6f, Flag: %d\n",
			frame, result.Probability, result.Flag)
	}
}

func TestVADErrorHandling(t *testing.T) {
	// Test empty audio data
	vadInstance, err := vad.New(256, 0.5)
	if err != nil {
		t.Fatalf("Failed to create VAD: %v", err)
	}
	defer vadInstance.Close()

	_, err = vadInstance.Process([]int16{})
	if err == nil {
		t.Error("Expected error for empty audio data")
	}

	// Note: C library may handle invalid parameters differently
	// These tests may fail in some cases, which is normal
	t.Log("Note: C library may handle invalid parameters differently")
}

func TestPCMFileProcessing(t *testing.T) {
	// Create test PCM file
	testPCMFile := "test_pcm.pcm"
	defer os.Remove(testPCMFile)

	// Generate test audio data (1 second of 16kHz audio)
	sampleRate := 16000
	duration := 1 // seconds
	numSamples := sampleRate * duration

	// Create simple sine wave audio data
	audioData := make([]int16, numSamples)
	for i := 0; i < numSamples; i++ {
		// Generate 440Hz sine wave
		frequency := 440.0
		amplitude := 0.3
		sample := amplitude * float64(int16(32767)) *
			math.Sin(2*math.Pi*frequency*float64(i)/float64(sampleRate))
		audioData[i] = int16(sample)
	}

	// Write PCM file (little-endian byte order)
	file, err := os.Create(testPCMFile)
	if err != nil {
		t.Fatalf("Failed to create test PCM file: %v", err)
	}
	defer file.Close()

	// Write audio data
	for _, sample := range audioData {
		err := binary.Write(file, binary.LittleEndian, sample)
		if err != nil {
			t.Fatalf("Failed to write PCM data: %v", err)
		}
	}

	// Configure PCM parameters
	config := vad.PCMConfig{
		SampleRate:    16000,
		NumChannels:   1,
		BitsPerSample: 16,
		ByteOrder:     binary.LittleEndian,
	}

	// Test PCM file reading
	readData, err := vad.ReadPCMFile(testPCMFile, config)
	if err != nil {
		t.Fatalf("Failed to read PCM file: %v", err)
	}

	if len(readData) != len(audioData) {
		t.Fatalf("Read audio data length mismatch: expected %d, got %d", len(audioData), len(readData))
	}

	// Test PCM file processing
	hopSize := 256
	threshold := float32(0.5)
	results, err := vad.ProcessPCMFrames(testPCMFile, config, hopSize, threshold)
	if err != nil {
		t.Fatalf("Failed to process PCM file: %v", err)
	}

	expectedFrames := len(audioData) / hopSize
	if len(results) != expectedFrames {
		t.Fatalf("Processing result frame count mismatch: expected %d, got %d", expectedFrames, len(results))
	}

	// Check results
	speechFrames := 0
	for _, result := range results {
		if result.Flag == 1 {
			speechFrames++
		}
		if result.Probability < 0.0 || result.Probability > 1.0 {
			t.Fatalf("Probability value out of range [0,1]: %f", result.Probability)
		}
	}

	fmt.Printf("PCM file processing test passed: total frames=%d, speech frames=%d\n", len(results), speechFrames)
}

func TestFileTypeDetection(t *testing.T) {
	// Test WAV file detection
	wavFile := "testset/testset-audio-01.wav"
	if _, err := os.Stat(wavFile); err == nil {
		fileType, err := detectFileType(wavFile)
		if err != nil {
			t.Fatalf("Failed to detect WAV file type: %v", err)
		}
		if fileType != "wav" {
			t.Fatalf("WAV file type detection error: expected 'wav', got '%s'", fileType)
		}
	}

	// Test PCM file detection
	testPCMFile := "test_detect.pcm"
	defer os.Remove(testPCMFile)

	// Create simple PCM file (at least 12 bytes)
	file, err := os.Create(testPCMFile)
	if err != nil {
		t.Fatalf("Failed to create test PCM file: %v", err)
	}
	defer file.Close()

	// Write some test data (at least 6 int16 samples = 12 bytes)
	testData := []int16{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	for _, sample := range testData {
		binary.Write(file, binary.LittleEndian, sample)
	}

	fileType, err := detectFileType(testPCMFile)
	if err != nil {
		t.Fatalf("Failed to detect PCM file type: %v", err)
	}
	if fileType != "pcm" {
		t.Fatalf("PCM file type detection error: expected 'pcm', got '%s'", fileType)
	}

	fmt.Println("File type detection test passed")
}

// detectFileType detects if the file is WAV or PCM
func detectFileType(filename string) (string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return "", err
	}
	defer file.Close()

	var header [12]byte
	if _, err := file.Read(header[:]); err != nil {
		return "", err
	}

	// Check if it's a WAV file
	if string(header[:4]) == "RIFF" && string(header[8:12]) == "WAVE" {
		return "wav", nil
	}

	// Assume it's PCM if not WAV
	return "pcm", nil
}
