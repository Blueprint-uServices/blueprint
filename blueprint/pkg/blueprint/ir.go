package blueprint

import (
	"strings"

	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/stringutil"
)

// This file contains some of the core IR nodes for Blueprint

// The base IRNode type
type IRNode interface {
	Name() string
	String() string
	// TODO: all nodes should have a Build() function
}

type IRMetadata interface {
	ImplementsIRMetadata()
}

type IRConfig interface {
	IRNode
	Optional() bool
	HasValue() bool
	Value() string
	ImplementsIRConfig()
}

type ArtifactGenerator interface {

	/* Generate artifacts to the provided fully-qualified directory on the local filesystem */
	GenerateArtifacts(dir string) error
}

// The IR Node that represents the whole application
type ApplicationNode struct {
	IRNode

	name     string
	Children []IRNode
}

func (node *ApplicationNode) Name() string {
	return node.name
}

// Print the IR graph
func (node *ApplicationNode) String() string {
	var b strings.Builder
	b.WriteString(node.name)
	b.WriteString(" = BlueprintApplication() {\n")
	var children []string
	for _, node := range node.Children {
		children = append(children, node.String())
	}
	b.WriteString(stringutil.Indent(strings.Join(children, "\n"), 2))
	b.WriteString("\n}")
	return b.String()
}

func (app *ApplicationNode) Compile(outputDir string) error {
	return defaultBuilders.BuildAll(outputDir, app.Children)
}
