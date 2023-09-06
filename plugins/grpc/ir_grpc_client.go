package grpc

import (
	"bytes"
	"fmt"
	"reflect"
	"text/template"

	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/blueprint"
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/core/service"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/golang"
)

// IRNode representing a client to a Golang server
type GolangClient struct {
	golang.Node
	golang.Service

	InstanceName string
	ServerAddr   *GolangServerAddress
}

func newGolangClient(name string, serverAddr blueprint.IRNode) (*GolangClient, error) {
	addr, is_addr := serverAddr.(*GolangServerAddress)
	if !is_addr {
		return nil, fmt.Errorf("GRPC client %s expected %s to be an address, but got %s", name, serverAddr.Name(), reflect.TypeOf(serverAddr).String())
	}

	node := &GolangClient{}
	node.InstanceName = name
	node.ServerAddr = addr

	// // TODO package and files correctly, get correct interface
	// node.ServiceDetails.Package = "TODO"
	// node.ServiceDetails.Files = []string{}
	// node.ServiceDetails.Interface.Name = name
	// constructorArg := service.Variable{}
	// constructorArg.Name = "RemoteAddr"
	// constructorArg.Type = "string"
	// node.ServiceDetails.Interface.ConstructorArgs = []service.Variable{constructorArg}

	return node, nil
}

func (n *GolangClient) String() string {
	return n.InstanceName + " = GRPCClient(" + n.ServerAddr.Name() + ")"
}

func (n *GolangClient) Name() string {
	return n.InstanceName
}

func (node *GolangClient) GetInterface() service.ServiceInterface {
	grpc, isGrpc := node.ServerAddr.GetInterface().(*GRPCInterface)
	if !isGrpc {
		return nil
	}
	return grpc.Wrapped
}

var clientBuildFuncTemplate = `func(ctr golang.Container) (any, error) {

		// TODO: generated grpc client constructor

		return nil, nil

	}`

func (node *GolangClient) AddInstantiation(builder golang.DICodeBuilder) error {
	// Only generate instantiation code for this instance once
	if builder.Visited(node.InstanceName) {
		return nil
	}

	// TODO: generate the grpc stubs

	// Instantiate the code template
	t, err := template.New(node.InstanceName).Parse(clientBuildFuncTemplate)
	if err != nil {
		return err
	}

	// Generate the code
	buf := &bytes.Buffer{}
	err = t.Execute(buf, node)
	if err != nil {
		return err
	}

	return builder.Declare(node.InstanceName, buf.String())
}

func (node *GolangClient) ImplementsGolangNode()    {}
func (node *GolangClient) ImplementsGolangService() {}
