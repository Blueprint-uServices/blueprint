package golang

import "gitlab.mpi-sws.org/cld/blueprint/pkg/blueprint"

// Represents an application-level golang node that can generate, package, and instantiate code
type GolangCodeNode interface {
	blueprint.IRNode

	// Golang code nodes can create instances
	GenerateInstantiationCode(*GolangCodeGenerator) error
}

type GolangArtifactNode interface {
	blueprint.IRNode

	// Golang artifact nodes can generate output artifacts like files and code
	CollectArtifacts(*GolangArtifactGenerator) error
}
