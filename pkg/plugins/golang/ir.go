package golang

import "gitlab.mpi-sws.org/cld/blueprint/pkg/blueprint"

// Base representation for any application-level golang object
type Node interface {
	blueprint.IRNode
}

// Represents an application-level golang node that can generate, package, and instantiate code
type GolangCodeNode interface {
	Node

	// Golang code nodes can create instances
	GenerateInstantiationCode(*GolangCodeGenerator) error
}

// Represents an application-level golang node that wants to include files, code, and dependencies with the generated artifact
type GolangArtifactNode interface {
	Node

	// Golang artifact nodes can generate output artifacts like files and code
	CollectArtifacts(*GolangArtifactGenerator) error
}
