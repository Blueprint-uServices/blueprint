package opentelemetry

import (
	"bytes"
	"text/template"

	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/core/address"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/golang"
	"golang.org/x/exp/slog"
)

type OpenTelemetryCollectorClient struct {
	golang.Node
	golang.Instantiable

	ClientName string
	ServerAddr *address.Address[*OpenTelemetryCollector]
}

func newOpenTelemetryCollectorClient(name string, addr *address.Address[*OpenTelemetryCollector]) (*OpenTelemetryCollectorClient, error) {
	node := &OpenTelemetryCollectorClient{}
	node.ClientName = name
	node.ServerAddr = addr
	return node, nil
}

func (node *OpenTelemetryCollectorClient) Name() string {
	return node.ClientName
}

func (node *OpenTelemetryCollectorClient) String() string {
	return node.Name() + " = OTClient(" + node.ServerAddr.Dial.Name() + ")"
}

var collectorClientBuildFuncTemplate = `func(ctr golang.Container) (any, error) {

		// TODO: generated OT collector client constructor

		return nil, nil

	}`

func (node *OpenTelemetryCollectorClient) AddInstantiation(builder golang.GraphBuilder) error {
	// Only generate instantiation code for this instance once
	if builder.Visited(node.ClientName) {
		return nil
	}

	// TODO: generate the OT wrapper code

	// Instantiate the code template
	t, err := template.New(node.ClientName).Parse(collectorClientBuildFuncTemplate)
	if err != nil {
		return err
	}

	// Generate the code
	buf := &bytes.Buffer{}
	err = t.Execute(buf, node)
	if err != nil {
		return err
	}

	slog.Info("instantiating ot client")
	return builder.Declare(node.ClientName, buf.String())
}

func (node *OpenTelemetryCollectorClient) ImplementsGolangNode() {}
