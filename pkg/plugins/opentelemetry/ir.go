package opentelemetry

import (
	"fmt"
	"reflect"

	"gitlab.mpi-sws.org/cld/blueprint/pkg/blueprint"
	"gitlab.mpi-sws.org/cld/blueprint/pkg/core/pointer"
	"gitlab.mpi-sws.org/cld/blueprint/pkg/core/process"
	"gitlab.mpi-sws.org/cld/blueprint/pkg/plugins/golang"
)

type OpenTelemetryCollector struct {
	process.ProcessNode

	CollectorName string
	Addr          *pointer.Address
}

type OpenTelemetryCollectorClient struct {
	golang.Node
	golang.ArtifactGenerator
	golang.CodeGenerator

	ClientName string
	ServerAddr *pointer.Address
}

type OpenTelemetryServerWrapper struct {
	golang.Service
	golang.ArtifactGenerator
	golang.CodeGenerator

	WrapperName string
	Wrapped     golang.Service
	Collector   *OpenTelemetryCollectorClient
}

type OpenTelemetryClientWrapper struct {
	golang.Service
	golang.ArtifactGenerator
	golang.CodeGenerator

	WrapperName string
	Server      golang.Service
	Collector   *OpenTelemetryCollectorClient
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

func newOpenTelemetryCollectorClient(name string, addr blueprint.IRNode) (*OpenTelemetryCollectorClient, error) {
	addrNode, is_addr := addr.(*pointer.Address)
	if !is_addr {
		return nil, fmt.Errorf("unable to create OpenTelemetryCollectorClient node because %s is not an address", addr.Name())
	}

	node := &OpenTelemetryCollectorClient{}
	node.ClientName = name
	node.ServerAddr = addrNode
	return node, nil
}

func newOpenTelemetryServerWrapper(name string, server blueprint.IRNode, collector blueprint.IRNode) (*OpenTelemetryServerWrapper, error) {
	serverNode, is_callable := server.(golang.Service)
	if !is_callable {
		return nil, fmt.Errorf("opentelemetry server wrapper requires %s to be a golang service but got %s", server.Name(), reflect.TypeOf(server).String())
	}

	collectorClient, is_collector_client := collector.(*OpenTelemetryCollectorClient)
	if !is_collector_client {
		return nil, fmt.Errorf("opentelemetry server wrapper requires %s to be an opentelemetry collector client", collector.Name())
	}

	node := &OpenTelemetryServerWrapper{}
	node.WrapperName = name
	node.Wrapped = serverNode
	node.Collector = collectorClient
	return node, nil
}

func newOpenTelemetryClientWrapper(name string, server blueprint.IRNode, collector blueprint.IRNode) (*OpenTelemetryClientWrapper, error) {
	serverNode, is_callable := server.(golang.Service)
	if !is_callable {
		return nil, fmt.Errorf("opentelemetry client wrapper requires %s to be a golang service but got %s", server.Name(), reflect.TypeOf(server).String())
	}

	collectorClient, is_collector_client := collector.(*OpenTelemetryCollectorClient)
	if !is_collector_client {
		return nil, fmt.Errorf("opentelemetry client  wrapper requires %s to be an opentelemetry collector client", collector.Name())
	}

	node := &OpenTelemetryClientWrapper{}
	node.WrapperName = name
	node.Server = serverNode
	node.Collector = collectorClient
	return node, nil
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
	return node.Name() + " = OTCollector(" + node.Addr.Name() + ")"
}

func (node *OpenTelemetryCollectorClient) String() string {
	return node.Name() + " = OTClient(" + node.ServerAddr.Name() + ")"
}

func (node *OpenTelemetryServerWrapper) String() string {
	return node.Name() + " = OTServerWrapper(" + node.Wrapped.Name() + ", " + node.Collector.Name() + ")"
}

func (node *OpenTelemetryClientWrapper) String() string {
	return node.Name() + " = OTClientWrapper(" + node.Server.Name() + ", " + node.Collector.Name() + ")"
}

func (node *OpenTelemetryServerWrapper) ImplementsGolangNode()    {}
func (node *OpenTelemetryServerWrapper) ImplementsGolangService() {}
func (node *OpenTelemetryClientWrapper) ImplementsGolangNode()    {}
func (node *OpenTelemetryClientWrapper) ImplementsGolangService() {}
