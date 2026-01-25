#!/bin/bash
# Build iclip_client_gui using Docker with older glibc for compatibility
# This builds against Ubuntu 20.04 (glibc 2.31) for wider compatibility

set -e

cd "$(dirname "$0")"

IMAGE_NAME="iclip-builder"
DOCKERFILE="Dockerfile.build"

# Create Dockerfile
cat > "$DOCKERFILE" << 'DOCKERFILE_EOF'
FROM ubuntu:20.04

ENV DEBIAN_FRONTEND=noninteractive
ENV http_proxy=http://sercomm.f0g.dev:2802
ENV https_proxy=http://sercomm.f0g.dev:2802
ENV HTTP_PROXY=http://sercomm.f0g.dev:2802
ENV HTTPS_PROXY=http://sercomm.f0g.dev:2802

# Use mirrors and install dependencies
RUN sed -i 's|http://archive.ubuntu.com|http://mirrors.aliyun.com|g' /etc/apt/sources.list && \
    sed -i 's|http://security.ubuntu.com|http://mirrors.aliyun.com|g' /etc/apt/sources.list && \
    apt-get update && \
    apt-get install -y --no-install-recommends \
    wget \
    ca-certificates \
    gcc \
    libc6-dev \
    libgl1-mesa-dev \
    xorg-dev \
    && rm -rf /var/lib/apt/lists/*

# Install Go
RUN wget -q https://go.dev/dl/go1.21.5.linux-amd64.tar.gz \
    && tar -C /usr/local -xzf go1.21.5.linux-amd64.tar.gz \
    && rm go1.21.5.linux-amd64.tar.gz

ENV PATH="/usr/local/go/bin:${PATH}"

WORKDIR /build
DOCKERFILE_EOF

echo "=== Building Docker image (Ubuntu 20.04 with Go) ==="
docker build -t "$IMAGE_NAME" -f "$DOCKERFILE" .

echo "=== Building iclip_client_gui inside Docker ==="
docker run --rm \
    -v "$(pwd)":/build \
    -w /build \
    -e http_proxy=http://sercomm.f0g.dev:2802 \
    -e https_proxy=http://sercomm.f0g.dev:2802 \
    -e GOPROXY=https://goproxy.io,direct \
    "$IMAGE_NAME" \
    bash -c "go mod tidy && go build -v -ldflags='-s -w' -o iclip_client_gui"

# Cleanup
rm -f "$DOCKERFILE"

echo ""
echo "=== Build complete ==="
ls -lh iclip_client_gui
echo ""
echo "This binary should work on systems with glibc 2.31+"
