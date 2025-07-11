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
	var (
		inputFile   = flag.String("input", "", "Input audio file path (supports WAV and PCM formats)")
		outputFile  = flag.String("output", "", "Output result file path")
		hopSize     = flag.Int("hop", 256, "Frame size (number of samples)")
		threshold   = flag.Float64("threshold", 0.5, "VAD detection threshold [0.0, 1.0]")
		showVersion = flag.Bool("version", false, "Show version information")
	)
	flag.Parse()

	if *showVersion {
		fmt.Printf("TEN VAD Version: %s\n", goten.GetVersion())
		return
	}

	if *inputFile == "" {
		log.Fatal("Please specify input audio file path (-input)")
	}

	if _, err := os.Stat(*inputFile); os.IsNotExist(err) {
		log.Fatalf("Input file does not exist: %s", *inputFile)
	}

	if *outputFile == "" {
		ext := filepath.Ext(*inputFile)
		base := filepath.Base(*inputFile)
		*outputFile = base[:len(base)-len(ext)] + "_vad_result.txt"
	}

	if *hopSize <= 0 {
		log.Fatal("Frame size must be greater than 0")
	}
	if *threshold < 0.0 || *threshold > 1.0 {
		log.Fatal("Threshold must be in range [0.0, 1.0]")
	}

	fmt.Printf("Processing file: %s\n", *inputFile)
	fmt.Printf("Frame size: %d samples\n", *hopSize)
	fmt.Printf("Threshold: %.2f\n", *threshold)
	fmt.Printf("Output file: %s\n", *outputFile)

	fileType, err := goten.DetectFileType(*inputFile)
	if err != nil {
		log.Fatalf("Failed to detect file type: %v", err)
	}

	var results []*goten.VADResult

	if fileType == "wav" {
		fmt.Printf("Detected WAV file\n")
		results, err = goten.ProcessWAVFile(*inputFile, *hopSize, float32(*threshold))
		if err != nil {
			log.Fatalf("Failed to process WAV file: %v", err)
		}
	} else {
		fmt.Printf("Detected PCM file (sample rate=16000Hz, mono, 16-bit, little-endian)\n")
		config := goten.PCMConfig{
			SampleRate:    16000,
			NumChannels:   1,
			BitsPerSample: 16,
			ByteOrder:     binary.LittleEndian,
		}
		results, err = goten.ProcessPCMFile(*inputFile, config, *hopSize, float32(*threshold))
		if err != nil {
			log.Fatalf("Failed to process PCM file: %v", err)
		}
	}

	fmt.Printf("Processing completed, total %d frames\n", len(results))

	speechFrames := 0
	totalProbability := 0.0
	for _, result := range results {
		if result.Flag == 1 {
			speechFrames++
		}
		totalProbability += float64(result.Probability)
	}

	fmt.Printf("Speech frames: %d (%.2f%%)\n", speechFrames, float64(speechFrames)/float64(len(results))*100)
	fmt.Printf("Average probability: %.6f\n", totalProbability/float64(len(results)))

	output, err := os.Create(*outputFile)
	if err != nil {
		log.Fatalf("Failed to create output file: %v", err)
	}
	defer output.Close()

	fmt.Fprintf(output, "# TEN VAD Processing Results\n")
	fmt.Fprintf(output, "# Input file: %s\n", *inputFile)
	fmt.Fprintf(output, "# File type: %s\n", fileType)
	if fileType == "pcm" {
		fmt.Fprintf(output, "# PCM config: sample rate=16000Hz, channels=1, bit depth=16-bit, byte order=little\n")
	}
	fmt.Fprintf(output, "# Frame size: %d samples\n", *hopSize)
	fmt.Fprintf(output, "# Threshold: %.2f\n", *threshold)
	fmt.Fprintf(output, "# Total frames: %d\n", len(results))
	fmt.Fprintf(output, "# Speech frames: %d (%.2f%%)\n", speechFrames, float64(speechFrames)/float64(len(results))*100)
	fmt.Fprintf(output, "# Format: [frame_index] [probability] [flag]\n")

	for i, result := range results {
		fmt.Fprintf(output, "[%d] %.6f %d\n", i, result.Probability, result.Flag)
	}

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
			// If the last frame is speech, end should be +1
			if result.Flag == 1 && i == len(results)-1 {
				end = i + 1
			}
			segments = append(segments, [2]int{start, end})
		}
	}

	fmt.Fprintf(output, "# Speech segments (unit: ms)\n")
	fmt.Println("Speech segments (unit: ms):")
	for _, seg := range segments {
		startMs := seg[0] * (*hopSize) * 1000 / 16000
		endMs := seg[1] * (*hopSize) * 1000 / 16000
		frameCount := seg[1] - seg[0]
		fmt.Fprintf(output, "[%d,%d], %d frames\n", startMs, endMs, frameCount)
		fmt.Printf("[%d,%d], %d frames\n", startMs, endMs, frameCount)
	}

	fmt.Printf("Results saved to: %s\n", *outputFile)
}
