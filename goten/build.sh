#!/bin/bash

# TEN VAD Go 封装构建脚本

set -e

echo "=== TEN VAD Go 封装构建脚本 ==="

# 检查Go是否安装
if ! command -v go &> /dev/null; then
    echo "错误: 未找到Go编译器，请先安装Go"
    exit 1
fi

echo "Go版本: $(go version)"

# 检查是否在正确的目录
if [ ! -f "go.mod" ]; then
    echo "错误: 请在vad目录下运行此脚本"
    exit 1
fi

# 检查TEN VAD库文件是否存在
echo "检查TEN VAD库文件..."
if [ ! -d "../lib" ]; then
    echo "警告: 未找到../lib目录，请确保TEN VAD库文件已正确安装"
fi

# 清理之前的构建
echo "清理之前的构建..."
rm -rf build/
rm -f vad.test
mkdir -p build/

# 注意：运行测试请使用 ./run_test.sh
echo "开始构建..."

# 构建命令行工具
echo "构建命令行工具..."
go build -o build/vad_demo cmd/main.go

# macOS平台需要patch rpath
if [[ "$OSTYPE" == "darwin"* ]]; then
    echo "macOS平台，patch rpath..."
    # 获取当前脚本所在目录的绝对路径
    SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
    # 计算lib目录的绝对路径
    LIB_PATH="$SCRIPT_DIR/../lib/macOS"
    install_name_tool -add_rpath "$LIB_PATH" build/vad_demo 2>/dev/null || echo "vad_demo rpath已存在"
fi

# 构建库（可选）
# echo "构建库..."
# go build -buildmode=c-shared -o build/libten_vad_go.so .

echo "=== 构建完成 ==="
echo "可执行文件: build/vad_demo"

# 显示使用说明
echo ""
echo "=== 使用说明 ==="
echo "运行命令行工具:"
echo "  ./build/vad_demo -input <WAV文件> -output <结果文件>"
echo ""
echo "查看帮助:"
echo "  ./build/vad_demo -help"
echo ""
echo "显示版本:"
echo "  ./build/vad_demo -version"
echo ""
echo "测试testset音频文件:"
echo "  ./build/vad_demo -input ../testset/testset-audio-01.wav -output test_result.txt" 