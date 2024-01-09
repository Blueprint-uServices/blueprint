package grpc

import (
	"fmt"

	"github.com/blueprint-uservices/blueprint/blueprint/pkg/coreplugins/address"
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/coreplugins/service"
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/ir"
	"github.com/blueprint-uservices/blueprint/plugins/golang"
	"github.com/blueprint-uservices/blueprint/plugins/golang/gocode"
	"github.com/blueprint-uservices/blueprint/plugins/grpc/grpccodegen"
	"golang.org/x/exp/slog"
)

/*
IRNode representing a Golang GPRC server.
This node does not introduce any new runtime interfaces or types that can be used by other IRNodes
GRPC code generation happens during the ModuleBuilder GenerateFuncs pass
*/
type golangServer struct {
	service.ServiceNode
	golang.GeneratesFuncs
	golang.Instantiable

	InstanceName string
	Bind         *address.BindConfig
	Wrapped      golang.Service

	outputPackage string
}

// Represents a service that is exposed over GRPC
type gRPCInterface struct {
	service.ServiceInterface
	Wrapped service.ServiceInterface
}

func (grpc *gRPCInterface) GetName() string {
	return "grpc(" + grpc.Wrapped.GetName() + ")"
}

func (grpc *gRPCInterface) GetMethods() []service.Method {
	return grpc.Wrapped.GetMethods()
}

func newGolangServer(name string, service golang.Service) (*golangServer, error) {
	node := &golangServer{}
	node.InstanceName = name
	node.Wrapped = service
	node.outputPackage = "grpc"
	return node, nil
}

func (n *golangServer) String() string {
	return n.InstanceName + " = GRPCServer(" + n.Wrapped.Name() + ", " + n.Bind.Name() + ")"
}

func (n *golangServer) Name() string {
	return n.InstanceName
}

// Generates proto files and the RPC server handler
func (node *golangServer) GenerateFuncs(builder golang.ModuleBuilder) error {
	// Get the service that we are wrapping
	iface, err := golang.GetGoInterface(builder, node.Wrapped)
	if err != nil {
		return err
	}

	// Only generate grpc server instantiation code for this service once
	if builder.Visited(iface.Name + ".grpc.server") {
		return nil
	}

	// Generate the .proto files
	err = grpccodegen.GenerateGRPCProto(builder, iface, node.outputPackage)
	if err != nil {
		return err
	}

	// Generate the RPC server handler
	err = grpccodegen.GenerateServerHandler(builder, iface, node.outputPackage)
	if err != nil {
		return err
	}

	return nil
}

func (node *golangServer) AddInstantiation(builder golang.NamespaceBuilder) error {
	// Only generate instantiation code for this instance once
	if builder.Visited(node.InstanceName) {
		return nil
	}

	iface, err := golang.GetGoInterface(builder.Module(), node.Wrapped)
	if err != nil {
		return err
	}

	constructor := &gocode.Constructor{
		Package: builder.Module().Info().Name + "/" + node.outputPackage,
		Func: gocode.Func{
			Name: fmt.Sprintf("New_%v_GRPCServerHandler", iface.BaseName),
			Arguments: []gocode.Variable{
				{Name: "ctx", Type: &gocode.UserType{Package: "context", Name: "Context"}},
				{Name: "service", Type: iface},
				{Name: "serverAddr", Type: &gocode.BasicType{Name: "string"}},
			},
		},
	}

	slog.Info(fmt.Sprintf("Instantiating GRPCServer %v in %v/%v", node.InstanceName, builder.Info().Package.PackageName, builder.Info().FileName))
	return builder.DeclareConstructor(node.InstanceName, constructor, []ir.IRNode{node.Wrapped, node.Bind})
}

func (node *golangServer) GetInterface(ctx ir.BuildContext) (service.ServiceInterface, error) {
	iface, err := node.Wrapped.GetInterface(ctx)
	return &gRPCInterface{Wrapped: iface}, err
}
func (node *golangServer) ImplementsGolangNode() {}
