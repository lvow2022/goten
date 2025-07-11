package goten

import (
	"encoding/binary"
	"fmt"
	"math"
	"os"
	"testing"
)

func TestVADBasicUsage(t *testing.T) {
	// 打印版本信息
	fmt.Printf("TEN VAD Version: %s\n", GetVersion())

	// 创建VAD实例
	// hopSize: 256 samples (16ms at 16kHz)
	// threshold: 0.5 (默认阈值)
	vad, err := CreateVAD(256, 0.5)
	if err != nil {
		t.Fatalf("Failed to create VAD: %v", err)
	}
	defer vad.Close()

	// 模拟音频数据 (256个int16样本)
	// 这里使用随机数据作为示例
	audioData := make([]int16, 256)
	for i := range audioData {
		audioData[i] = int16(i % 1000) // 简单的测试数据
	}

	// 处理音频帧
	result, err := vad.Process(audioData)
	if err != nil {
		t.Fatalf("Failed to process audio: %v", err)
	}

	fmt.Printf("VAD Result - Probability: %.6f, Flag: %d\n",
		result.Probability, result.Flag)

	// 验证结果
	if result.Probability < 0.0 || result.Probability > 1.0 {
		t.Errorf("Probability out of range [0,1]: %f", result.Probability)
	}

	if result.Flag != 0 && result.Flag != 1 {
		t.Errorf("Flag should be 0 or 1, got: %d", result.Flag)
	}
}

func TestVADMultipleFrames(t *testing.T) {
	vad, err := CreateVAD(256, 0.5)
	if err != nil {
		t.Fatalf("Failed to create VAD: %v", err)
	}
	defer vad.Close()

	// 处理多个音频帧
	for frame := 0; frame < 5; frame++ {
		audioData := make([]int16, 256)
		for i := range audioData {
			// 模拟不同的音频模式
			audioData[i] = int16((frame*100 + i) % 2000)
		}

		result, err := vad.Process(audioData)
		if err != nil {
			t.Fatalf("Failed to process frame %d: %v", frame, err)
		}

		fmt.Printf("Frame %d - Probability: %.6f, Flag: %d\n",
			frame, result.Probability, result.Flag)
	}
}

func TestVADErrorHandling(t *testing.T) {
	// 测试空音频数据
	vad, err := CreateVAD(256, 0.5)
	if err != nil {
		t.Fatalf("Failed to create VAD: %v", err)
	}
	defer vad.Close()

	_, err = vad.Process([]int16{})
	if err == nil {
		t.Error("Expected error for empty audio data")
	}

	// 注意：C库对无效参数的处理可能直接报错而不是返回错误代码
	// 这些测试可能在某些情况下会失败，这是正常的
	t.Log("Note: C library may handle invalid parameters differently")
}

func TestPCMFileProcessing(t *testing.T) {
	// 创建测试PCM文件
	testPCMFile := "test_pcm.pcm"
	defer os.Remove(testPCMFile)

	// 生成测试音频数据（1秒的16kHz音频）
	sampleRate := 16000
	duration := 1 // 秒
	numSamples := sampleRate * duration

	// 创建简单的正弦波音频数据
	audioData := make([]int16, numSamples)
	for i := 0; i < numSamples; i++ {
		// 生成440Hz的正弦波
		frequency := 440.0
		amplitude := 0.3
		sample := amplitude * float64(int16(32767)) *
			math.Sin(2*math.Pi*frequency*float64(i)/float64(sampleRate))
		audioData[i] = int16(sample)
	}

	// 写入PCM文件（小端字节序）
	file, err := os.Create(testPCMFile)
	if err != nil {
		t.Fatalf("创建测试PCM文件失败: %v", err)
	}
	defer file.Close()

	// 写入音频数据
	for _, sample := range audioData {
		err := binary.Write(file, binary.LittleEndian, sample)
		if err != nil {
			t.Fatalf("写入PCM数据失败: %v", err)
		}
	}

	// 配置PCM参数
	config := PCMConfig{
		SampleRate:    16000,
		NumChannels:   1,
		BitsPerSample: 16,
		ByteOrder:     binary.LittleEndian,
	}

	// 测试PCM文件读取
	readData, err := ReadPCMFile(testPCMFile, config)
	if err != nil {
		t.Fatalf("读取PCM文件失败: %v", err)
	}

	if len(readData) != len(audioData) {
		t.Fatalf("读取的音频数据长度不匹配: 期望 %d, 实际 %d", len(audioData), len(readData))
	}

	// 测试PCM文件处理
	hopSize := 256
	threshold := float32(0.5)
	results, err := ProcessPCMFile(testPCMFile, config, hopSize, threshold)
	if err != nil {
		t.Fatalf("处理PCM文件失败: %v", err)
	}

	expectedFrames := len(audioData) / hopSize
	if len(results) != expectedFrames {
		t.Fatalf("处理结果帧数不匹配: 期望 %d, 实际 %d", expectedFrames, len(results))
	}

	// 检查结果
	speechFrames := 0
	for _, result := range results {
		if result.Flag == 1 {
			speechFrames++
		}
		if result.Probability < 0.0 || result.Probability > 1.0 {
			t.Fatalf("概率值超出范围 [0,1]: %f", result.Probability)
		}
	}

	fmt.Printf("PCM文件处理测试通过: 总帧数=%d, 语音帧数=%d\n", len(results), speechFrames)
}

func TestFileTypeDetection(t *testing.T) {
	// 测试WAV文件检测
	wavFile := "../testset/testset-audio-01.wav"
	if _, err := os.Stat(wavFile); err == nil {
		fileType, err := DetectFileType(wavFile)
		if err != nil {
			t.Fatalf("检测WAV文件类型失败: %v", err)
		}
		if fileType != "wav" {
			t.Fatalf("WAV文件类型检测错误: 期望 'wav', 实际 '%s'", fileType)
		}
	}

	// 测试PCM文件检测
	testPCMFile := "test_detect.pcm"
	defer os.Remove(testPCMFile)

	// 创建简单的PCM文件（至少12字节）
	file, err := os.Create(testPCMFile)
	if err != nil {
		t.Fatalf("创建测试PCM文件失败: %v", err)
	}
	defer file.Close()

	// 写入一些测试数据（至少6个int16样本 = 12字节）
	testData := []int16{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	for _, sample := range testData {
		binary.Write(file, binary.LittleEndian, sample)
	}

	fileType, err := DetectFileType(testPCMFile)
	if err != nil {
		t.Fatalf("检测PCM文件类型失败: %v", err)
	}
	if fileType != "pcm" {
		t.Fatalf("PCM文件类型检测错误: 期望 'pcm', 实际 '%s'", fileType)
	}

	fmt.Println("文件类型检测测试通过")
}
