#!/bin/bash

set -euo pipefail

export GO111MODULE="on"

GIT_REVISION="${GIT_REVISION:-$(git describe --tags --always --dirty 2>/dev/null || git rev-parse --short HEAD)}"
BUILD_STAMP="$(date -u '+%Y-%m-%d_%I:%M:%S%p')"
LDFLAGS="-X main.buildStamp=${BUILD_STAMP} -X main.gitRevision=${GIT_REVISION} -s -w"

mkdir -p dist

build() {
	local os=$1 arch=$2
	local out="dist/selenoid_${os}_${arch}"
	echo "Building ${out} (revision=${GIT_REVISION})..."
	CGO_ENABLED=0 GOOS="$os" GOARCH="$arch" go build -ldflags "$LDFLAGS" -o "$out" .
}

build linux amd64
build linux arm64
build darwin amd64
build darwin arm64
build windows amd64
build windows 386

if ! strings dist/selenoid_linux_amd64 | grep -F "$GIT_REVISION" >/dev/null; then
	echo "ERROR: ${GIT_REVISION} not found in dist/selenoid_linux_amd64 — ldflags were not applied" >&2
	exit 1
fi
if ! strings dist/selenoid_linux_amd64 | grep -F "$BUILD_STAMP" >/dev/null; then
	echo "ERROR: ${BUILD_STAMP} not found in dist/selenoid_linux_amd64 — ldflags were not applied" >&2
	exit 1
fi

echo "Build OK: gitRevision=${GIT_REVISION} buildStamp=${BUILD_STAMP}"
