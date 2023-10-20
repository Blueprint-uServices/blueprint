package opentelemetry

import (
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/core/address"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/docker"
)

type OpenTelemetryCollector struct {
	docker.Container

	CollectorName string
	BindAddr      *address.BindConfig
}

func newOpenTelemetryCollector(name string, addr *address.BindConfig) (*OpenTelemetryCollector, error) {
	return &OpenTelemetryCollector{
		CollectorName: name,
		BindAddr:      addr,
	}, nil
}

func (node *OpenTelemetryCollector) Name() string {
	return node.CollectorName
}

func (node *OpenTelemetryCollector) String() string {
	return node.Name() + " = OTCollector(" + node.BindAddr.Name() + ")"
}
