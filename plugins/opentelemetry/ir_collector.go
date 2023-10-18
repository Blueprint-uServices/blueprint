package opentelemetry

import "gitlab.mpi-sws.org/cld/blueprint/plugins/docker"

type OpenTelemetryCollector struct {
	docker.Container

	CollectorName string
	Addr          *OpenTelemetryCollectorAddr
}

func newOpenTelemetryCollector(name string, addr *OpenTelemetryCollectorAddr) (*OpenTelemetryCollector, error) {
	return &OpenTelemetryCollector{
		CollectorName: name,
		Addr:          addr,
	}, nil
}

func (node *OpenTelemetryCollector) Name() string {
	return node.CollectorName
}

func (node *OpenTelemetryCollector) String() string {
	return node.Name() + " = OTCollector(" + node.Addr.Name() + ")"
}
