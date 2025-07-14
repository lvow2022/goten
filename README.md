# TEN VAD Go Library

本项目是 TEN VAD 的 Go 语言库封装，支持 16kHz 单声道 16bit PCM/WAV 语音活动检测。

## 项目结构

```
ten-vad/
├── vad/                    # VAD 库包
│   ├── vad.go             # 核心 VAD 功能
│   ├── audio.go           # 音频文件处理
│   ├── process.go         # 批量处理功能
│   └── util.go            # 工具函数
├── cmd/                   # 命令行工具
│   └── vad_demo.go        # 命令行程序
├── include/               # C 头文件
├── lib/                   # 动态库文件
├── build.sh               # 构建脚本
├── run_test.sh            # 测试脚本
└── go.mod                 # Go 模块定义
```

## 安装

### 1. 安装动态库文件

在使用此库之前，需要将动态库文件安装到系统目录：

#### Linux 系统
```bash
# 复制动态库到系统目录
sudo cp vad/lib/Linux/x64/libten_vad.so /usr/lib/
sudo chmod 755 /usr/lib/libten_vad.so

# 更新动态库缓存
sudo ldconfig
```

#### macOS 系统
```bash
# 复制框架到系统目录
sudo cp -R vad/lib/macOS/ten_vad.framework /Library/Frameworks/
```

#### Windows 系统
```bash
# 复制动态库到系统目录（需要管理员权限）
copy vad\lib\Windows\x64\ten_vad.dll C:\Windows\System32\
```

### 2. 安装 Go 库

将项目作为 Go module 引入：

```bash
go get github.com/lvow2022/goten
```

## 用法示例

### 作为库使用

```go
package main

import (
    "fmt"
    "github.com/lvow2022/goten/vad"
)

func main() {
    // 处理 WAV 文件
    results, err := vad.ProcessWAVFrames("test.wav", 256, 0.5)
    if err != nil {
        panic(err)
    }
    for i, r := range results {
        fmt.Printf("Frame %d: prob=%.3f, flag=%d\n", i, r.Probability, r.Flag)
    }
}
```

### 命令行工具

```bash
# 构建命令行工具
go build -o vad-demo cmd/vad_demo.go

# 或使用构建脚本
./build.sh

# 处理 WAV 文件
./build/vad_demo -input test.wav -output result.txt

# 处理 PCM 文件
./build/vad_demo -input test.pcm -output result.txt

# 显示版本
./build/vad_demo -version
```

## Makefile 使用说明

- `make build`  构建命令行工具（等价于执行 `./build.sh`），生成 `build/vad_demo` 可执行文件。
- `make test`   运行所有 Go 单元测试（等价于执行 `./run_test.sh`）。
- `make clean`  清理构建产物（删除 `build/` 目录和测试二进制文件）。


## API 说明

### 核心 API

- `vad.New(hopSize int, threshold float32) (*vad.VAD, error)` 创建 VAD 实例
- `vad.(*VAD).Process(audioData []int16) (*vad.Result, error)` 处理一帧音频
- `vad.(*VAD).Close() error` 关闭 VAD 实例

### 文件处理 API

- `vad.ProcessWAVFrames(filename string, hopSize int, threshold float32) ([]*vad.Result, error)` 处理 WAV 文件并返回所有帧的预测结果
- `vad.ProcessPCMFrames(filename string, config vad.PCMConfig, hopSize int, threshold float32) ([]*vad.Result, error)` 处理 PCM 文件并返回所有帧的预测结果

### 音频处理 API

- `vad.ReadWAVFile(filename string) ([]int16, *vad.WAVHeader, error)` 读取 WAV 文件
- `vad.ReadPCMFile(filename string, config vad.PCMConfig) ([]int16, error)` 读取 PCM 文件

## 依赖

- 需要 TEN VAD C/C++ 动态库和头文件，见 `lib/` 和 `include/`

## 许可证

MIT 