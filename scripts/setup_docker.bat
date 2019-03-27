docker version
docker build --tag %DOPPELGANGER_TEST_DOCKER_IMAGE_NAME% --file scripts/dockerfile_windows scripts
docker run --name %DOPPELGANGER_TEST_DOCKER_CONTAINER_NAME% --detach %DOPPELGANGER_TEST_DOCKER_IMAGE_NAME%
