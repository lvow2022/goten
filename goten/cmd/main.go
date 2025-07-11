package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"encoding/binary"

	"github.com/ten-framework/ten-vad/goten"
)

func main() {
	// 解析命令行参数
	var (
		inputFile   = flag.String("input", "", "输入音频文件路径 (支持WAV和PCM格式)")
		outputFile  = flag.String("output", "", "输出结果文件路径")
		hopSize     = flag.Int("hop", 256, "帧大小 (样本数)")
		threshold   = flag.Float64("threshold", 0.5, "VAD检测阈值 [0.0, 1.0]")
		showVersion = flag.Bool("version", false, "显示版本信息")
	)
	flag.Parse()

	// 显示版本信息
	if *showVersion {
		fmt.Printf("TEN VAD Version: %s\n", goten.GetVersion())
		return
	}

	// 检查输入文件
	if *inputFile == "" {
		log.Fatal("请指定输入音频文件路径 (-input)")
	}

	// 检查文件是否存在
	if _, err := os.Stat(*inputFile); os.IsNotExist(err) {
		log.Fatalf("输入文件不存在: %s", *inputFile)
	}

	// 设置输出文件
	if *outputFile == "" {
		ext := filepath.Ext(*inputFile)
		base := filepath.Base(*inputFile)
		*outputFile = base[:len(base)-len(ext)] + "_vad_result.txt"
	}

	// 验证参数
	if *hopSize <= 0 {
		log.Fatal("帧大小必须大于0")
	}
	if *threshold < 0.0 || *threshold > 1.0 {
		log.Fatal("阈值必须在 [0.0, 1.0] 范围内")
	}

	fmt.Printf("处理文件: %s\n", *inputFile)
	fmt.Printf("帧大小: %d samples\n", *hopSize)
	fmt.Printf("阈值: %.2f\n", *threshold)
	fmt.Printf("输出文件: %s\n", *outputFile)

	// 检测文件类型
	fileType, err := goten.DetectFileType(*inputFile)
	if err != nil {
		log.Fatalf("检测文件类型失败: %v", err)
	}

	var results []*goten.VADResult

	// 根据文件类型处理
	if fileType == "wav" {
		fmt.Printf("检测到WAV文件\n")
		results, err = goten.ProcessWAVFile(*inputFile, *hopSize, float32(*threshold))
		if err != nil {
			log.Fatalf("处理WAV文件失败: %v", err)
		}
	} else {
		fmt.Printf("检测到PCM文件 (采样率=16000Hz, 单声道, 16位, 小端)\n")
		config := goten.PCMConfig{
			SampleRate:    16000,
			NumChannels:   1,
			BitsPerSample: 16,
			ByteOrder:     binary.LittleEndian,
		}
		results, err = goten.ProcessPCMFile(*inputFile, config, *hopSize, float32(*threshold))
		if err != nil {
			log.Fatalf("处理PCM文件失败: %v", err)
		}
	}

	fmt.Printf("处理完成，共 %d 帧\n", len(results))

	// 统计结果
	speechFrames := 0
	totalProbability := 0.0
	for _, result := range results {
		if result.Flag == 1 {
			speechFrames++
		}
		totalProbability += float64(result.Probability)
	}

	fmt.Printf("语音帧数: %d (%.2f%%)\n", speechFrames, float64(speechFrames)/float64(len(results))*100)
	fmt.Printf("平均概率: %.6f\n", totalProbability/float64(len(results)))

	// 保存结果到文件
	output, err := os.Create(*outputFile)
	if err != nil {
		log.Fatalf("创建输出文件失败: %v", err)
	}
	defer output.Close()

	// 写入头部信息
	fmt.Fprintf(output, "# TEN VAD 处理结果\n")
	fmt.Fprintf(output, "# 输入文件: %s\n", *inputFile)
	fmt.Fprintf(output, "# 文件类型: %s\n", fileType)
	if fileType == "pcm" {
		fmt.Fprintf(output, "# PCM配置: 采样率=16000Hz, 声道数=1, 位深度=16位, 字节序=little\n")
	}
	fmt.Fprintf(output, "# 帧大小: %d samples\n", *hopSize)
	fmt.Fprintf(output, "# 阈值: %.2f\n", *threshold)
	fmt.Fprintf(output, "# 总帧数: %d\n", len(results))
	fmt.Fprintf(output, "# 语音帧数: %d (%.2f%%)\n", speechFrames, float64(speechFrames)/float64(len(results))*100)
	fmt.Fprintf(output, "# 平均概率: %.6f\n", totalProbability/float64(len(results)))
	fmt.Fprintf(output, "# 格式: [帧索引] [概率] [标志]\n")

	// 写入详细结果
	for i, result := range results {
		fmt.Fprintf(output, "[%d] %.6f %d\n", i, result.Probability, result.Flag)
	}

	// 提取语音区间
	var segments [][2]int
	inSpeech := false
	start := 0
	for i, result := range results {
		if result.Flag == 1 && !inSpeech {
			inSpeech = true
			start = i
		}
		if (result.Flag == 0 || i == len(results)-1) && inSpeech {
			inSpeech = false
			end := i
			// 如果最后一帧是语音，end要+1
			if result.Flag == 1 && i == len(results)-1 {
				end = i + 1
			}
			segments = append(segments, [2]int{start, end})
		}
	}

	// 输出语音区间到文件和控制台
	fmt.Fprintf(output, "# 语音区间 (单位: ms)\n")
	fmt.Println("语音区间 (单位: ms):")
	for _, seg := range segments {
		startMs := seg[0] * (*hopSize) * 1000 / 16000
		endMs := seg[1] * (*hopSize) * 1000 / 16000
		frameCount := seg[1] - seg[0]
		fmt.Fprintf(output, "[%d,%d], %d帧\n", startMs, endMs, frameCount)
		fmt.Printf("[%d,%d], %d帧\n", startMs, endMs, frameCount)
	}

	fmt.Printf("结果已保存到: %s\n", *outputFile)
}
