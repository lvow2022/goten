package vad

// ProcessWAVFrames processes a WAV file and returns VAD results for each frame
func ProcessWAVFrames(filename string, hopSize int, threshold float32) ([]*Result, error) {
	audioData, _, err := ReadWAVFile(filename)
	if err != nil {
		return nil, err
	}
	vad, err := New(hopSize, threshold)
	if err != nil {
		return nil, err
	}
	defer vad.Close()
	frameCount := len(audioData) / hopSize
	results := make([]*Result, frameCount)
	for i := 0; i < frameCount; i++ {
		start := i * hopSize
		end := start + hopSize
		if end > len(audioData) {
			end = len(audioData)
		}
		frame := audioData[start:end]
		if len(frame) < hopSize {
			padded := make([]int16, hopSize)
			copy(padded, frame)
			frame = padded
		}
		result, err := vad.Process(frame)
		if err != nil {
			return nil, err
		}
		results[i] = result
	}
	return results, nil
}

// ProcessPCMFrames processes a PCM file and returns VAD results for each frame
func ProcessPCMFrames(filename string, config PCMConfig, hopSize int, threshold float32) ([]*Result, error) {
	audioData, err := ReadPCMFile(filename, config)
	if err != nil {
		return nil, err
	}
	vad, err := New(hopSize, threshold)
	if err != nil {
		return nil, err
	}
	defer vad.Close()
	frameCount := len(audioData) / hopSize
	results := make([]*Result, frameCount)
	for i := 0; i < frameCount; i++ {
		start := i * hopSize
		end := start + hopSize
		if end > len(audioData) {
			end = len(audioData)
		}
		frame := audioData[start:end]
		if len(frame) < hopSize {
			padded := make([]int16, hopSize)
			copy(padded, frame)
			frame = padded
		}
		result, err := vad.Process(frame)
		if err != nil {
			return nil, err
		}
		results[i] = result
	}
	return results, nil
}
