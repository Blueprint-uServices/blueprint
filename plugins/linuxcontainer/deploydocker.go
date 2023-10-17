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
	docker.Container
	docker.ProvidesContainerImage
	docker.ProvidesContainerInstance
}

func (node *Container) AddContainerImage(set docker.ImageSet) error {
	// The image only needs to be created in the output directory once
	if node.Visited(node.ImageName) {
		return nil
	}

	// Create a new subdirectory to construct the image
	builder, err := set.NewImageBuilder(node.ImageName)
	if err != nil {
		return err
	}

	// Generate artifacts to the image directory
	if err := node.GenerateArtifacts(builder.ImageDir); err != nil {
		return err
	}

	// Generate the dockerfile into the image directory
	if err := node.generateDockerfile(builder.ImageDir); err != nil {
		return err
	}
}

func (node *Container) AddContainerInstance(app docker.DockerApp) error {
	// The instance only needs to be added to the output directory once
	if node.Visited(node.InstanceName) {
		return nil
	}

	return app.DeclareInstance(node.InstanceName, node.ImageName, node.ArgNodes)
}

func (node *Container) generateDockerfile(outputDir string) error {

}

func (node *Container) ImplementsDockerContainer() {}
