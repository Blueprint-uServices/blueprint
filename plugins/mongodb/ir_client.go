package mongodb

import (
	"fmt"

	"github.com/blueprint-uservices/blueprint/blueprint/pkg/coreplugins/address"
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/coreplugins/backend"
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/coreplugins/service"
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/ir"
	"github.com/blueprint-uservices/blueprint/plugins/golang"
	"github.com/blueprint-uservices/blueprint/plugins/golang/gocode"
	"github.com/blueprint-uservices/blueprint/plugins/golang/goparser"
	"github.com/blueprint-uservices/blueprint/plugins/workflow"
	"golang.org/x/exp/slog"
)

// Blueprint IR Node that represents the generated client for the mongodb container
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

// Implements ir.IRNode
func (m *MongoDBGoClient) String() string {
	return m.InstanceName + " = MongoClient(" + m.Addr.Name() + ")"
}

// Implements ir.IRNode
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

// Implements service.ServiceNode
func (n *MongoDBGoClient) GetInterface(ctx ir.BuildContext) (service.ServiceInterface, error) {
	return n.Iface.ServiceInterface(ctx), nil
}

// Implements golang.ProvidesModule
func (n *MongoDBGoClient) AddToWorkspace(builder golang.WorkspaceBuilder) error {
	// TODO: move runtime implementation into this package and out of Blueprint runtime package
	//       afterwards, need to add interfaces from node.Iface and node.Constructor
	return fmt.Errorf("not implemented")
	// return golang.AddRuntimeModule(builder)
}

// Implements golang.ProvidesInterface
func (n *MongoDBGoClient) AddInterfaces(builder golang.ModuleBuilder) error {
	// TODO: move runtime implementation into this package and out of Blueprint runtime package
	//       afterwards, need to add interfaces from node.Iface and node.Constructor
	return fmt.Errorf("not implemented")
	// return n.AddToWorkspace(builder.Workspace())
}

// Implements golang.Instantiable
func (n *MongoDBGoClient) AddInstantiation(builder golang.NamespaceBuilder) error {
	if builder.Visited(n.InstanceName) {
		return nil
	}

	slog.Info(fmt.Sprintf("Instantiating MongoClient %v in %v/%v", n.InstanceName, builder.Info().Package.PackageName, builder.Info().FileName))

	return builder.DeclareConstructor(n.InstanceName, n.Constructor, []ir.IRNode{n.Addr})
}

func (node *MongoDBGoClient) ImplementsGolangNode()    {}
func (node *MongoDBGoClient) ImplementsGolangService() {}
