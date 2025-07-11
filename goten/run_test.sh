#!/bin/bash

set -e

echo "=== 编译 Go 测试二进制 ==="
go test -c -o goten.test

# 检测平台
if [[ "$OSTYPE" == "darwin"* ]]; then
    echo "=== 检测到macOS，patch rpath 到 ten_vad.framework ==="
    # 获取当前脚本所在目录的绝对路径
    SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
    # 计算lib目录的绝对路径
    LIB_PATH="$SCRIPT_DIR/../lib/macOS"
    install_name_tool -add_rpath "$LIB_PATH" ./goten.test 2>/dev/null || echo "rpath已存在"
else
    echo "=== 检测到Linux/Windows，直接运行测试 ==="
fi

echo "=== 运行测试 ==="
./goten.test -test.v