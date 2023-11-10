package thrift

import (
	"fmt"

	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/blueprint"
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/coreplugins/address"
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/coreplugins/service"
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/ir"
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
	ServerAddr    *address.Address[*GolangThriftServer]
	outputPackage string
}

func newGolangThriftClient(name string, addr *address.Address[*GolangThriftServer]) (*GolangThriftClient, error) {
	node := &GolangThriftClient{}
	node.InstanceName = name
	node.ServerAddr = addr
	node.outputPackage = "thrift"

	return node, nil
}

func (n *GolangThriftClient) String() string {
	return n.InstanceName + " = ThriftClient(" + n.ServerAddr.Dial.Name() + ")"
}

func (n *GolangThriftClient) Name() string {
	return n.InstanceName
}

func (node *GolangThriftClient) GetInterface(ctx ir.BuildContext) (service.ServiceInterface, error) {
	iface, err := node.ServerAddr.Server.GetInterface(ctx)
	if err != nil {
		return nil, err
	}
	tiface, isthrift := iface.(*ThriftInterface)
	if !isthrift {
		return nil, blueprint.Errorf("thrift client expected a Thrift interface from %v but found %v", node.ServerAddr.Server.Name(), iface)
	}
	wrapped, isValid := tiface.Wrapped.(*gocode.ServiceInterface)
	if !isValid {
		return nil, blueprint.Errorf("thrift client expected the server's Thrift interface to wrap a gocode interface but found %v", tiface)
	}
	return wrapped, nil
}

func (node *GolangThriftClient) AddInterfaces(builder golang.ModuleBuilder) error {
	return node.ServerAddr.Server.Wrapped.AddInterfaces(builder)
}

func (node *GolangThriftClient) GenerateFuncs(builder golang.ModuleBuilder) error {
	iface, err := golang.GetGoInterface(builder, node)
	if err != nil {
		return nil
	}

	if builder.Visited(iface.Name + ".grpc.client") {
		return nil
	}

	// Generate the .thrift files
	err = thriftcodegen.GenerateThrift(builder, iface, node.outputPackage)
	if err != nil {
		return err
	}

	err = thriftcodegen.GenerateClient(builder, iface, node.outputPackage)
	if err != nil {
		return err
	}

	return nil
}

func (node *GolangThriftClient) AddInstantiation(builder golang.GraphBuilder) error {
	if builder.Visited(node.InstanceName) {
		return nil
	}

	iface, err := golang.GetGoInterface(builder, node)
	if err != nil {
		return err
	}

	constructor := &gocode.Constructor{
		Package: builder.Module().Info().Name + "/" + node.outputPackage,
		Func: gocode.Func{
			Name: fmt.Sprintf("New_%v_ThriftClient", iface.BaseName),
			Arguments: []gocode.Variable{
				{Name: "ctx", Type: &gocode.UserType{Package: "context", Name: "Context"}},
				{Name: "addr", Type: &gocode.BasicType{Name: "string"}},
			},
		},
	}

	slog.Info(fmt.Sprintf("Instantiating ThriftClient %v in %v/%v", node.InstanceName, builder.Info().Package.PackageName, builder.Info().FileName))
	return builder.DeclareConstructor(node.InstanceName, constructor, []ir.IRNode{node.ServerAddr.Dial})
}

func (node *GolangThriftClient) ImplementsGolangNode()    {}
func (node *GolangThriftClient) ImplementsGolangService() {}
