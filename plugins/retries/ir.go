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

// Blueprint IR node representing a Retrier with Fixed Delay
type RetrierFixedDelayClient struct {
	golang.Service
	golang.GeneratesFuncs
	golang.Instantiable

	InstanceName string
	Wrapped      golang.Service

	outputPackage string
	Max           int64
	Delay         string
}

func (node *RetrierFixedDelayClient) ImplementsGolangNode() {}

func (node *RetrierFixedDelayClient) Name() string {
	return node.InstanceName
}

func (node *RetrierFixedDelayClient) String() string {
	return node.Name() + " = Retrier(" + node.Wrapped.Name() + ")"
}

func newRetrierFixedDelayClient(name string, server ir.IRNode, max_clients int64, delay string) (*RetrierFixedDelayClient, error) {
	serverNode, is_callable := server.(golang.Service)
	if !is_callable {
		return nil, blueprint.Errorf("retrier server wrapper requires %s to be a golang service but got %s", server.Name(), reflect.TypeOf(server).String())
	}

	node := &RetrierFixedDelayClient{}
	node.InstanceName = name
	node.Wrapped = serverNode
	node.outputPackage = "retries"
	node.Max = max_clients
	node.Delay = delay

	return node, nil
}

func (node *RetrierFixedDelayClient) AddInterfaces(builder golang.ModuleBuilder) error {
	return node.Wrapped.AddInterfaces(builder)
}

func (node *RetrierFixedDelayClient) GetInterface(ctx ir.BuildContext) (service.ServiceInterface, error) {
	return node.Wrapped.GetInterface(ctx)
}

func (node *RetrierFixedDelayClient) GenerateFuncs(builder golang.ModuleBuilder) error {
	if builder.Visited(node.InstanceName + ".generateFuncs") {
		return nil
	}

	iface, err := golang.GetGoInterface(builder, node)
	if err != nil {
		return err
	}

	return generateFixedDelayClient(builder, iface, node.outputPackage, node.Max, node.Delay)
}

func (node *RetrierFixedDelayClient) AddInstantiation(builder golang.NamespaceBuilder) error {
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
			Name: fmt.Sprintf("New_%v_RetrierFixedDelayClient", iface.BaseName),
			Arguments: []gocode.Variable{
				{Name: "ctx", Type: &gocode.UserType{Package: "context", Name: "Context"}},
				{Name: "client", Type: iface},
			},
		},
	}

	return builder.DeclareConstructor(node.InstanceName, constructor, []ir.IRNode{node.Wrapped})
}

// Blueprint IR node representing a Retrier with Exponential Backoff
type RetrierExponentialBackoffClient struct {
	golang.Service
	golang.GeneratesFuncs
	golang.Instantiable

	InstanceName string
	Wrapped      golang.Service

	outputPackage string
	StartDelay    string
	BackoffLimit  string

	UseJitter bool
}

func (node *RetrierExponentialBackoffClient) ImplementsGolangNode() {}

func (node *RetrierExponentialBackoffClient) Name() string {
	return node.InstanceName
}

func (node *RetrierExponentialBackoffClient) String() string {
	return node.Name() + " = Retrier(" + node.Wrapped.Name() + ")"
}

func newRetrierExponentialBackoffClient(name string, server ir.IRNode, delay string, backoff_limit string, use_jitter bool) (*RetrierExponentialBackoffClient, error) {
	serverNode, is_callable := server.(golang.Service)
	if !is_callable {
		return nil, blueprint.Errorf("retrier server wrapper requires %s to be a golang service but got %s", server.Name(), reflect.TypeOf(server).String())
	}

	node := &RetrierExponentialBackoffClient{}
	node.InstanceName = name
	node.Wrapped = serverNode
	node.outputPackage = "retries"
	node.StartDelay = delay
	node.BackoffLimit = backoff_limit
	node.UseJitter = use_jitter
	return node, nil
}

