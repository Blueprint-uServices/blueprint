package golang_process

import (
	"fmt"

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
	ContainedArtifactNodes []golang.GolangArtifactNode
	ContainedInstanceNodes []golang.GolangCodeNode
}

func newGolangProcessNode(name string, nodes []blueprint.IRNode) (*GolangProcessNode, error) {
	node := GolangProcessNode{}

	node.InstanceName = name

	for _, n := range nodes {
		if artifactNode, ok := n.(golang.GolangArtifactNode); ok {
			node.ContainedArtifactNodes = append(node.ContainedArtifactNodes, artifactNode)
		} else if instanceNode, ok := n.(golang.GolangCodeNode); ok {
			node.ContainedInstanceNodes = append(node.ContainedInstanceNodes, instanceNode)
		} else {
			return nil, fmt.Errorf("cannot construct a golang process node with unsupported IR node type %s", n)
		}
	}

	return &node, nil
}

func (node *GolangProcessNode) Name() string {
	return node.InstanceName
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
