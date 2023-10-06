package healthchecker

import (
	"fmt"
	"reflect"

	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/blueprint"
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/core/service"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/golang"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/golang/gocode"
)

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

func newHealthCheckerServerWrapper(name string, server blueprint.IRNode) (*HealthCheckerServerWrapper, error) {
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

func (node *HealthCheckerServerWrapper) genInterface(ctx blueprint.BuildContext) (*gocode.ServiceInterface, error) {
	iface, err := golang.GetGoInterface(ctx, node.Wrapped)
	if err != nil {
		return nil, err
	}
	i := gocode.CopyServiceInterface(fmt.Sprintf("%v_HealthChecker", iface.BaseName), node.outputPackage, iface)
	health_check_method := &gocode.Func{}
	health_check_method.Name = "Health"
	health_check_method.Returns = append(health_check_method.Returns, gocode.Variable{Type: &gocode.BasicType{Name: "string"}})
	i.AddMethod(*health_check_method)
	return i, nil
}

func (node *HealthCheckerServerWrapper) GetInterface(ctx blueprint.BuildContext) (service.ServiceInterface, error) {
	return node.genInterface(ctx)
}

func (node *HealthCheckerServerWrapper) GenerateFuncs(builder golang.ModuleBuilder) error {
	service, err := golang.GetGoInterface(builder, node)
	if err != nil {
		return err
	}
	err = generateServerHandler(builder, service, node.outputPackage)
	if err != nil {
		return err
	}
	return nil
}

func (node *HealthCheckerServerWrapper) AddInstantiation(builder golang.GraphBuilder) error {
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
				{Name: "service", Type: iface},
			},
		},
	}

	return builder.DeclareConstructor(node.InstanceName, constructor, []blueprint.IRNode{node.Wrapped})
}
