package opentelemetry

import (
	"fmt"

	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/blueprint"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/golang"
)

type OpenTelemetryCollectorClient struct {
	golang.Node
	golang.Instantiable

	ClientName string
	ServerAddr *OpenTelemetryCollectorAddr
}

func newOpenTelemetryCollectorClient(name string, addr blueprint.IRNode) (*OpenTelemetryCollectorClient, error) {
	addrNode, is_addr := addr.(*OpenTelemetryCollectorAddr)
	if !is_addr {
		return nil, fmt.Errorf("unable to create OpenTelemetryCollectorClient node because %s is not an address", addr.Name())
	}

	node := &OpenTelemetryCollectorClient{}
	node.ClientName = name
	node.ServerAddr = addrNode
	return node, nil
}

func (node *OpenTelemetryCollectorClient) Name() string {
	return node.ClientName
}

func (node *OpenTelemetryCollectorClient) String() string {
	return node.Name() + " = OTClient(" + node.ServerAddr.Name() + ")"
}

func (node *OpenTelemetryCollectorClient) AddInstantiation(builder golang.DICodeBuilder) error {
	// TODO
	return nil
}

func (node *OpenTelemetryCollectorClient) ImplementsGolangNode() {}
