package blueprint

import (
	"strings"
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
	ImplementsIRConfig()
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
	b.WriteString(Indent(strings.Join(children, "\n"), 2))
	b.WriteString("\n}")
	return b.String()
}

func (app *ApplicationNode) Compile(outputDir string) error {
	return defaultBuilders.BuildAll(outputDir, app.Children)
}
