package docker

import "gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/blueprint"

type (
	Container interface {
	}

	Deployment interface {
	}

	ContainerNode interface {
		blueprint.IRNode
		ImplementsDockerNode()
	}
)

type (
	ProvidesContainerImage interface {
		AddContainerImage(ImageSet) error
	}

	ProvidesContainerInstance interface {
		AddContainerInstance(DockerApp) error
	}
)

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
