package xtrace

import "gitlab.mpi-sws.org/cld/blueprint/plugins/golang"

type XTraceClient struct {
	golang.Node
	golang.Instantiable

	ClientName string
	ServerAddr *GolangXTraceAddress
}

func newXTraceClient(name string, addr *GolangXTraceAddress) (*XTraceClient, error) {
	node := &XTraceClient{}
	node.ClientName = name
	node.ServerAddr = addr
	return node, nil
}

func (node *XTraceClient) Name() string {
	return node.ClientName
}

func (node *XTraceClient) String() string {
	return node.Name() + " = XTraceClient(" + node.ServerAddr.Name() + ")"
}

func (node *XTraceClient) AddInstantiation(builder golang.GraphBuilder) error {
	if builder.Visited(node.ClientName) {
		return nil
	}

	// TODO: Implement the instantiation code

	return nil
}

func (node *XTraceClient) ImplementsGolangNode() {}
