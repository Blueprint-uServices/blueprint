package blueprint

// This file contains some of the core IR nodes for Blueprint

// The base IRNode type
type IRNode interface {
	Name() string
	String() string
}

// For generating output artifacts (e.g. code)
type ArtifactGenerator interface {
	GenerateOutput(string) error
}
