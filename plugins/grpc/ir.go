package grpc

import (
	"fmt"
	"reflect"

	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/blueprint"
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/core/pointer"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/golang"
)

type GolangServer struct {
	golang.Node
	golang.ArtifactGenerator
	golang.CodeGenerator

	InstanceName string
	Addr         *pointer.Address
	Wrapped      golang.Service
}

type GolangClient struct {
	golang.Node
	golang.ArtifactGenerator
	golang.CodeGenerator
	golang.Service

	InstanceName   string
	ServerAddr     *pointer.Address
	ServiceDetails golang.GolangServiceDetails
}

func newGolangServer(name string, serverAddr blueprint.IRNode, wrapped blueprint.IRNode) (*GolangServer, error) {
	addr, is_addr := serverAddr.(*pointer.Address)
	if !is_addr {
		return nil, fmt.Errorf("GRPC server %s expected %s to be an address, but got %s", name, serverAddr.Name(), reflect.TypeOf(serverAddr).String())
	}

	service, is_service := wrapped.(golang.Service)
	if !is_service {
		return nil, fmt.Errorf("GRPC server %s expected %s to be a golang service, but got %s", name, wrapped.Name(), reflect.TypeOf(wrapped).String())
	}

	node := &GolangServer{}
	node.InstanceName = name
	node.Addr = addr
	node.Wrapped = service
	return node, nil
}

func newGolangClient(name string, serverAddr blueprint.IRNode) (*GolangClient, error) {
	addr, is_addr := serverAddr.(*pointer.Address)
	if !is_addr {
		return nil, fmt.Errorf("GRPC client %s expected %s to be an address, but got %s", name, serverAddr.Name(), reflect.TypeOf(serverAddr).String())
	}

	node := &GolangClient{}
	node.InstanceName = name
	node.ServerAddr = addr

	// // TODO package and files correctly
	// node.ServiceDetails.Package = "TODO"
	// node.ServiceDetails.Files = []string{}
	// node.ServiceDetails.Interface.Name = name
	// constructorArg := service.Variable{}
	// constructorArg.Name = "RemoteAddr"
	// constructorArg.Type = "string"
	// node.ServiceDetails.Interface.ConstructorArgs = []service.Variable{constructorArg}

	return node, nil
}

func (client *GolangClient) SetInterface(node golang.Service) {
	client.ServiceDetails.Interface.Methods = node.GetInterface().Methods
}

func (n *GolangServer) String() string {
	return n.InstanceName + " = GRPCServer(" + n.Wrapped.Name() + ", " + n.Addr.Name() + ")"
}

func (n *GolangServer) Name() string {
	return n.InstanceName
}

func (n *GolangClient) String() string {
	return n.InstanceName + " = GRPCClient(" + n.ServerAddr.Name() + ")"
}

func (n *GolangClient) Name() string {
	return n.InstanceName
}

func (node *GolangServer) ImplementsGolangNode()    {}
func (node *GolangClient) ImplementsGolangNode()    {}
func (node *GolangClient) ImplementsGolangService() {}
