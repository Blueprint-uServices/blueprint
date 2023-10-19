package redis

import (
	"fmt"

	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/blueprint"
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/core/address"
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/core/backend"
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/core/service"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/golang"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/golang/gocode"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/golang/goparser"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/workflow"
	"golang.org/x/exp/slog"
)

type RedisGoClient struct {
	golang.Service
	backend.Cache
	InstanceName string
	Addr         *address.Address[*RedisProcess]

	Iface       *goparser.ParsedInterface
	Constructor *gocode.Constructor
}

func newRedisGoClient(name string, addr *address.Address[*RedisProcess]) (*RedisGoClient, error) {
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
	return n.InstanceName + " = RedisClient(" + n.Addr.Dial.Name() + ")"
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

func (n *RedisGoClient) GetInterface(ctx blueprint.BuildContext) (service.ServiceInterface, error) {
	return n.Iface.ServiceInterface(ctx), nil
}

func (n *RedisGoClient) AddToWorkspace(builder golang.WorkspaceBuilder) error {
	return golang.AddRuntimeModule(builder)
}

func (n *RedisGoClient) AddInterfaces(builder golang.ModuleBuilder) error {
	return n.AddToWorkspace(builder.Workspace())
}

func (n *RedisGoClient) AddInstantiation(builder golang.GraphBuilder) error {
	if builder.Visited(n.InstanceName) {
		return nil
	}

	slog.Info(fmt.Sprintf("Instantiating RedisClient %v in %v/%v", n.InstanceName, builder.Info().Package.PackageName, builder.Info().FileName))

	return builder.DeclareConstructor(n.InstanceName, n.Constructor, []blueprint.IRNode{n.Addr.Dial})
}

func (node *RedisGoClient) ImplementsGolangNode()    {}
func (node *RedisGoClient) ImplementsGolangService() {}
