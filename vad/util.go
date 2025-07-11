package vad

// ConvertFloat32ToInt16 converts float32 audio data to int16
func ConvertFloat32ToInt16(floatData []float32) []int16 {
	int16Data := make([]int16, len(floatData))
	for i, sample := range floatData {
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
		floatData[i] = float32(sample) / 32767.0
	}
	return floatData
}
