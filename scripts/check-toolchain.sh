#!/usr/bin/env bash
set -euo pipefail

EXPECTED_GO_MIN="1.26"
EXPECTED_DOCKER_API="1.45"
EXPECTED_DOCKER_ENGINE_PREFIX="26.1"

fail=0

echo "==> Go (minimum ${EXPECTED_GO_MIN}.x, .go-version / toolchain go1.26.x)"
if command -v go >/dev/null 2>&1; then
  go_v="$(go version)"
  echo "    $go_v"
  if [[ "$go_v" =~ go([0-9]+)\.([0-9]+) ]]; then
    major="${BASH_REMATCH[1]}"
    minor="${BASH_REMATCH[2]}"
    if (( major < 1 || (major == 1 && minor < 26) )); then
      echo "    ERROR: need Go >= ${EXPECTED_GO_MIN}" >&2
      fail=1
    elif (( minor > 26 )); then
      echo "    NOTE: newer than ${EXPECTED_GO_MIN}.x is fine for local builds" >&2
    fi
  fi
else
  echo "    WARN: go not in PATH (build scripts use golang:1.26)" >&2
  fail=1
fi

echo "==> Docker Engine (recommended 26.1.5 / series ${EXPECTED_DOCKER_ENGINE_PREFIX}.x, API ${EXPECTED_DOCKER_API})"
if command -v docker >/dev/null 2>&1; then
  docker version
  engine_v="$(docker version --format '{{.Server.Version}}' 2>/dev/null || true)"
  api_v="$(docker version --format '{{.Server.APIVersion}}' 2>/dev/null || true)"
  if [[ -n "$engine_v" && "$engine_v" != ${EXPECTED_DOCKER_ENGINE_PREFIX}* ]]; then
    echo "    NOTE: Engine $engine_v (not ${EXPECTED_DOCKER_ENGINE_PREFIX}.x) — use DOCKER_API_VERSION=${EXPECTED_DOCKER_API} for hub" >&2
  fi
  if [[ -n "$api_v" && "$api_v" != "$EXPECTED_DOCKER_API" ]]; then
    echo "    NOTE: daemon API $api_v — hub scripts set DOCKER_API_VERSION=${EXPECTED_DOCKER_API}" >&2
  fi
else
  echo "    ERROR: docker not found" >&2
  fail=1
fi

echo "==> Hub env"
echo "    DOCKER_API_VERSION=${DOCKER_API_VERSION:-<unset, start-selenoid.sh sets 1.45>}"

if [[ "$fail" -ne 0 ]]; then
  exit 1
fi

echo "OK"
