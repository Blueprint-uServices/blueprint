package linuxcontainer

import "gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/blueprint"

type Container struct {
	blueprint.IRNode
	// container.ContainerNode
	// container.ArtifactGenerator

	InstanceName   string
	ArgNodes       []blueprint.IRNode
	ContainedNodes []blueprint.IRNode
}

// Generates a shell script to invoke all the contained nodes
