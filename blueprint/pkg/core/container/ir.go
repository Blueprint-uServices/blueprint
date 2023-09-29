package container

import "gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/blueprint"

// This Node represents a container
type ContainerNode interface {
	blueprint.IRNode
}

// For generating output artifacts (e.g. container files) into a directory
// Most container nodes will generate artifacts
type ArtifactGenerator interface {
	GenerateArtifacts(string) error
}
