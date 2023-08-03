package golang_process

import (
	"fmt"
	"strings"

	"gitlab.mpi-sws.org/cld/blueprint/pkg/blueprint"
	"gitlab.mpi-sws.org/cld/blueprint/pkg/plugins/golang"
	"gitlab.mpi-sws.org/cld/blueprint/pkg/plugins/process"
)

// This Node represents a Golang process that internally will instantiate a number of application-level services
type GolangProcessNode struct {
	blueprint.IRNode
	process.ProcessNode
	golang.GolangArtifactNode

	InstanceName           string
	ArgNodes               []blueprint.IRNode
	ContainedArtifactNodes []golang.GolangArtifactNode
	ContainedInstanceNodes []golang.GolangCodeNode
}

// A Golang Process Node can either be given the child nodes ahead of time, or they can be added using AddArtifactNode / AddCodeNode
func newGolangProcessNode(name string) *GolangProcessNode {
	node := GolangProcessNode{}
	node.InstanceName = name
	return &node
}

func (node *GolangProcessNode) Name() string {
	return node.InstanceName
}

func (node *GolangProcessNode) String() string {
	var b strings.Builder
	b.WriteString("GolangProcessNode ")
	b.WriteString(node.InstanceName)
	b.WriteString(" = (")
	var args []string
	for _, arg := range node.ArgNodes {
		args = append(args, arg.Name())
	}
	b.WriteString(strings.Join(args, ", "))
	b.WriteString(") {\n")
	var children []string
	for _, child := range node.ContainedArtifactNodes {
		children = append(children, child.String())
	}
	b.WriteString(blueprint.Indent(strings.Join(children, "\n"), 2))
	b.WriteString("\n}")
	return b.String()
}

func (node *GolangProcessNode) AddArg(argnode blueprint.IRNode) {
	node.ArgNodes = append(node.ArgNodes, argnode)
}

func (node *GolangProcessNode) AddChild(child blueprint.IRNode) error {
	if artifactNode, ok := child.(golang.GolangArtifactNode); ok {
		node.ContainedArtifactNodes = append(node.ContainedArtifactNodes, artifactNode)
	} else if instanceNode, ok := child.(golang.GolangCodeNode); ok {
		node.ContainedInstanceNodes = append(node.ContainedInstanceNodes, instanceNode)
	} else {
		return fmt.Errorf("golang process nodes do not support IR nodes of type: %s", child)
	}
	return nil
}

func (node *GolangProcessNode) CollectArtifacts(ag *golang.GolangArtifactGenerator) error {
	// Collect all the artifacts of the contained nodes
	for _, n := range node.ContainedArtifactNodes {
		n.CollectArtifacts(ag)
	}

	// Now generate our own artifacts, using code generator
	ca := golang.NewGolangCodeAccumulator()
	for _, n := range node.ContainedInstanceNodes {
		n.GenerateInstantiationCode(ca)
	}

	code := `

	`

	// TODO: correct output path
	ag.AddCode(node.InstanceName, code)
	return nil
}
