package thrift

import (
	"fmt"

	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/coreplugins/address"
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/coreplugins/service"
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/ir"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/golang"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/golang/gocode"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/thrift/thriftcodegen"
	"golang.org/x/exp/slog"
)

type GolangThriftServer struct {
	service.ServiceNode
	golang.GeneratesFuncs
	golang.Instantiable

	InstanceName string
	Addr         *address.Address[*GolangThriftServer]
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

func newGolangThriftServer(name string, addr *address.Address[*GolangThriftServer], service golang.Service) (*GolangThriftServer, error) {
	node := &GolangThriftServer{}
	node.InstanceName = name
	node.Addr = addr
	node.Wrapped = service
	node.outputPackage = "thrift"
	return node, nil
}

func (n *GolangThriftServer) String() string {
	return n.InstanceName + " = ThriftServer(" + n.Wrapped.Name() + ", " + n.Addr.Bind.Name() + ")"
}

func (n *GolangThriftServer) Name() string {
	return n.InstanceName
}

func (node *GolangThriftServer) GenerateFuncs(builder golang.ModuleBuilder) error {
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

func (node *GolangThriftServer) AddInstantiation(builder golang.NamespaceBuilder) error {
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
	return builder.DeclareConstructor(node.InstanceName, constructor, []ir.IRNode{node.Wrapped, node.Addr.Bind})
}

func (node *GolangThriftServer) GetInterface(ctx ir.BuildContext) (service.ServiceInterface, error) {
	iface, err := node.Wrapped.GetInterface(ctx)
	if err != nil {
		return nil, err
	}
	return &ThriftInterface{Wrapped: iface}, err
}

func (node *GolangThriftServer) ImplementsGolangNode() {}
