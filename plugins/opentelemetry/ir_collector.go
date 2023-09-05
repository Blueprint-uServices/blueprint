package opentelemetry

import (
	"fmt"

	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/blueprint"
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/core/address"
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/core/process"
)

type OpenTelemetryCollectorAddr struct {
	address.Address
	AddrName  string
	Collector *OpenTelemetryCollector
}

type OpenTelemetryCollector struct {
	process.ProcessNode
	// TODO: artifact generation

	CollectorName string
	Addr          *OpenTelemetryCollectorAddr
}

func newOpenTelemetryCollector(name string, addr blueprint.IRNode) (*OpenTelemetryCollector, error) {
	addrNode, is_addr := addr.(*OpenTelemetryCollectorAddr)
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

func (addr *OpenTelemetryCollectorAddr) Name() string {
	return addr.AddrName
}

func (addr *OpenTelemetryCollectorAddr) String() string {
	return addr.AddrName + " = OpenTelemetryCollectorAddr()"
}

func (addr *OpenTelemetryCollectorAddr) GetDestination() blueprint.IRNode {
	if addr.Collector != nil {
		return addr.Collector
	}
	return nil
}

func (addr *OpenTelemetryCollectorAddr) SetDestination(node blueprint.IRNode) error {
	collector, isCollector := node.(*OpenTelemetryCollector)
	if !isCollector {
		return fmt.Errorf("address %v should point to an OpenTelemetry collector but got %v", addr.AddrName, node)
	}
	addr.Collector = collector
	return nil
}

func (addr *OpenTelemetryCollectorAddr) ImplementsAddressNode() {}
