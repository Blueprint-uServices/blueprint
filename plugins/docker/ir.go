package docker

import (
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/blueprint"
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/core"
)

/*
The base IRNode interface for docker containers
*/
type Container interface {
	core.ContainerNode
	ImplementsDockerContainer()
}

/*
Code and artifact generation interfaces that IRNodes
can implement to provide docker images
*/
type (
	ProvidesContainerImage interface {
		AddContainerImage(ImageSet) error
	}

	ProvidesContainerInstance interface {
		AddContainerInstance(DockerApp) error
	}
)

/*
Builders used by the above code and artifact generation interfaces
*/
type (
	// accumualtes images
	ImageSet interface {
		blueprint.BuildContext

		// For using something external
		AddExternalImage(name string) error

		// For building a Dockerfile locally
		AddLocal(name string, path string) error
	}

	// accumulates instances
	DockerApp interface {
		// somehow need to figure out args and deps
		DeclareInstance(name string, image string) error
	}
)
