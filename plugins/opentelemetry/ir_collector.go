package opentelemetry

import "gitlab.mpi-sws.org/cld/blueprint/plugins/process"

type OpenTelemetryCollector struct {
	process.Node
	process.InstantiableProcess
	process.ProvidesProcessArtifacts

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

func (n *OpenTelemetryCollector) AddProcessArtifacts(builder process.WorkspaceBuilder) error {
	// TODO: generate artifacts
	return nil
}

func (n *OpenTelemetryCollector) AddProcessInstance(builder process.GraphBuilder) error {
	// TODO: instantiate the process
	return nil
}
