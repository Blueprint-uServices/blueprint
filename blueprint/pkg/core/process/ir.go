package process

import (
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/blueprint"
)

// This Node represents a process
type ProcessNode interface {
	blueprint.IRNode
}

// For generating output artifacts (e.g. code) into a directory
// Most process nodes will generate artifacts, but it is not strictly required
type ArtifactGenerator interface {
	GenerateArtifacts(string) error
}
