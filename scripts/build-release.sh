#!/bin/bash

# GitHub Release 构建脚本
# 支持 Linux、macOS、Windows 的 arm64 和 amd64 架构

set -e

# 颜色输出
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# 项目信息
PROJECT_NAME="serve"
VERSION="${1:-dev}"
BUILD_DIR=".build"
BINARY_NAME="${PROJECT_NAME}"

# 清理旧的构建目录
echo -e "${YELLOW}Cleaning old build directory...${NC}"
rm -rf "${BUILD_DIR}"
mkdir -p "${BUILD_DIR}"

# 定义构建目标平台和架构
declare -a TARGETS=(
    "linux/amd64"
    "linux/arm64"
    "darwin/amd64"
    "darwin/arm64"
    "windows/amd64"
    "windows/arm64"
)

# 构建函数
build_target() {
    local os=$1
    local arch=$2
    local output_name="${BINARY_NAME}"
    
    # Windows 平台添加 .exe 后缀
    if [ "${os}" = "windows" ]; then
        output_name="${BINARY_NAME}.exe"
    fi
    
    local output_dir="${BUILD_DIR}/${os}_${arch}"
    local output_path="${output_dir}/${output_name}"
    
    echo -e "${GREEN}Building for ${os}/${arch}...${NC}"
    
    # 设置环境变量并构建
    GOOS=${os} GOARCH=${arch} go build -ldflags "-X main.version=${VERSION}" \
        -o "${output_path}" \
        ./cmd/serve
    
    if [ $? -ne 0 ]; then
        echo -e "${RED}Failed to build for ${os}/${arch}${NC}"
        exit 1
    fi
    
    # 复制 README.md 到输出目录
    if [ -f "README.md" ]; then
        cp README.md "${output_dir}/"
        echo -e "${GREEN}  Copied README.md to output directory${NC}"
    else
        echo -e "${YELLOW}  Warning: README.md not found${NC}"
    fi
    
    # 创建压缩包
    local archive_name="${PROJECT_NAME}_${VERSION}_${os}_${arch}"
    if [ "${os}" = "windows" ]; then
        archive_name="${archive_name}.zip"
        cd "${output_dir}"
        zip -q "../${archive_name}" "${output_name}" README.md
        cd - > /dev/null
    else
        archive_name="${archive_name}.tar.gz"
        tar -czf "${BUILD_DIR}/${archive_name}" -C "${output_dir}" "${output_name}" README.md
    fi
    
    echo -e "${GREEN}✓ Built ${archive_name}${NC}"
}

# 生成校验和文件
generate_checksums() {
    echo -e "${YELLOW}Generating checksums...${NC}"
    cd "${BUILD_DIR}"
    
    # 生成 SHA256 校验和
    shasum -a 256 *.tar.gz *.zip > "checksums.txt" 2>/dev/null || sha256sum *.tar.gz *.zip > "checksums.txt" 2>/dev/null
    
    cd - > /dev/null
    echo -e "${GREEN}✓ Generated checksums.txt${NC}"
}

# 主构建流程
main() {
    echo -e "${GREEN}Starting build for version: ${VERSION}${NC}"
    echo ""
    
    # 遍历所有目标平台
    for target in "${TARGETS[@]}"; do
        IFS='/' read -r os arch <<< "${target}"
        build_target "${os}" "${arch}"
    done
    
    # 生成校验和
    generate_checksums
    
    echo ""
    echo -e "${GREEN}Build completed successfully!${NC}"
    echo -e "${YELLOW}Output directory: ${BUILD_DIR}/${NC}"
    echo ""
    echo "Release files:"
    ls -lh "${BUILD_DIR}"/*.tar.gz "${BUILD_DIR}"/*.zip "${BUILD_DIR}"/checksums.txt 2>/dev/null | awk '{print $9, "(" $5 ")"}'
}

# 运行主函数
main

