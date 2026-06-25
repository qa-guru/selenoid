#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
export DOCKER_API_VERSION="${DOCKER_API_VERSION:-1.45}"

SELENOID_BIN="${SELENOID_BIN:-$ROOT/selenoid}"
CONF="${SELENOID_CONF:-$ROOT/config/browsers.json}"
LIMIT="${SELENOID_LIMIT:-5}"

if [[ ! -x "$SELENOID_BIN" ]]; then
  echo "Selenoid binary not found: $SELENOID_BIN" >&2
  echo "Run: go build -o selenoid .  or  ./scripts/build-selenoid.sh" >&2
  exit 1
fi

exec "$SELENOID_BIN" -conf "$CONF" -limit "$LIMIT" "$@"
