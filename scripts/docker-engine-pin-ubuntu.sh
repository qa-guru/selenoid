#!/usr/bin/env bash
# Pin Docker Engine 26.1.5 on Ubuntu/Debian (apt).
# Run on the server once. Requires root.
set -euo pipefail

TARGET_SERIES="26.1"
TARGET_PATCH="${DOCKER_ENGINE_VERSION:-26.1.5}"

if [[ "$(id -u)" -ne 0 ]]; then
  echo "Run as root: sudo $0" >&2
  exit 1
fi

if ! command -v apt-get >/dev/null 2>&1; then
  echo "This script supports apt-based systems only." >&2
  exit 1
fi

echo "==> Current Docker"
docker version 2>/dev/null || true

echo "==> Pinning docker-ce to ${TARGET_PATCH}*"
mkdir -p /etc/apt/preferences.d
cat > /etc/apt/preferences.d/docker-engine-pin <<EOF
Package: docker-ce docker-ce-cli containerd.io docker-buildx-plugin docker-compose-plugin
Pin: version ${TARGET_PATCH}*
Pin-Priority: 1001
EOF

apt-get update
apt-get install -y --allow-downgrades \
  "docker-ce=${TARGET_PATCH}*" \
  "docker-ce-cli=${TARGET_PATCH}*" \
  containerd.io \
  docker-buildx-plugin \
  docker-compose-plugin

echo "==> Hold packages (optional, prevents accidental upgrade)"
apt-mark hold docker-ce docker-ce-cli 2>/dev/null || true

echo "==> Done"
docker version | grep -E 'Version:|API'
echo "Expected: Engine ${TARGET_PATCH}, API 1.55"
