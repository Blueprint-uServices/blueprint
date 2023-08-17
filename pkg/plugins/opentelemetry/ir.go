package opentelemetry

import (
	"gitlab.mpi-sws.org/cld/blueprint/pkg/blueprint"
	"gitlab.mpi-sws.org/cld/blueprint/pkg/core/process"
	"gitlab.mpi-sws.org/cld/blueprint/pkg/plugins/golang"
	"golang.org/x/exp/slog"
)

type OpenTelemetryCollector struct {
	process.ProcessNode

	CollectorName string
	Addr          blueprint.IRNode
}

type OpenTelemetryCollectorClient struct {
	golang.Node
	golang.ArtifactGenerator
	golang.CodeGenerator

	ClientName string
	Server     blueprint.IRNode
}

type OpenTelemetryServerWrapper struct {
	golang.ArtifactGenerator
	golang.CodeGenerator

	WrapperName string
	Wrapped     golang.Service
	Collector   blueprint.IRNode
}

type OpenTelemetryClientWrapper struct {
	golang.Service
	golang.ArtifactGenerator
	golang.CodeGenerator

	WrapperName    string
	Server         blueprint.IRNode
	Collector      blueprint.IRNode
	ServiceDetails *golang.GolangServiceDetails
}

func newOpenTelemetryCollector(name string, addr blueprint.IRNode) *OpenTelemetryCollector {
	node := OpenTelemetryCollector{}
	node.CollectorName = name
	node.Addr = addr
	slog.Info("Build OpenTelemetry Collector node " + name)
	return &node
}

func newOpenTelemetryCollectorClient(name string, server blueprint.IRNode) *OpenTelemetryCollectorClient {
	node := OpenTelemetryCollectorClient{}
	node.ClientName = name
	node.Server = server
	slog.Info("Build OpenTelemetry Collector client " + name)
	return &node
}

func newOpenTelemetryServerWrapper(name string, wrapped golang.Service, addr blueprint.IRNode, collector blueprint.IRNode) *OpenTelemetryServerWrapper {
	node := OpenTelemetryServerWrapper{}
	node.WrapperName = name
	node.Wrapped = wrapped
	slog.Info("Build OpenTelemetry Server wrapper " + name)
	return &node
}

func newOpenTelemetryClientWrapper(name string, server blueprint.IRNode, collector blueprint.IRNode) *OpenTelemetryClientWrapper {
	node := OpenTelemetryClientWrapper{}
	node.WrapperName = name
	node.Server = server
	slog.Info("Build OpenTelemetry Client wrapper " + name)
	return &node
}

func (node *OpenTelemetryCollector) Name() string {
	return node.CollectorName
}

func (node *OpenTelemetryCollectorClient) Name() string {
	return node.ClientName
}

func (node *OpenTelemetryServerWrapper) Name() string {
	return node.WrapperName
}

func (node *OpenTelemetryClientWrapper) Name() string {
	return node.WrapperName
}

func (node *OpenTelemetryCollector) String() string {
	return node.Name()
}

func (node *OpenTelemetryCollectorClient) String() string {
	return node.Name()
}

func (node *OpenTelemetryServerWrapper) String() string {
	return node.Name()
}

func (node *OpenTelemetryClientWrapper) String() string {
	return node.Name()
}
