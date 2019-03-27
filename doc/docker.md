Doppelganger's support for Docker is considered "experimental". 
# Docker

Doppelganger has support for synchronizing with filesystems inside Docker containers.

Doppelganger requires the `docker` command to be in the user's path.

Docker container filesystem endpoints can be specified to Doppelganger's `create`
command using URLs of the form:

    docker://[user@]container/path

The `user` componet is optional. 
The `container` component can specify any type of container identifier
understood by `docker cp` and `docker exec`

The `path` component:

	absolute path (/var/www)
    docker://container/var/www
	
	home-directory-relative path (~/project)
    docker://container/~/project

	alternate user home-directory-relative path (~otheruser/project)
    docker://container/~otheruser/project

	Windows absolute path (C:\path)
    docker://container/C:\path


The Docker client's behavior is controlled by three environment variables:

- `DOCKER_HOST`
- `DOCKER_TLS_VERIFY`
- `DOCKER_CERT_PATH`

Doppelganger is aware of these environment variables and will lock them in at session creation time.

