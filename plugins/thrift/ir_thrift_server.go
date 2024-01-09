package thrift

import (
	"fmt"

	"github.com/blueprint-uservices/blueprint/blueprint/pkg/coreplugins/address"
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/coreplugins/service"
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/ir"
	"github.com/blueprint-uservices/blueprint/plugins/golang"
	"github.com/blueprint-uservices/blueprint/plugins/golang/gocode"
	"github.com/blueprint-uservices/blueprint/plugins/thrift/thriftcodegen"
	"golang.org/x/exp/slog"
)

// IRNode representing a Golang thrift server.
// This node does not introduce any new runtime interfaces or types that can be used by other IRNodes.
// Thrift code generation happens during the ModuleBuilder GeneratesFuncs pass
type golangThriftServer struct {
	service.ServiceNode
	golang.GeneratesFuncs
	golang.Instantiable

	InstanceName string
	Bind         *address.BindConfig
	Wrapped      golang.Service

	outputPackage string
}

type ThriftInterface struct {
	service.ServiceInterface
	Wrapped service.ServiceInterface
}

func (thrift *ThriftInterface) GetName() string {
	return "thrift(" + thrift.Wrapped.GetName() + ")"
}

func (thrift *ThriftInterface) GetMethods() []service.Method {
	return thrift.Wrapped.GetMethods()
}

func newGolangThriftServer(name string, service golang.Service) (*golangThriftServer, error) {
	node := &golangThriftServer{}
	node.InstanceName = name
	node.Wrapped = service
	node.outputPackage = "thrift"
	return node, nil
}

func (n *golangThriftServer) String() string {
	return n.InstanceName + " = ThriftServer(" + n.Wrapped.Name() + ", " + n.Bind.Name() + ")"
}

func (n *golangThriftServer) Name() string {
	return n.InstanceName
}

// Generates thrift files and the RPC server handler
func (node *golangThriftServer) GenerateFuncs(builder golang.ModuleBuilder) error {
	iface, err := golang.GetGoInterface(builder, node.Wrapped)
	if err != nil {
		return err
	}

	if builder.Visited(iface.Name + ".thrift.server") {
		return nil
	}

	err = thriftcodegen.GenerateThrift(builder, iface, node.outputPackage)
	if err != nil {
		return err
	}

	err = thriftcodegen.GenerateServerHandler(builder, iface, node.outputPackage)
	if err != nil {
		return err
	}
	return nil
}

func (node *golangThriftServer) AddInstantiation(builder golang.NamespaceBuilder) error {
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
			Name: fmt.Sprintf("New_%v_ThriftServerHandler", iface.BaseName),
			Arguments: []gocode.Variable{
				{Name: "ctx", Type: &gocode.UserType{Package: "context", Name: "Context"}},
				{Name: "service", Type: iface},
				{Name: "serverAddr", Type: &gocode.BasicType{Name: "string"}},
			},
		},
	}

	slog.Info(fmt.Sprintf("Instantiating ThriftServer %v in %v/%v", node.InstanceName, builder.Info().Package.PackageName, builder.Info().FileName))
	return builder.DeclareConstructor(node.InstanceName, constructor, []ir.IRNode{node.Wrapped, node.Bind})
}

func (node *golangThriftServer) GetInterface(ctx ir.BuildContext) (service.ServiceInterface, error) {
	iface, err := node.Wrapped.GetInterface(ctx)
	if err != nil {
		return nil, err
	}
	return &ThriftInterface{Wrapped: iface}, err
}

func (node *golangThriftServer) ImplementsGolangNode() {}
