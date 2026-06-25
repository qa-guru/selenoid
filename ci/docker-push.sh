#!/bin/bash

set -e

DOCKER_IMAGE="${DOCKER_IMAGE:-qaguru/selenoid}"

if [ -z "${DOCKER_USERNAME:-}" ] || [ -z "${DOCKER_PASSWORD:-}" ]; then
	echo "Skipping Docker push for ${DOCKER_IMAGE}: DOCKER_USERNAME/DOCKER_PASSWORD not set"
	exit 0
fi

docker login -u="$DOCKER_USERNAME" -p="$DOCKER_PASSWORD"
docker buildx build --pull --push \
	-t "${DOCKER_IMAGE}:${1}" \
	--platform linux/amd64,linux/arm64 .
