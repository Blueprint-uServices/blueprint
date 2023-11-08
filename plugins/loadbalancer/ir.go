package loadbalancer

import (
	"fmt"
	"strings"

	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/blueprint"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/golang"
)

type LoadBalancerClient struct {
	golang.Service
	golang.GeneratesFuncs

	BalancerName   string
	Clients        []golang.Service
	ContainedNodes []blueprint.IRNode
}

func newLoadBalancerClient(name string, arg_nodes []blueprint.IRNode) *LoadBalancerClient {
	return &LoadBalancerClient{
		BalancerName:   name,
		ContainedNodes: arg_nodes,
	}
}

func (node *LoadBalancerClient) Name() string {
	return node.BalancerName
}

func (node *LoadBalancerClient) String() string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("%v = LoadBalancer() {\n", node.BalancerName))
	var children []string
	for _, child := range node.ContainedNodes {
		children = append(children, child.String())
	}
	b.WriteString(blueprint.Indent(strings.Join(children, "\n"), 2))
	b.WriteString("\n}")
	return b.String()
}

func (lb *LoadBalancerClient) AddInterfaces(module golang.ModuleBuilder) error {
	for _, node := range lb.ContainedNodes {
		if n, valid := node.(golang.ProvidesInterface); valid {
			if err := n.AddInterfaces(module); err != nil {
				return err
			}
		}
	}
	return nil
}
