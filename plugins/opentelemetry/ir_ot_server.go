package opentelemetry

import (
	"fmt"
	"reflect"

	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/blueprint"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/golang"
)

type OpenTelemetryServerWrapper struct {
	golang.Service
	golang.Instantiable

	WrapperName string
	Wrapped     golang.Service
	Collector   *OpenTelemetryCollectorClient
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

func (node *OpenTelemetryServerWrapper) Name() string {
	return node.WrapperName
}

func (node *OpenTelemetryServerWrapper) String() string {
	return node.Name() + " = OTServerWrapper(" + node.Wrapped.Name() + ", " + node.Collector.Name() + ")"
}

func (node *OpenTelemetryServerWrapper) AddInstantiation(builder golang.DICodeBuilder) error {
	// TODO
	return nil
}

func (node *OpenTelemetryServerWrapper) ImplementsGolangNode()    {}
func (node *OpenTelemetryServerWrapper) ImplementsGolangService() {}
