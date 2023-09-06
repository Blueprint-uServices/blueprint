package opentelemetry

import (
	"bytes"
	"fmt"
	"reflect"
	"text/template"

	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/blueprint"
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/core/service"
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

func (node *OpenTelemetryServerWrapper) GetInterface() service.ServiceInterface {
	// TODO: extend wrapped interface with tracing stuff
	return node.Wrapped.GetInterface()
}

var serverBuildFuncTemplate = `func(ctr golang.Container) (any, error) {

		// TODO: generated OT server constructor

		return nil, nil

	}`

func (node *OpenTelemetryServerWrapper) AddInstantiation(builder golang.DICodeBuilder) error {
	// Only generate instantiation code for this instance once
	if builder.Visited(node.WrapperName) {
		return nil
	}

	// TODO: generate the OT wrapper code

	// Instantiate the code template
	t, err := template.New(node.WrapperName).Parse(serverBuildFuncTemplate)
	if err != nil {
		return err
	}

	// Generate the code
	buf := &bytes.Buffer{}
	err = t.Execute(buf, node)
	if err != nil {
		return err
	}

	return builder.Declare(node.WrapperName, buf.String())
}

func (node *OpenTelemetryServerWrapper) ImplementsGolangNode()    {}
func (node *OpenTelemetryServerWrapper) ImplementsGolangService() {}
