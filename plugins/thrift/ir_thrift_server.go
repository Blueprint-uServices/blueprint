package thrift

import (
	"fmt"
	"reflect"

	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/blueprint"
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/core/service"
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
	Addr         *GolangThriftServerAddress
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

func newGolangThriftServer(name string, serverAddr blueprint.IRNode, wrapped blueprint.IRNode) (*GolangThriftServer, error) {
	addr, is_addr := serverAddr.(*GolangThriftServerAddress)
	if !is_addr {
		return nil, blueprint.Errorf("Thrift server %s expected %s to be an address, but got %s", name, serverAddr.Name(), reflect.TypeOf(serverAddr).String())
	}

	service, is_service := wrapped.(golang.Service)
	if !is_service {
		return nil, blueprint.Errorf("Thrift server %s expected %s to be a golang service, but got %s", name, wrapped.Name(), reflect.TypeOf(wrapped).String())
	}

	node := &GolangThriftServer{}
	node.InstanceName = name
	node.Addr = addr
	node.Wrapped = service
	node.outputPackage = "thrift"
	return node, nil
}

func (n *GolangThriftServer) String() string {
	return n.InstanceName + " = ThriftServer(" + n.Wrapped.Name() + ", " + n.Addr.Name() + ")"
}

func (n *GolangThriftServer) Name() string {
	return n.InstanceName
}

func (node *GolangThriftServer) GenerateFuncs(builder golang.ModuleBuilder) error {
	service := node.Wrapped.GetGoInterface()

	if builder.Visited(service.Name + ".thrift.server") {
		return nil
	}

	err := thriftcodegen.GenerateThrift(builder, service, node.outputPackage)
	if err != nil {
		return err
	}

	err = thriftcodegen.GenerateServerHandler(builder, service, node.outputPackage)
	if err != nil {
		return err
	}
	return nil
}

func (node *GolangThriftServer) AddInstantiation(builder golang.GraphBuilder) error {
	if builder.Visited(node.InstanceName) {
		return nil
	}

	service := node.Wrapped.GetGoInterface()

	constructor := &gocode.Constructor{
		Package: builder.Module().Info().Name + "/" + node.outputPackage,
		Func: gocode.Func{
			Name: fmt.Sprintf("New_%v_ThriftServerHandler", service.Name),
			Arguments: []gocode.Variable{
				{Name: "service", Type: service},
				{Name: "serverAddr", Type: &gocode.BasicType{Name: "string"}},
			},
		},
	}

	slog.Info(fmt.Sprintf("Instantiating ThriftServer %v in %v/%v", node.InstanceName, builder.Info().Package.PackageName, builder.Info().FileName))
	return builder.DeclareConstructor(node.InstanceName, constructor, []blueprint.IRNode{node.Wrapped, node.Addr})
}

func (node *GolangThriftServer) GetInterface() service.ServiceInterface {
	return &ThriftInterface{Wrapped: node.Wrapped.GetInterface()}
}

func (node *GolangThriftServer) ImplementsGolangNode() {}
