#!/bin/bash

docker version

docker pull "${DOPPELGANGER_TEST_DOCKER_BASE_IMAGE_NAME}" || exit $?

docker build \
    --tag "${DOPPELGANGER_TEST_DOCKER_IMAGE_NAME}" \
    --file scripts/dockerfile_linux \
    scripts || exit $?

docker run \
    --name "${DOPPELGANGER_TEST_DOCKER_CONTAINER_NAME}" \
    --detach \
    "${DOPPELGANGER_TEST_DOCKER_IMAGE_NAME}" || exit $?
