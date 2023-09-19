package opentelemetry

import (
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/blueprint"
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/core/process"
)

type OpenTelemetryCollector struct {
	process.ProcessNode
	process.ArtifactGenerator

	CollectorName string
	Addr          *OpenTelemetryCollectorAddr
}

func newOpenTelemetryCollector(name string, addr blueprint.IRNode) (*OpenTelemetryCollector, error) {
	addrNode, is_addr := addr.(*OpenTelemetryCollectorAddr)
	if !is_addr {
		return nil, blueprint.Errorf("unable to create OpenTelemetryCollector node because %s is not an address", addr.Name())
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

func (node *OpenTelemetryCollector) GenerateArtifacts(outputDir string) error {
	// TODO: generate artifacts for the OT collector process
	return nil
}
