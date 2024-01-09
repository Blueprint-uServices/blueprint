package latency

import (
	"fmt"
	"reflect"

	"github.com/blueprint-uservices/blueprint/blueprint/pkg/blueprint"
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/coreplugins/service"
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/ir"
	"github.com/blueprint-uservices/blueprint/plugins/golang"
	"github.com/blueprint-uservices/blueprint/plugins/golang/gocode"
)

// Blueprint IR Node representing a server side latency injector
type LatencyInjectorWrapper struct {
	golang.Service
	golang.GeneratesFuncs
	golang.Instantiable

	InstanceName  string
	Wrapped       golang.Service
	outputPackage string
	LatencyValue  *ir.IRValue
}

func newLatencyInjectorWrapper(name string, server ir.IRNode, latency string) (*LatencyInjectorWrapper, error) {
	serverNode, is_callable := server.(golang.Service)
	if !is_callable {
		return nil, blueprint.Errorf("latency injector wrapper requires %s to be a golang service but got %s", server.Name(), reflect.TypeOf(server).String())
	}

	node := &LatencyInjectorWrapper{}
	node.InstanceName = name
	node.Wrapped = serverNode
	node.outputPackage = "latencyinjector"
	node.LatencyValue = &ir.IRValue{Value: latency}
	return node, nil
}

// Implements [ir.IRNode]
func (node *LatencyInjectorWrapper) ImplementsGolangNode() {}

// Implements [golang.Service]
func (node *LatencyInjectorWrapper) ImplementsGolangService() {}

// Implements [ir.IRNode]
func (node *LatencyInjectorWrapper) Name() string {
	return node.InstanceName
}

// Implements [ir.IRNode]
func (node *LatencyInjectorWrapper) String() string {
	return node.Name() + " = LatencyInjector(" + node.Wrapped.Name() + ")"
}

// Implements [golang.Service]
func (node *LatencyInjectorWrapper) AddInterfaces(builder golang.ModuleBuilder) error {
	return node.Wrapped.AddInterfaces(builder)
}

// Implements [golang.Service]
func (node *LatencyInjectorWrapper) GetInterface(ctx ir.BuildContext) (service.ServiceInterface, error) {
	return node.Wrapped.GetInterface(ctx)
}

// Implements [golang.GeneratesFuncs]
func (node *LatencyInjectorWrapper) GenerateFuncs(builder golang.ModuleBuilder) error {
	if builder.Visited(node.InstanceName + ".generateFuncs") {
		return nil
	}

	iface, err := golang.GetGoInterface(builder, node)
	if err != nil {
		return err
	}

	return generateServerWrapper(builder, iface, node.outputPackage)
}

// Implements golang.Instantiable
func (node *LatencyInjectorWrapper) AddInstantiation(builder golang.NamespaceBuilder) error {
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
			Name: fmt.Sprintf("New_%v_LatencyInjector", iface.BaseName),
			Arguments: []gocode.Variable{
				{Name: "ctx", Type: &gocode.UserType{Package: "context", Name: "Context"}},
				{Name: "server", Type: iface},
				{Name: "latency", Type: &gocode.BasicType{Name: "string"}},
			},
		},
	}

	return builder.DeclareConstructor(node.InstanceName, constructor, []ir.IRNode{node.Wrapped, node.LatencyValue})
}
