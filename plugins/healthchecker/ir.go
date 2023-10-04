package healthchecker

import (
	"fmt"
	"reflect"

	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/blueprint"
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/core/service"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/golang"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/golang/gocode"
	"golang.org/x/exp/slog"
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

func (node *HealthCheckerServerWrapper) GetGoInterface() *gocode.ServiceInterface {
	iface, valid := node.Wrapped.GetInterface().(*gocode.ServiceInterface)
	if !valid {
		slog.Error(blueprint.Errorf("expected %v to have a gocode.ServiceInterface but got %v", node.Name(), node.Wrapped.GetInterface()).Error())
	}
	i := gocode.CopyServiceInterface(fmt.Sprintf("%v_HealthCheckHandler", iface.BaseName), node.outputPackage, iface)
	health_check_method := &gocode.Func{}
	health_check_method.Name = "Health"
	health_check_method.Returns = append(health_check_method.Returns, gocode.Variable{Type: &gocode.BasicType{Name: "string"}})
	i.AddMethod(*health_check_method)
	return i
}

func (node *HealthCheckerServerWrapper) GetInterface() service.ServiceInterface {
	return node.GetGoInterface()
}

func (node *HealthCheckerServerWrapper) GenerateFuncs(builder golang.ModuleBuilder) error {
	service, valid := node.Wrapped.GetInterface().(*gocode.ServiceInterface)
	if !valid {
		return blueprint.Errorf("expected %v to have a gocode.ServiceInterface but got %v", node.Name(), node.Wrapped.GetInterface())
	}
	err := generateServerHandler(builder, service, node.outputPackage)
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

	service, valid := node.Wrapped.GetInterface().(*gocode.ServiceInterface)
	if !valid {
		return blueprint.Errorf("expected %v to have a gocode.ServiceInterface but got %v", node.Name(), node.Wrapped.GetInterface())
	}

	constructor := &gocode.Constructor{
		Package: builder.Module().Info().Name + "/" + node.outputPackage,
		Func: gocode.Func{
			Name: fmt.Sprintf("New_%v_HealthCheckHandler", service.Name),
			Arguments: []gocode.Variable{
				{Name: "service", Type: service},
			},
		},
	}

	return builder.DeclareConstructor(node.InstanceName, constructor, []blueprint.IRNode{node.Wrapped})
}
