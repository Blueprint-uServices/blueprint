package core

import "gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/blueprint"

/*
The namespace granularities currently supported by Blueprint
*/
type (
	InstanceNode             interface{
		blueprint.IRNode
	}
	ProcessNode              interface{
		blueprint.IRNode
	}
	ContainerNode            interface{
		blueprint.IRNode
	}
	DeploymentNode           interface{
		blueprint.IRNode
	}
	BlueprintApplicationNode interface{
		blueprint.IRNode
	}
)

type ArtifactGenerator interface {

	/* Generate artifacts to the provided fully-qualified directory on the local filesystem */
	GenerateArtifacts(dir string) error
}
