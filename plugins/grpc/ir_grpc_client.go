package grpc

import (
	"bytes"
	"fmt"
	"reflect"
	"text/template"

	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/blueprint"
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/core/service"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/golang"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/golang/gocode"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/grpc/grpccodegen"
)

// IRNode representing a client to a Golang server
type GolangClient struct {
	golang.Node
	golang.Service
	golang.RequiresPackages

	InstanceName  string
	ServerAddr    *GolangServerAddress
	OutputPackage string
}

func newGolangClient(name string, serverAddr blueprint.IRNode) (*GolangClient, error) {
	addr, is_addr := serverAddr.(*GolangServerAddress)
	if !is_addr {
		return nil, fmt.Errorf("GRPC client %s expected %s to be an address, but got %s", name, serverAddr.Name(), reflect.TypeOf(serverAddr).String())
	}

	node := &GolangClient{}
	node.InstanceName = name
	node.ServerAddr = addr
	node.OutputPackage = "grpc"

	return node, nil
}

func (n *GolangClient) String() string {
	return n.InstanceName + " = GRPCClient(" + n.ServerAddr.Name() + ")"
}

func (n *GolangClient) Name() string {
	return n.InstanceName
}

func (node *GolangClient) GetInterface() service.ServiceInterface {
	return node.GetGoInterface()
}

func (node *GolangClient) GetGoInterface() *gocode.ServiceInterface {
	grpc, isGrpc := node.ServerAddr.GetInterface().(*GRPCInterface)
	if !isGrpc {
		return nil
	}
	wrapped, isValid := grpc.Wrapped.(*gocode.ServiceInterface)
	if !isValid {
		return nil
	}
	return wrapped
}

// This does the heavy lifting of generating proto files and wrapper classes
func (node *GolangClient) AddToModule(builder golang.ModuleBuilder) error {
	// Only generate instantiation code for this instance once
	if builder.Visited(node.InstanceName) {
		return nil
	}

	// Generating and compiling the .proto files is done by the server
	err := builder.Visit(node.ServerAddr.Server)
	if err != nil {
		return err
	}

	service := node.GetGoInterface()
	if service == nil {
		return fmt.Errorf("expected %v to have a gocode.ServiceInterface but got %v",
			node.Name(), node.ServerAddr.GetInterface())
	}

	// Generate the RPC client
	err = grpccodegen.GenerateClient(builder, service, node.OutputPackage)
	if err != nil {
		return err
	}

	return nil
}

type instantiateClientArgs struct {
	Client                 *GolangClient
	GeneratedPackageImport string
}

var instantiateClientTemplate = `func(ctr golang.Container) (any, error) {
		addr, err := ctr.Get("{{.Client.ServerAddr.AddrName}}")
		if err != nil {
			return nil, err
		}

		addrString, isString := addr.(string)
		if !isString {
			return nil, fmt.Errorf("Expected string value for {{.Client.ServerAddr.AddrName}} but got %v", addr)
		}

		return {{.GeneratedPackageImport}}.New_{{.Client.GetGoInterface.Name}}_GRPCClient(addrString)
	}`

func (node *GolangClient) AddInstantiation(builder golang.GraphBuilder) error {
	// Only generate instantiation code for this instance once
	if builder.Visited(node.InstanceName) {
		return nil
	}

	err := builder.Module().Visit(node)
	if err != nil {
		return err
	}

	fqPackageName := builder.Module().Info().Name + "/" + node.OutputPackage
	importedAs := builder.Import(fqPackageName)

	// TODO: generate the proper client wrapper instantiation code
	args := instantiateClientArgs{
		Client:                 node,
		GeneratedPackageImport: importedAs,
	}

	// Instantiate the code template
	t, err := template.New(node.InstanceName).Parse(instantiateClientTemplate)
	if err != nil {
		return err
	}

	// Generate the code
	buf := &bytes.Buffer{}
	err = t.Execute(buf, args)
	if err != nil {
		return err
	}

	return builder.Declare(node.InstanceName, buf.String())
}

func (node *GolangClient) ImplementsGolangNode()    {}
func (node *GolangClient) ImplementsGolangService() {}
