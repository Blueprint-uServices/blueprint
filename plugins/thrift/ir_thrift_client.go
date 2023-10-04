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

type GolangThriftClient struct {
	golang.Node
	golang.Service
	golang.GeneratesFuncs
	golang.Instantiable

	InstanceName  string
	ServerAddr    *GolangThriftServerAddress
	outputPackage string
}

func newGolangThriftClient(name string, serverAddr blueprint.IRNode) (*GolangThriftClient, error) {
	addr, is_addr := serverAddr.(*GolangThriftServerAddress)
	if !is_addr {
		return nil, blueprint.Errorf("Thrift client %s expected %s to be an address, but got %s", name, serverAddr.Name(), reflect.TypeOf(serverAddr).String())
	}

	node := &GolangThriftClient{}
	node.InstanceName = name
	node.ServerAddr = addr
	node.outputPackage = "thrift"

	return node, nil
}

func (n *GolangThriftClient) String() string {
	return n.InstanceName + " = ThriftClient(" + n.ServerAddr.Name() + ")"
}

func (n *GolangThriftClient) Name() string {
	return n.InstanceName
}

func (node *GolangThriftClient) GetInterface() service.ServiceInterface {
	return node.GetGoInterface()
}

func (node *GolangThriftClient) GetGoInterface() *gocode.ServiceInterface {
	thrift, isThrift := node.ServerAddr.GetInterface().(*ThriftInterface)
	if !isThrift {
		return nil
	}
	wrapped, isValid := thrift.Wrapped.(*gocode.ServiceInterface)
	if !isValid {
		return nil
	}
	return wrapped
}

func (node *GolangThriftClient) GenerateFuncs(builder golang.ModuleBuilder) error {
	service := node.GetGoInterface()
	if service == nil {
		return blueprint.Errorf("expected %v to have a gocode.ServiceInterface but got %v", node.Name(), node.ServerAddr.GetInterface())
	}

	if builder.Visited(service.Name + ".grpc.client") {
		return nil
	}

	// Generate the .thrift files
	err := thriftcodegen.GenerateThrift(builder, service, node.outputPackage)
	if err != nil {
		return err
	}

	err = thriftcodegen.GenerateClient(builder, service, node.outputPackage)
	if err != nil {
		return err
	}

	return nil
}

func (node *GolangThriftClient) AddInstantiation(builder golang.GraphBuilder) error {
	if builder.Visited(node.InstanceName) {
		return nil
	}

	constructor := &gocode.Constructor{
		Package: builder.Module().Info().Name + "/" + node.outputPackage,
		Func: gocode.Func{
			Name: fmt.Sprintf("New_%v_ThriftClient", node.GetGoInterface().Name),
			Arguments: []gocode.Variable{
				{Name: "addr", Type: &gocode.BasicType{Name: "string"}},
			},
		},
	}

	slog.Info(fmt.Sprintf("Instantiating ThriftClient %v in %v/%v", node.InstanceName, builder.Info().Package.PackageName, builder.Info().FileName))
	return builder.DeclareConstructor(node.InstanceName, constructor, []blueprint.IRNode{node.ServerAddr})
}

func (node *GolangThriftClient) ImplementsGolangNode()    {}
func (node *GolangThriftClient) ImplementsGolangService() {}
