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

type OpenTelemetryClientWrapper struct {
	golang.Service
	golang.RequiresPackages

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

func (node *OpenTelemetryClientWrapper) AddToModule(builder golang.ModuleBuilder) error {
	// Only generate instantiation code for this instance once
	if builder.Visited(node.WrapperName) {
		return nil
	}

	// The client wrapper requires the server node code dependencies
	err := builder.Visit(node.Server)
	if err != nil {
		return err
	}

	// TODO: here we would generate the tracing wrapper code and add it to the module
	//       also we would add the OT library as a module dependency

	return nil
}

var clientBuildFuncTemplate = `func(ctr golang.Container) (any, error) {

		// TODO: generated OT client constructor

		return nil, nil

	}`

func (node *OpenTelemetryClientWrapper) AddInstantiation(builder golang.GraphBuilder) error {
	// Only generate instantiation code for this instance once
	if builder.Visited(node.WrapperName) {
		return nil
	}

	// TODO: generate the OT wrapper instantiation code

	// Instantiate the code template
	t, err := template.New(node.WrapperName).Parse(clientBuildFuncTemplate)
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

func (node *OpenTelemetryClientWrapper) ImplementsGolangNode()    {}
func (node *OpenTelemetryClientWrapper) ImplementsGolangService() {}
