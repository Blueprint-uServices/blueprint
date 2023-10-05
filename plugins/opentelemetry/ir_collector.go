package opentelemetry

import (
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/core/process"
)

type OpenTelemetryCollector struct {
	process.ProcessNode
	process.ArtifactGenerator

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

func (node *OpenTelemetryCollector) GenerateArtifacts(outputDir string) error {
	// TODO: generate artifacts for the OT collector process
	return nil
}
