package goten

/*
#cgo CFLAGS: -I../include
#cgo linux LDFLAGS: -L../lib/Linux/x64 -lten_vad
#cgo windows,amd64 LDFLAGS: -L../lib/Windows/x64 -lten_vad
#cgo windows,386 LDFLAGS: -L../lib/Windows/x86 -lten_vad
#cgo darwin LDFLAGS: -F/Users/luowei/workspace/ten-vad/lib/macOS -framework ten_vad
#cgo android,arm LDFLAGS: -L../lib/Android/armeabi-v7a -lten_vad
#cgo android,arm64 LDFLAGS: -L../lib/Android/arm64-v8a -lten_vad
#cgo ios LDFLAGS: -L../lib/iOS -F../lib/iOS -framework ten_vad

#include "ten_vad.h"
#include <stdlib.h>
*/
import "C"
import (
	"fmt"
	"unsafe"
)

// VADHandle 表示TEN VAD实例的句柄
type VADHandle struct {
	handle C.ten_vad_handle_t
}

// VADResult 表示VAD处理结果
type VADResult struct {
	Probability float32 // 语音活动概率 [0.0, 1.0]
	Flag        int     // 二进制语音活动决策: 0=无语音, 1=检测到语音
}

// CreateVAD 创建并初始化TEN VAD实例
// hopSize: 两个连续分析帧起始点之间的样本数 (例如: 256)
// threshold: VAD检测阈值，范围[0.0, 1.0]，用于与输出概率比较确定语音活动
func CreateVAD(hopSize int, threshold float32) (*VADHandle, error) {
	var handle C.ten_vad_handle_t

	result := C.ten_vad_create(&handle, C.size_t(hopSize), C.float(threshold))
	if result != 0 {
		return nil, fmt.Errorf("failed to create TEN VAD instance")
	}

	return &VADHandle{handle: handle}, nil
}

// Process 处理一帧音频进行语音活动检测
// audioData: int16_t样本数组，缓冲区长度必须等于CreateVAD时指定的hopSize
func (v *VADHandle) Process(audioData []int16) (*VADResult, error) {
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

	return &VADResult{
		Probability: float32(probability),
		Flag:        int(flag),
	}, nil
}

// Destroy 销毁TEN VAD实例并释放资源
func (v *VADHandle) Destroy() error {
	result := C.ten_vad_destroy(&v.handle)
	if result != 0 {
		return fmt.Errorf("failed to destroy TEN VAD instance")
	}
	return nil
}

// GetVersion 获取TEN VAD库版本字符串
func GetVersion() string {
	return C.GoString(C.ten_vad_get_version())
}

// Close 关闭VAD句柄的便捷方法
func (v *VADHandle) Close() error {
	return v.Destroy()
}
