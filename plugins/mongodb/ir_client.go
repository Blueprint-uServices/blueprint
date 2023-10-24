package mongodb

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

type MongoDBGoClient struct {
	golang.Service
	backend.NoSQLDB
	InstanceName string
	Addr         *address.DialConfig
	Iface        *goparser.ParsedInterface
	Constructor  *gocode.Constructor
}

func newMongoDBGoClient(name string, addr *address.DialConfig) (*MongoDBGoClient, error) {
	client := &MongoDBGoClient{}
	err := client.init(name)
	if err != nil {
		return nil, err
	}
	client.InstanceName = name
	client.Addr = addr
	return client, nil
}

func (m *MongoDBGoClient) String() string {
	return m.InstanceName + " = MongoClient(" + m.Addr.Name() + ")"
}

func (m *MongoDBGoClient) Name() string {
	return m.InstanceName
}

func (node *MongoDBGoClient) init(name string) error {
	workflow.Init("../../runtime")

	spec, err := workflow.GetSpec()
	if err != nil {
		return err
	}

	details, err := spec.Get("MongoDB")
	if err != nil {
		return err
	}

	node.InstanceName = name
	node.Iface = details.Iface
	node.Constructor = details.Constructor.AsConstructor()

	return nil
}

func (n *MongoDBGoClient) GetInterface(ctx blueprint.BuildContext) (service.ServiceInterface, error) {
	return n.Iface.ServiceInterface(ctx), nil
}

func (n *MongoDBGoClient) AddToWorkspace(builder golang.WorkspaceBuilder) error {
	return golang.AddRuntimeModule(builder)
}

func (n *MongoDBGoClient) AddInterfaces(builder golang.ModuleBuilder) error {
	return n.AddToWorkspace(builder.Workspace())
}

func (n *MongoDBGoClient) AddInstantiation(builder golang.GraphBuilder) error {
	if builder.Visited(n.InstanceName) {
		return nil
	}

	slog.Info(fmt.Sprintf("Instantiating MongoClient %v in %v/%v", n.InstanceName, builder.Info().Package.PackageName, builder.Info().FileName))

	return builder.DeclareConstructor(n.InstanceName, n.Constructor, []blueprint.IRNode{n.Addr})
}

func (node *MongoDBGoClient) ImplementsGolangNode()    {}
func (node *MongoDBGoClient) ImplementsGolangService() {}
