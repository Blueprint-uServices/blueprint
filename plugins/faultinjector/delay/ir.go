package delay

import (
	"fmt"

	"github.com/blueprint-uservices/blueprint/blueprint/pkg/coreplugins/service"
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/ir"
	"github.com/blueprint-uservices/blueprint/plugins/golang"
	"github.com/blueprint-uservices/blueprint/plugins/golang/gocode"
)

// Blueprint IR node representing a RandomDelayServerWrapper
type RandomDelayServerWrapper struct {
	golang.Service
	golang.GeneratesFuncs
	golang.Instantiable

	InstanceName string
	Wrapped      golang.Service

	MaxDelay int64

	outputPackage string
}

func (node *RandomDelayServerWrapper) ImplementsGolangNode() {}

func (node *RandomDelayServerWrapper) Name() string {
	return node.InstanceName
}

func (node *RandomDelayServerWrapper) String() string {
	return node.Name() + " = RandomDelayServerWrapper(" + node.Wrapped.Name() + ")"
}

func (node *RandomDelayServerWrapper) AddInterfaces(builder golang.ModuleBuilder) error {
	return node.Wrapped.AddInterfaces(builder)
}

func (node *RandomDelayServerWrapper) GetInterface(ctx ir.BuildContext) (service.ServiceInterface, error) {
	return node.Wrapped.GetInterface(ctx)
}

func (node *RandomDelayServerWrapper) GenerateFuncs(builder golang.ModuleBuilder) error {
	service, err := golang.GetGoInterface(builder, node.Wrapped)
	if err != nil {
		return err
	}

	iface, err := golang.GetGoInterface(builder, node)
	if err != nil {
		return err
	}

	return generateServerHandler(builder, iface, service, node.outputPackage)
}

func (node *RandomDelayServerWrapper) AddInstantiation(builder golang.NamespaceBuilder) error {
	// Only generate instantiation code for this instance once
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
			Name: fmt.Sprintf("New_%v_RandomDelayHandler", iface.BaseName),
			Arguments: []gocode.Variable{
				{Name: "ctx", Type: &gocode.UserType{Package: "context", Name: "Context"}},
				{Name: "service", Type: iface},
				{Name: "max_delay", Type: &gocode.BasicType{Name: "int64"}},
			},
		},
	}

	return builder.DeclareConstructor(node.InstanceName, constructor, []ir.IRNode{node.Wrapped, &ir.IRValue{Value: fmt.Sprintf("%v", node.MaxDelay)}})
}

func NewRandomDelayServerWrapper(name string, serverNode golang.Service, maxDelay int64) (*RandomDelayServerWrapper, error) {
	node := &RandomDelayServerWrapper{}
	node.InstanceName = name
	node.Wrapped = serverNode
	node.outputPackage = "delay"
	node.MaxDelay = maxDelay

	return node, nil
}
