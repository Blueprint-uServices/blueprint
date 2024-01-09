package healthchecker

import (
	"fmt"
	"reflect"

	"github.com/blueprint-uservices/blueprint/blueprint/pkg/blueprint"
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/coreplugins/service"
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/ir"
	"github.com/blueprint-uservices/blueprint/plugins/golang"
	"github.com/blueprint-uservices/blueprint/plugins/golang/gocode"
)

// Blueprint IR node representing a HealthChecker
type HealthCheckerServerWrapper struct {
	golang.Service
	golang.GeneratesFuncs
	golang.Instantiable

	InstanceName string
	Wrapped      golang.Service

	outputPackage string
}

func (node *HealthCheckerServerWrapper) ImplementsGolangNode() {}

func (node *HealthCheckerServerWrapper) Name() string {
	return node.InstanceName
}

func (node *HealthCheckerServerWrapper) String() string {
	return node.Name() + " = HealthCheckerServerWrapper(" + node.Wrapped.Name() + ")"
}

func (node *HealthCheckerServerWrapper) AddInterfaces(builder golang.ModuleBuilder) error {
	iface, err := node.genInterface(builder)
	if err != nil {
		return err
	}
	err = generateClientSideInterfaces(builder, iface, node.outputPackage)
	if err != nil {
		return err
	}
	return node.Wrapped.AddInterfaces(builder)
}

func newHealthCheckerServerWrapper(name string, server ir.IRNode) (*HealthCheckerServerWrapper, error) {
	serverNode, is_callable := server.(golang.Service)
	if !is_callable {
		return nil, fmt.Errorf("healthchecker server wrapper requires %s to be a golang service but got %s", server.Name(), reflect.TypeOf(server).String())
	}

	node := &HealthCheckerServerWrapper{}
	node.InstanceName = name
	node.Wrapped = serverNode
	node.outputPackage = "healthcheck"

	return node, nil
}

func (node *HealthCheckerServerWrapper) genInterface(ctx ir.BuildContext) (*gocode.ServiceInterface, error) {
	iface, err := golang.GetGoInterface(ctx, node.Wrapped)
	if err != nil {
		return nil, err
	}
	module_ctx, valid := ctx.(golang.ModuleBuilder)
	if !valid {
		return nil, blueprint.Errorf("Healthchecker expected build context to be a ModuleBuilder, got %v", ctx)
	}
	i := gocode.CopyServiceInterface(fmt.Sprintf("%v_HealthChecker", iface.BaseName), module_ctx.Info().Name+"/"+node.outputPackage, iface)
	health_check_method := &gocode.Func{}
	health_check_method.Name = "Health"
	health_check_method.Returns = append(health_check_method.Returns, gocode.Variable{Type: &gocode.BasicType{Name: "string"}})
	i.AddMethod(*health_check_method)
	return i, nil
}

func (node *HealthCheckerServerWrapper) GetInterface(ctx ir.BuildContext) (service.ServiceInterface, error) {
	return node.genInterface(ctx)
}

func (node *HealthCheckerServerWrapper) GenerateFuncs(builder golang.ModuleBuilder) error {
	service, err := golang.GetGoInterface(builder, node.Wrapped)
	if err != nil {
		return err
	}
	iface, err := golang.GetGoInterface(builder, node)
	if err != nil {
		return err
	}
	err = generateServerHandler(builder, iface, service, node.outputPackage)
	if err != nil {
		return err
	}
	return nil
}

func (node *HealthCheckerServerWrapper) AddInstantiation(builder golang.NamespaceBuilder) error {
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
			Name: fmt.Sprintf("New_%v_HealthCheckHandler", iface.BaseName),
			Arguments: []gocode.Variable{
				{Name: "ctx", Type: &gocode.UserType{Package: "context", Name: "Context"}},
				{Name: "service", Type: iface},
			},
		},
	}

	return builder.DeclareConstructor(node.InstanceName, constructor, []ir.IRNode{node.Wrapped})
}
