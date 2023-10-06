package clientpool

import (
	"fmt"
	"strings"

	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/blueprint"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/golang"
)

type ClientPool struct {
	golang.Service

	PoolName       string
	N              int
	Client         golang.Service
	ArgNodes       []blueprint.IRNode
	ContainedNodes []blueprint.IRNode
}

func newClientPool(name string, n int) *ClientPool {
	return &ClientPool{
		PoolName: name,
		N:        n,
	}
}

func (node *ClientPool) Name() string {
	return node.PoolName
}

func (node *ClientPool) String() string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("%v = ClientPool(%v, %v) {\n", node.PoolName, node.Client.Name(), node.N))
	var children []string
	for _, child := range node.ContainedNodes {
		children = append(children, child.String())
	}
	b.WriteString(blueprint.Indent(strings.Join(children, "\n"), 2))
	b.WriteString("\n}")
	return b.String()
}

func (node *ClientPool) AddArg(argnode blueprint.IRNode) {
	node.ArgNodes = append(node.ArgNodes, argnode)
}

func (node *ClientPool) AddChild(child blueprint.IRNode) error {
	node.ContainedNodes = append(node.ContainedNodes, child)
	return nil
}
