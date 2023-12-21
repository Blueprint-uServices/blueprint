package timeouts

import (
	"fmt"
	"reflect"

	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/blueprint"
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/coreplugins/service"
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/ir"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/golang"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/golang/gocode"
)

// Blueprint IR node representing a Timeout node
type TimeoutClient struct {
	golang.Service
	golang.GeneratesFuncs
	golang.Instantiable

	InstanceName string
	Wrapped      golang.Service

	TimeoutValue  *ir.IRValue
	outputPackage string
}

func newTimeoutClient(name string, server ir.IRNode, timeout string) (*TimeoutClient, error) {
	serverNode, is_callable := server.(golang.Service)
	if !is_callable {
		return nil, blueprint.Errorf("timeout server wrapper requires %s to be a golang service but got %s", server.Name(), reflect.TypeOf(server).String())
	}

	node := &TimeoutClient{}
	node.InstanceName = name
	node.Wrapped = serverNode
	node.outputPackage = "timeouts"
	node.TimeoutValue = &ir.IRValue{Value: timeout}
	return node, nil
}

func (node *TimeoutClient) ImplementsGolangNode() {}

func (node *TimeoutClient) Name() string {
	return node.InstanceName
}

func (node *TimeoutClient) String() string {
	return node.Name() + " = TimeoutClient(" + node.Wrapped.Name() + ")"
}

func (node *TimeoutClient) AddInterfaces(builder golang.ModuleBuilder) error {
	return node.Wrapped.AddInterfaces(builder)
}

func (node *TimeoutClient) GetInterface(ctx ir.BuildContext) (service.ServiceInterface, error) {
	return node.Wrapped.GetInterface(ctx)
}

func (node *TimeoutClient) GenerateFuncs(builder golang.ModuleBuilder) error {
	if builder.Visited(node.InstanceName + ".generateFuncs") {
		return nil
	}

	iface, err := golang.GetGoInterface(builder, node)
	if err != nil {
		return err
	}

	return generateClient(builder, iface, node.outputPackage)
}

func (node *TimeoutClient) AddInstantiation(builder golang.NamespaceBuilder) error {
	if builder.Visited(node.InstanceName) {
		return nil
	}

	iface, err := golang.GetGoInterface(builder, node.Wrapped)
	if err != nil {
		return err
	}

	constructor := &gocode.Constructor{
		Package: builder.Module().Info().Name + "/" + node.outputPackage,
		Func: gocode.Func{
			Name: fmt.Sprintf("New_%v_TimeoutClient", iface.BaseName),
			Arguments: []gocode.Variable{
				{Name: "ctx", Type: &gocode.UserType{Package: "context", Name: "Context"}},
				{Name: "client", Type: iface},
				{Name: "timeout_val", Type: &gocode.BasicType{Name: "string"}},
			},
		},
	}

	return builder.DeclareConstructor(node.InstanceName, constructor, []ir.IRNode{node.Wrapped, node.TimeoutValue})
}
