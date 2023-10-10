package circuitbreaker

import (
	"fmt"
	"reflect"

	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/blueprint"
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/core/service"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/golang"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/golang/gocode"
)

type CircuitBreakerClient struct {
	golang.Service
	golang.GeneratesFuncs
	golang.Instantiable

	InstanceName string
	Wrapped      golang.Service

	outputPackage string
	Min_Reqs      int64
	FailureRate   float64
	Interval      string
}

func (node *CircuitBreakerClient) ImplementsGolangNode() {}

func (node *CircuitBreakerClient) Name() string {
	return node.InstanceName
}

func (node *CircuitBreakerClient) String() string {
	return node.Name() + " = CircuitBreaker(" + node.Wrapped.Name() + ")"
}

func newCircuitBreakerClient(name string, server blueprint.IRNode, min_reqs int64, failure_rate float64, interval string) (*CircuitBreakerClient, error) {
	serverNode, is_callable := server.(golang.Service)
	if !is_callable {
		return nil, blueprint.Errorf("circuitbreaker client wrapper requires %s to be a golang service but got %s", server.Name(), reflect.TypeOf(server).String())
	}

	node := &CircuitBreakerClient{}
	node.InstanceName = name
	node.Wrapped = serverNode
	node.outputPackage = "cb"
	node.Min_Reqs = min_reqs
	node.FailureRate = failure_rate
	node.Interval = interval

	return node, nil
}

func (node *CircuitBreakerClient) AddInterfaces(builder golang.ModuleBuilder) error {
	return node.Wrapped.AddInterfaces(builder)
}

func (node *CircuitBreakerClient) GetInterface(ctx blueprint.BuildContext) (service.ServiceInterface, error) {
	return node.Wrapped.GetInterface(ctx)
}

func (node *CircuitBreakerClient) GenerateFuncs(builder golang.ModuleBuilder) error {
	if builder.Visited(node.InstanceName + ".generateFuncs") {
		return nil
	}

	iface, err := golang.GetGoInterface(builder, node)
	if err != nil {
		return err
	}

	return generateClient(builder, iface, node.outputPackage, node.Min_Reqs, node.FailureRate, node.Interval)
}

func (node *CircuitBreakerClient) AddInstantiation(builder golang.GraphBuilder) error {
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
			Name: fmt.Sprintf("New_%v_CircuitBreakerClient", iface.BaseName),
			Arguments: []gocode.Variable{
				{Name: "ctx", Type: &gocode.UserType{Package: "context", Name: "Context"}},
				{Name: "client", Type: iface},
			},
		},
	}

	return builder.DeclareConstructor(node.InstanceName, constructor, []blueprint.IRNode{node.Wrapped})
}
