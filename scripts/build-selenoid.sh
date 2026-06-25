#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
OUT="${SELENOID_OUT:-$ROOT/selenoid}"

if command -v go >/dev/null 2>&1; then
  (cd "$ROOT" && go build -o "$OUT" .)
  echo "Built $OUT"
  exit 0
fi

if ! command -v docker >/dev/null 2>&1; then
  echo "Install Go 1.23+ or Docker to build Selenoid" >&2
  exit 1
fi

docker run --rm \
  -v "$ROOT:/src" \
  -w /src \
  golang:1.23 \
  go build -o /src/selenoid .

echo "Built $OUT via Docker"
