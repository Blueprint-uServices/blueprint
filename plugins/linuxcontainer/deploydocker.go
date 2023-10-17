package linuxcontainer

import (
	"gitlab.mpi-sws.org/cld/blueprint/plugins/docker"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/linuxcontainer/workspace"
)

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
	docker.Container
	docker.ProvidesContainerImage
	docker.ProvidesContainerInstance
}

func (node *Container) AddContainerArtifacts(target docker.ContainerWorkspace) error {
	// The image only needs to be created in the output directory once
	if target.Visited(node.ImageName) {
		return nil
	}

	// Create a new subdirectory to construct the image
	dir, err := target.CreateImageDir(node.ImageName)
	if err != nil {
		return err
	}

	// Generate artifacts to the image directory in the normal way.
	// By providing a docker workspace, processes will be able to
	// add dockerfile commands
	// The docker workspace extends the Finish() implementation
	// to also generate the Dockerfile
	workspace := workspace.NewDockerWorkspace(node.Name(), dir)
	if err := node.generateArtifacts(workspace); err != nil {
		return err
	}
	return nil
}

func (node *Container) AddContainerInstance(target docker.ContainerWorkspace) error {
	// The instance only needs to be added to the output directory once
	if target.Visited(node.InstanceName) {
		return nil
	}

	// TODO: all address and port related shenanigans will need to go here

	return target.DeclareLocalImage(node.InstanceName, node.ImageName)
}

func (node *Container) ImplementsDockerContainer() {}
