package linuxcontainer

import "gitlab.mpi-sws.org/cld/blueprint/plugins/docker"

/*
The docker deployer for linux containers extends the default deployer,
in that it collects and packages the process artifacts for the container.
In addition, it then creates a Dockerfile for the container.

The Dockerfile will copy all process artifacts into the container.
By default, the Dockerfile will then call the build.sh from within the
container.

Each process within the container can optionally provide additional
commands to add to the Dockerfile, if implemented.
*/

type DockerLinuxContainer interface {
	docker.ContainerNode
	docker.ProvidesContainerImage
	docker.ProvidesContainerInstance
}
