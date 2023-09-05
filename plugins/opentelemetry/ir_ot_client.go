package opentelemetry

import (
	"fmt"
	"reflect"

	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/blueprint"
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/core/service"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/golang"
)

type OpenTelemetryClientWrapper struct {
	golang.Service

	WrapperName string
	Server      golang.Service
	Collector   *OpenTelemetryCollectorClient
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

func (node *OpenTelemetryClientWrapper) Name() string {
	return node.WrapperName
}

func (node *OpenTelemetryClientWrapper) String() string {
	return node.Name() + " = OTClientWrapper(" + node.Server.Name() + ", " + node.Collector.Name() + ")"
}

func (node *OpenTelemetryClientWrapper) GetInterface() service.ServiceInterface {
	// TODO: unwrap server interface to remove tracing stuff
	return node.Server.GetInterface()
}

func (node *OpenTelemetryClientWrapper) AddInstantiation(builder golang.DICodeBuilder) error {
	// TODO add OT library dependency to module
	// TODO generate client wrapper code and add to output module
	// module := builder.Module()

	// TODO import the client wrapper
	// builder.Import(...)
	// TODO add code to instantiate the client wrapper
	// builder.Declare(...)
	return nil
}

func (node *OpenTelemetryClientWrapper) ImplementsGolangNode()    {}
func (node *OpenTelemetryClientWrapper) ImplementsGolangService() {}
