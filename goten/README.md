# TEN VAD Go 封装

这是TEN VAD的Go语言cgo封装，使Go开发者能够轻松使用高性能的TEN VAD语音活动检测功能。

## 特性

- 🚀 **高性能**: 基于TEN VAD的高精度语音检测，RTF: 0.0086-0.0570
- 🔧 **易用性**: 简洁的Go API接口，完整的错误处理
- 🌍 **跨平台**: 支持Linux、Windows、macOS、Android、iOS
- ⚡ **实时性**: 快速检测语音活动，减少延迟
- 🎯 **轻量级**: 低计算复杂度和内存占用，库大小: 306KB-731KB
- 📁 **多格式支持**: 支持WAV和PCM音频文件格式

## 项目结构

```
goten/
├── go.mod                    # Go模块定义
├── ten_vad.go               # 主要cgo封装文件
├── audio_utils.go           # 音频处理工具函数
├── example_test.go          # 测试示例
├── build.sh                 # 构建脚本（仅构建）
├── run_test.sh              # 测试脚本（仅测试）
├── cmd/main.go             # 命令行示例程序
```

## 快速开始

### 安装

```bash
git clone https://github.com/TEN-framework/ten-vad.git
cd ten-vad/goten
```

### 基本用法

```go
package main

import (
    "fmt"
    "log"
    "github.com/ten-framework/ten-vad/goten"
)

func main() {
    // 创建VAD实例
    vad, err := goten.CreateVAD(256, 0.5)
    if err != nil {
        log.Fatal(err)
    }
    defer vad.Close()
    
    // 处理音频数据
    audioData := make([]int16, 256)
    // ... 填充音频数据 ...
    
    result, err := vad.Process(audioData)
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("概率: %.6f, 标志: %d\n", result.Probability, result.Flag)
}
```

### 处理音频文件

#### WAV文件

```go
// 处理整个WAV文件
results, err := goten.ProcessWAVFile("input.wav", 256, 0.5)
if err != nil {
    log.Fatal(err)
}

// 分析结果
for i, result := range results {
    if result.Flag == 1 {
        fmt.Printf("帧 %d: 检测到语音 (概率: %.6f)\n", i, result.Probability)
    }
}
```

#### PCM文件

```go
// 配置PCM文件参数
config := goten.PCMConfig{
    SampleRate:    16000,                    // 采样率
    NumChannels:   1,                        // 声道数（单声道）
    BitsPerSample: 16,                       // 位深度
    ByteOrder:     binary.LittleEndian,      // 字节序
}

// 处理PCM文件
results, err := goten.ProcessPCMFile("input.pcm", config, 256, 0.5)
if err != nil {
    log.Fatal(err)
}
```

## API 参考

### 主要函数

- **CreateVAD(hopSize int, threshold float32)**: 创建VAD实例
- **Process(audioData []int16)**: 处理音频帧
- **Destroy()**: 销毁VAD实例
- **GetVersion()**: 获取版本信息
- **ProcessWAVFile(filename, hopSize, threshold)**: 处理WAV文件
- **ProcessPCMFile(filename, config, hopSize, threshold)**: 处理PCM文件
- **DetectFileType(filename)**: 检测文件类型（WAV或PCM）

### 类型

- **VADHandle**: VAD实例句柄
- **VADResult**: 处理结果 (Probability, Flag)
- **PCMConfig**: PCM文件配置 (SampleRate, NumChannels, BitsPerSample, ByteOrder)

## 测试

### Linux/Windows 平台

```bash
go test -v
```

### macOS 平台

由于macOS下framework的动态库加载问题，需要使用patch脚本：

```bash
# 使用测试脚本（推荐）
./run_test.sh

# 或手动测试
go test -c -o goten.test
./goten.test -test.v
```

## 构建和运行

### 构建命令行工具

```bash
# 只构建（推荐）
./build.sh

# 先测试再构建
./run_test.sh && ./build.sh
```

### macOS 平台构建注意事项

`build.sh` 会自动处理macOS的rpath patch，无需手动操作。

### 运行命令行工具

```bash
# 显示帮助
./build/ten_vad_demo -help

# 处理WAV文件
./build/ten_vad_demo -input ../testset/testset-audio-01.wav -output result.txt

# 处理PCM文件（默认采样率=16000Hz, 单声道, 16位, 小端）
./build/ten_vad_demo -input input.pcm -output result.txt

# 显示版本
./build/ten_vad_demo -version
```

### 命令行参数

| 参数 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| `-input` | string | - | 输入音频文件路径（支持WAV和PCM） |
| `-output` | string | - | 输出结果文件路径 |
| `-hop` | int | 256 | 帧大小（样本数） |
| `-threshold` | float64 | 0.5 | VAD检测阈值 [0.0, 1.0] |
| `-version` | bool | false | 显示版本信息 |

#### PCM文件说明
- 仅支持采样率16000Hz、16位、单声道、小端字节序的原始PCM文件。
- 如需其他格式请自行转换或修改源码。

## 跨平台支持

| 平台 | 架构 | 库路径 | 状态 |
|------|------|--------|------|
| Linux | x64 | `../lib/Linux/x64/libten_vad.so` | ✅ 支持 |
| Windows | x64 | `../lib/Windows/x64/ten_vad.dll` | ✅ 支持 |
| Windows | x86 | `../lib/Windows/x86/ten_vad.dll` | ✅ 支持 |
| macOS | - | `../lib/macOS/ten_vad.framework` | ✅ 支持 |
| Android | arm | `../lib/Android/armeabi-v7a/libten_vad.so` | ✅ 支持 |
| Android | arm64 | `../lib/Android/arm64-v8a/libten_vad.so` | ✅ 支持 |
| iOS | - | `../lib/iOS/ten_vad.framework` | ✅ 支持 |

## 音频要求

- **采样率**: 16kHz
- **格式**: 16位PCM
- **通道**: 单声道
- **帧大小**: 建议256样本(16ms)或160样本(10ms)

### 支持的文件格式

#### WAV文件
- 标准RIFF WAV格式
- 16kHz采样率
- 16位PCM编码
- 单声道

#### PCM文件
- 原始PCM数据（无文件头）
- 16kHz采样率
- 16位PCM编码
- 单声道
- 支持小端和大端字节序

## 性能指标

- **RTF**: 0.0086-0.0570 (取决于平台)
- **库大小**: 306KB-731KB (取决于平台)
- **精度**: 优于WebRTC VAD和Silero VAD

## 注意事项

1. **内存管理**: 使用完VAD实例后必须调用`Close()`或`Destroy()`
2. **音频格式**: 确保音频数据符合要求(16kHz, 16位PCM, 单声道)
3. **帧大小**: 音频数据长度必须与创建时的hopSize一致
4. **阈值调整**: 根据应用场景调整threshold参数
5. **macOS运行**: 使用提供的patch脚本解决framework路径问题
6. **PCM文件**: 确保指定正确的字节序和音频参数

## 许可证

本项目遵循Apache 2.0许可证，与TEN VAD项目保持一致。 