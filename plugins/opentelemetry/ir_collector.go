package opentelemetry

import (
	"fmt"

	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/blueprint"
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/core/pointer"
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/core/process"
)

type OpenTelemetryCollector struct {
	process.ProcessNode
	// TODO: artifact generation

	CollectorName string
	Addr          *pointer.Address
}

func newOpenTelemetryCollector(name string, addr blueprint.IRNode) (*OpenTelemetryCollector, error) {
	addrNode, is_addr := addr.(*pointer.Address)
	if !is_addr {
		return nil, fmt.Errorf("unable to create OpenTelemetryCollector node because %s is not an address", addr.Name())
	}

	node := &OpenTelemetryCollector{}
	node.CollectorName = name
	node.Addr = addrNode
	return node, nil
}

func (node *OpenTelemetryCollector) Name() string {
	return node.CollectorName
}

func (node *OpenTelemetryCollector) String() string {
	return node.Name() + " = OTCollector(" + node.Addr.Name() + ")"
}
