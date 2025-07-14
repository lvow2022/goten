package vad

/*
#cgo CFLAGS: -I${SRCDIR}/include
#cgo linux LDFLAGS: -L${SRCDIR}/lib/Linux/x64 -lten_vad
#cgo windows,amd64 LDFLAGS: -L${SRCDIR}/lib/Windows/x64 -lten_vad
#cgo windows,386 LDFLAGS: -L${SRCDIR}/lib/Windows/x86 -lten_vad
#cgo darwin LDFLAGS: -F/Library/Frameworks -framework ten_vad
#cgo android,arm LDFLAGS: -L${SRCDIR}/lib/Android/armeabi-v7a -lten_vad
#cgo android,arm64 LDFLAGS: -L${SRCDIR}/lib/Android/arm64-v8a -lten_vad
#cgo ios LDFLAGS: -L${SRCDIR}/lib/iOS -F${SRCDIR}/lib/iOS -framework ten_vad

#include "ten_vad.h"
#include <stdlib.h>
*/
import "C"
import (
	"fmt"
	"unsafe"
)

// VAD represents a TEN VAD instance
type VAD struct {
	handle C.ten_vad_handle_t
}

// Result represents VAD processing result
type Result struct {
	Probability float32 // Speech activity probability [0.0, 1.0]
	Flag        int     // Binary speech activity decision: 0=no speech, 1=speech detected
}

// New creates and initializes a new TEN VAD instance
// hopSize: number of samples between start points of two consecutive analysis frames (e.g.: 256)
// threshold: VAD detection threshold, range [0.0, 1.0], used to compare with output probability to determine speech activity
func New(hopSize int, threshold float32) (*VAD, error) {
	var handle C.ten_vad_handle_t

	result := C.ten_vad_create(&handle, C.size_t(hopSize), C.float(threshold))
	if result != 0 {
		return nil, fmt.Errorf("failed to create TEN VAD instance")
	}

	return &VAD{handle: handle}, nil
}

// Process processes one frame of audio for speech activity detection
// audioData: int16_t sample array, buffer length must equal the hopSize specified in New
func (v *VAD) Process(audioData []int16) (*Result, error) {
	if len(audioData) == 0 {
		return nil, fmt.Errorf("audio data cannot be empty")
	}

	var probability C.float
	var flag C.int

	result := C.ten_vad_process(
		v.handle,
		(*C.int16_t)(unsafe.Pointer(&audioData[0])),
		C.size_t(len(audioData)),
		&probability,
		&flag,
	)

	if result != 0 {
		return nil, fmt.Errorf("failed to process audio frame")
	}

	return &Result{
		Probability: float32(probability),
		Flag:        int(flag),
	}, nil
}

// Close destroys the TEN VAD instance and releases resources
func (v *VAD) Close() error {
	result := C.ten_vad_destroy(&v.handle)
	if result != 0 {
		return fmt.Errorf("failed to destroy TEN VAD instance")
	}
	return nil
}

// GetVersion gets the TEN VAD library version string
func GetVersion() string {
	return C.GoString(C.ten_vad_get_version())
}
