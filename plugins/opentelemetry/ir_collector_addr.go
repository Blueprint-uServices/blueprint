package opentelemetry

import (
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/blueprint"
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/core/address"
)

type OpenTelemetryCollectorAddr struct {
	address.Address
	AddrName  string
	Collector *OpenTelemetryCollector
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
		return blueprint.Errorf("address %v should point to an OpenTelemetry collector but got %v", addr.AddrName, node)
	}
	addr.Collector = collector
	return nil
}

func (addr *OpenTelemetryCollectorAddr) ImplementsAddressNode() {}
