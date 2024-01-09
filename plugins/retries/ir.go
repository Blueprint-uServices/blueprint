package retries

import (
	"fmt"
	"reflect"

	"github.com/blueprint-uservices/blueprint/blueprint/pkg/blueprint"
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/coreplugins/service"
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/ir"
	"github.com/blueprint-uservices/blueprint/plugins/golang"
	"github.com/blueprint-uservices/blueprint/plugins/golang/gocode"
)

// Blueprint IR node representing a Retrier
type RetrierClient struct {
	golang.Service
	golang.GeneratesFuncs
	golang.Instantiable

	InstanceName string
	Wrapped      golang.Service

	outputPackage string
	Max           int64
}

func (node *RetrierClient) ImplementsGolangNode() {}

func (node *RetrierClient) Name() string {
	return node.InstanceName
}

func (node *RetrierClient) String() string {
	return node.Name() + " = Retrier(" + node.Wrapped.Name() + ")"
}

func newRetrierClient(name string, server ir.IRNode, max_clients int64) (*RetrierClient, error) {
	serverNode, is_callable := server.(golang.Service)
	if !is_callable {
		return nil, blueprint.Errorf("retrier server wrapper requires %s to be a golang service but got %s", server.Name(), reflect.TypeOf(server).String())
	}

	node := &RetrierClient{}
	node.InstanceName = name
	node.Wrapped = serverNode
	node.outputPackage = "retries"
	node.Max = max_clients

	return node, nil
}

func (node *RetrierClient) AddInterfaces(builder golang.ModuleBuilder) error {
	return node.Wrapped.AddInterfaces(builder)
}

func (node *RetrierClient) GetInterface(ctx ir.BuildContext) (service.ServiceInterface, error) {
	return node.Wrapped.GetInterface(ctx)
}

func (node *RetrierClient) GenerateFuncs(builder golang.ModuleBuilder) error {
	if builder.Visited(node.InstanceName + ".generateFuncs") {
		return nil
	}

	iface, err := golang.GetGoInterface(builder, node)
	if err != nil {
		return err
	}

	return generateClient(builder, iface, node.outputPackage, node.Max)
}

func (node *RetrierClient) AddInstantiation(builder golang.NamespaceBuilder) error {
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
			Name: fmt.Sprintf("New_%v_RetrierClient", iface.BaseName),
			Arguments: []gocode.Variable{
				{Name: "ctx", Type: &gocode.UserType{Package: "context", Name: "Context"}},
				{Name: "client", Type: iface},
			},
		},
	}

	return builder.DeclareConstructor(node.InstanceName, constructor, []ir.IRNode{node.Wrapped})
}
