docker stop %DOPPELGANGER_TEST_DOCKER_CONTAINER_NAME%
docker container prune --force
docker image rm --force %DOPPELGANGER_TEST_DOCKER_IMAGE_NAME%
