package redis

import (
	"fmt"

	"github.com/Blueprint-uServices/blueprint/blueprint/pkg/coreplugins/address"
	"github.com/Blueprint-uServices/blueprint/blueprint/pkg/coreplugins/backend"
	"github.com/Blueprint-uServices/blueprint/blueprint/pkg/coreplugins/service"
	"github.com/Blueprint-uServices/blueprint/blueprint/pkg/ir"
	"github.com/Blueprint-uServices/blueprint/plugins/golang"
	"github.com/Blueprint-uServices/blueprint/plugins/golang/gocode"
	"github.com/Blueprint-uServices/blueprint/plugins/golang/goparser"
	"github.com/Blueprint-uServices/blueprint/plugins/workflow"
	"golang.org/x/exp/slog"
)

// Blueprint IR Node that represents a client to a redis container
type RedisGoClient struct {
	golang.Service
	backend.Cache
	InstanceName string
	Addr         *address.DialConfig

	Iface       *goparser.ParsedInterface
	Constructor *gocode.Constructor
}

func newRedisGoClient(name string, addr *address.DialConfig) (*RedisGoClient, error) {
	client := &RedisGoClient{}
	err := client.init(name)
	if err != nil {
		return nil, err
	}
	client.InstanceName = name
	client.Addr = addr
	return client, nil
}

func (n *RedisGoClient) String() string {
	return n.InstanceName + " = RedisClient(" + n.Addr.Name() + ")"
}

func (n *RedisGoClient) Name() string {
	return n.InstanceName
}

func (node *RedisGoClient) init(name string) error {
	workflow.Init("../../runtime")

	spec, err := workflow.GetSpec()
	if err != nil {
		return err
	}

	details, err := spec.Get("RedisCache")
	if err != nil {
		return err
	}

	node.InstanceName = name
	node.Iface = details.Iface
	node.Constructor = details.Constructor.AsConstructor()

	return nil
}

func (n *RedisGoClient) GetInterface(ctx ir.BuildContext) (service.ServiceInterface, error) {
	return n.Iface.ServiceInterface(ctx), nil
}

func (n *RedisGoClient) AddToWorkspace(builder golang.WorkspaceBuilder) error {
	return golang.AddRuntimeModule(builder)
}

func (n *RedisGoClient) AddInterfaces(builder golang.ModuleBuilder) error {
	return n.AddToWorkspace(builder.Workspace())
}

func (n *RedisGoClient) AddInstantiation(builder golang.NamespaceBuilder) error {
	if builder.Visited(n.InstanceName) {
		return nil
	}

	slog.Info(fmt.Sprintf("Instantiating RedisClient %v in %v/%v", n.InstanceName, builder.Info().Package.PackageName, builder.Info().FileName))

	return builder.DeclareConstructor(n.InstanceName, n.Constructor, []ir.IRNode{n.Addr})
}

func (node *RedisGoClient) ImplementsGolangNode()    {}
func (node *RedisGoClient) ImplementsGolangService() {}