func (node *RetrierExponentialBackoffClient) AddInterfaces(builder golang.ModuleBuilder) error {
	return node.Wrapped.AddInterfaces(builder)
}

func (node *RetrierExponentialBackoffClient) GetInterface(ctx ir.BuildContext) (service.ServiceInterface, error) {
	return node.Wrapped.GetInterface(ctx)
}

func (node *RetrierExponentialBackoffClient) GenerateFuncs(builder golang.ModuleBuilder) error {
	if builder.Visited(node.InstanceName + ".generateFuncs") {
		return nil
	}

	iface, err := golang.GetGoInterface(builder, node)
	if err != nil {
		return err
	}

	return generateExpBackoffClient(builder, iface, node.outputPackage, node.StartDelay, node.BackoffLimit, node.UseJitter)
}

func (node *RetrierExponentialBackoffClient) AddInstantiation(builder golang.NamespaceBuilder) error {
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
			Name: fmt.Sprintf("New_%v_RetrierExpBackoffClient", iface.BaseName),
			Arguments: []gocode.Variable{
				{Name: "ctx", Type: &gocode.UserType{Package: "context", Name: "Context"}},
				{Name: "client", Type: iface},
			},
		},
	}

	return builder.DeclareConstructor(node.InstanceName, constructor, []ir.IRNode{node.Wrapped})
}

// Blueprint IR node representing a Retry Rate Limiter Client
type RetrierRateLimiterClient struct {
	golang.Service
	golang.GeneratesFuncs
	golang.Instantiable

	InstanceName string
	Wrapped      golang.Service

	outputPackage  string
	Max            int64
	RetryRateLimit int64 // retried times per second
}

func (node *RetrierRateLimiterClient) ImplementsGolangNode() {}

func (node *RetrierRateLimiterClient) Name() string {
	return node.InstanceName
}

func (node *RetrierRateLimiterClient) String() string {
	return node.Name() + " = Retrier(" + node.Wrapped.Name() + ")"
}

func newRetrierRateLimiterClient(name string, server ir.IRNode, max_clients int64, rateLimit int64) (*RetrierRateLimiterClient, error) {
	serverNode, is_callable := server.(golang.Service)
	if !is_callable {
		return nil, blueprint.Errorf("rate limiter client wrapper requires %s to be a golang service but got %s", server.Name(), reflect.TypeOf(server).String())
	}

	node := &RetrierRateLimiterClient{}
	node.InstanceName = name
	node.Wrapped = serverNode
	node.outputPackage = "retries"
	node.Max = max_clients
	node.RetryRateLimit = rateLimit

	return node, nil
}

func (node *RetrierRateLimiterClient) AddInterfaces(builder golang.ModuleBuilder) error {
	return node.Wrapped.AddInterfaces(builder)
}

func (node *RetrierRateLimiterClient) GetInterface(ctx ir.BuildContext) (service.ServiceInterface, error) {
	return node.Wrapped.GetInterface(ctx)
}

func (node *RetrierRateLimiterClient) GenerateFuncs(builder golang.ModuleBuilder) error {
	if builder.Visited(node.InstanceName + ".generateFuncs") {
		return nil
	}

	iface, err := golang.GetGoInterface(builder, node)
	if err != nil {
		return err
	}

	return generateRateLimiterClient(builder, iface, node.outputPackage, node.Max, node.RetryRateLimit)
}

func (node *RetrierRateLimiterClient) AddInstantiation(builder golang.NamespaceBuilder) error {
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
			Name: fmt.Sprintf("New_%v_RetrierRateLimiterClient", iface.BaseName),
			Arguments: []gocode.Variable{
				{Name: "ctx", Type: &gocode.UserType{Package: "context", Name: "Context"}},
				{Name: "client", Type: iface},
			},
		},
	}

	return builder.DeclareConstructor(node.InstanceName, constructor, []ir.IRNode{node.Wrapped})
}
