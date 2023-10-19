package grpc

import (
	"fmt"

	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/blueprint"
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/core/address"
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/core/service"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/golang"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/golang/gocode"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/grpc/grpccodegen"
	"golang.org/x/exp/slog"
)

/*
IRNode representing a client to a Golang server.
This node does not introduce any new runtime interfaces or types that can be used by other IRNodes
GRPC code generation happens during the ModuleBuilder GenerateFuncs pass
*/
type GolangClient struct {
	golang.Service
	golang.GeneratesFuncs

	InstanceName string
	ServerAddr   *address.Address[*GolangServer]

	outputPackage string
}

func newGolangClient(name string, addr *address.Address[*GolangServer]) (*GolangClient, error) {
	node := &GolangClient{}
	node.InstanceName = name
	node.ServerAddr = addr
	node.outputPackage = "grpc"

	return node, nil
}

func (n *GolangClient) String() string {
	return n.InstanceName + " = GRPCClient(" + n.ServerAddr.Name() + ")"
}

func (n *GolangClient) Name() string {
	return n.InstanceName
}

func (node *GolangClient) GetInterface(ctx blueprint.BuildContext) (service.ServiceInterface, error) {
	iface, err := node.ServerAddr.Server.GetInterface(ctx)
	if err != nil {
		return nil, err
	}
	grpc, isGrpc := iface.(*GRPCInterface)
	if !isGrpc {
		return nil, fmt.Errorf("grpc client expected a GRPC interface from %v but found %v", node.ServerAddr.Name(), iface)
	}
	wrapped, isValid := grpc.Wrapped.(*gocode.ServiceInterface)
	if !isValid {
		return nil, fmt.Errorf("grpc client expected the server's GRPC interface to wrap a gocode interface but found %v", grpc)
	}
	return wrapped, nil
}

// Just makes sure that the interface exposed by the server is included in the built module
func (node *GolangClient) AddInterfaces(builder golang.ModuleBuilder) error {
	return node.ServerAddr.Server.Wrapped.AddInterfaces(builder)
}

// Generates proto files and the RPC client
func (node *GolangClient) GenerateFuncs(builder golang.ModuleBuilder) error {
	// Get the service that we are wrapping
	iface, err := golang.GetGoInterface(builder, node)
	if err != nil {
		return err
	}

	// Only generate grpc client instantiation code for this service once
	if builder.Visited(iface.Name + ".grpc.client") {
		return nil
	}

	// Generate the .proto files
	err = grpccodegen.GenerateGRPCProto(builder, iface, node.outputPackage)
	if err != nil {
		return err
	}

	// Generate the RPC client
	err = grpccodegen.GenerateClient(builder, iface, node.outputPackage)
	if err != nil {
		return err
	}

	return nil
}

func (node *GolangClient) AddInstantiation(builder golang.GraphBuilder) error {
	// Only generate instantiation code for this instance once
	if builder.Visited(node.InstanceName) {
		return nil
	}

	// Get the service that we are wrapping
	iface, err := golang.GetGoInterface(builder, node)
	if err != nil {
		return err
	}

	constructor := &gocode.Constructor{
		Package: builder.Module().Info().Name + "/" + node.outputPackage,
		Func: gocode.Func{
			Name: fmt.Sprintf("New_%v_GRPCClient", iface.BaseName),
			Arguments: []gocode.Variable{
				{Name: "ctx", Type: &gocode.UserType{Package: "context", Name: "Context"}},
				{Name: "addr", Type: &gocode.BasicType{Name: "string"}},
			},
		},
	}

	slog.Info(fmt.Sprintf("Instantiating GRPCClient %v in %v/%v", node.InstanceName, builder.Info().Package.PackageName, builder.Info().FileName))
	return builder.DeclareConstructor(node.InstanceName, constructor, []blueprint.IRNode{node.ServerAddr})
}

func (node *GolangClient) ImplementsGolangNode()    {}
func (node *GolangClient) ImplementsGolangService() {}
