package opentelemetry

import (
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/core/address"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/docker"
)

type OpenTelemetryCollector struct {
	docker.Container

	CollectorName string
	Addr          *address.Address[*OpenTelemetryCollector]
}

func newOpenTelemetryCollector(name string, addr *address.Address[*OpenTelemetryCollector]) (*OpenTelemetryCollector, error) {
	return &OpenTelemetryCollector{
		CollectorName: name,
		Addr:          addr,
	}, nil
}

func (node *OpenTelemetryCollector) Name() string {
	return node.CollectorName
}

func (node *OpenTelemetryCollector) String() string {
	return node.Name() + " = OTCollector(" + node.Addr.Bind.Name() + ")"
}
