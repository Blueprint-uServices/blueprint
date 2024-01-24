package redis

import (
	"fmt"

	"github.com/blueprint-uservices/blueprint/blueprint/pkg/coreplugins/address"
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/coreplugins/backend"
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/coreplugins/service"
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/ir"
	"github.com/blueprint-uservices/blueprint/plugins/golang"
	"github.com/blueprint-uservices/blueprint/plugins/workflow/workflowspec"
	"github.com/blueprint-uservices/blueprint/runtime/plugins/redis"
	"golang.org/x/exp/slog"
)

// Blueprint IR Node that represents a client to a redis container
type RedisGoClient struct {
	golang.Service
	backend.Cache

	InstanceName string
	Addr         *address.DialConfig
	Spec         *workflowspec.Service
}

func newRedisGoClient(name string, addr *address.DialConfig) (*RedisGoClient, error) {
	spec, err := workflowspec.GetService[redis.RedisCache]()
	client := &RedisGoClient{
		InstanceName: name,
		Addr:         addr,
		Spec:         spec,
	}
	return client, err
}

// Implements ir.IRNode
func (n *RedisGoClient) String() string {
	return n.InstanceName + " = RedisClient(" + n.Addr.Name() + ")"
}

// Implements ir.IRNode
func (n *RedisGoClient) Name() string {
	return n.InstanceName
}

// Implements service.ServiceNode
func (n *RedisGoClient) GetInterface(ctx ir.BuildContext) (service.ServiceInterface, error) {
	return n.Spec.Iface.ServiceInterface(ctx), nil
}

// Implements golang.ProvidesModule
func (n *RedisGoClient) AddToWorkspace(builder golang.WorkspaceBuilder) error {
	return n.Spec.AddToWorkspace(builder)
}

// Implements golang.ProvidesInterface
func (n *RedisGoClient) AddInterfaces(builder golang.ModuleBuilder) error {
	return n.Spec.AddToModule(builder)
}

// Implements golang.Instantiable
func (n *RedisGoClient) AddInstantiation(builder golang.NamespaceBuilder) error {
	if builder.Visited(n.InstanceName) {
		return nil
	}

	slog.Info(fmt.Sprintf("Instantiating RedisClient %v in %v/%v", n.InstanceName, builder.Info().Package.PackageName, builder.Info().FileName))

	return builder.DeclareConstructor(n.InstanceName, n.Spec.Constructor.AsConstructor(), []ir.IRNode{n.Addr})
}

func (node *RedisGoClient) ImplementsGolangNode()    {}
func (node *RedisGoClient) ImplementsGolangService() {}
