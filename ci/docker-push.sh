#!/bin/bash

set -e

DOCKER_IMAGE="${DOCKER_IMAGE:-qaguru/selenoid}"

if [ -z "${DOCKER_USERNAME:-}" ] || [ -z "${DOCKER_PASSWORD:-}" ]; then
	echo "ERROR: Docker push for ${DOCKER_IMAGE} requires DOCKER_USERNAME and DOCKER_PASSWORD repository secrets" >&2
	exit 1
fi

docker login -u="$DOCKER_USERNAME" -p="$DOCKER_PASSWORD"
docker buildx build --pull --push \
	-t "${DOCKER_IMAGE}:${1}" \
	--platform linux/amd64,linux/arm64 .
