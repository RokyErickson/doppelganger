#!/bin/bash

stop "${DOPPELGANGER_TEST_DOCKER_CONTAINER_NAME}" || exit $?

docker container prune --force || exit $?

docker image rm --force "${DOPPELGANGER_TEST_DOCKER_IMAGE_NAME}" || exit $?
docker image rm --force "${DOPPELGANGER_TEST_DOCKER_BASE_IMAGE_NAME}" || exit $?
