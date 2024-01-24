package mongodb

import (
	"fmt"

	"github.com/blueprint-uservices/blueprint/blueprint/pkg/coreplugins/address"
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/coreplugins/backend"
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/coreplugins/service"
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/ir"
	"github.com/blueprint-uservices/blueprint/plugins/golang"
	"github.com/blueprint-uservices/blueprint/plugins/workflow/workflowspec"
	"github.com/blueprint-uservices/blueprint/runtime/plugins/mongodb"
	"golang.org/x/exp/slog"
)

// Blueprint IR Node that represents the generated client for the mongodb container
type MongoDBGoClient struct {
	golang.Service
	backend.NoSQLDB
	InstanceName string
	Addr         *address.DialConfig

	Spec *workflowspec.Service
}

func newMongoDBGoClient(name string, addr *address.DialConfig) (*MongoDBGoClient, error) {
	spec, err := workflowspec.GetService[mongodb.MongoDB]()
	client := &MongoDBGoClient{
		InstanceName: name,
		Addr:         addr,
		Spec:         spec,
	}
	return client, err
}

// Implements ir.IRNode
func (m *MongoDBGoClient) String() string {
	return m.InstanceName + " = MongoClient(" + m.Addr.Name() + ")"
}

// Implements ir.IRNode
func (m *MongoDBGoClient) Name() string {
	return m.InstanceName
}

// Implements service.ServiceNode
func (n *MongoDBGoClient) GetInterface(ctx ir.BuildContext) (service.ServiceInterface, error) {
	return n.Spec.Iface.ServiceInterface(ctx), nil
}

// Implements golang.ProvidesModule
func (n *MongoDBGoClient) AddToWorkspace(builder golang.WorkspaceBuilder) error {
	return n.Spec.AddToWorkspace(builder)
}

// Implements golang.ProvidesInterface
func (n *MongoDBGoClient) AddInterfaces(builder golang.ModuleBuilder) error {
	return n.Spec.AddToModule(builder)
}

// Implements golang.Instantiable
func (n *MongoDBGoClient) AddInstantiation(builder golang.NamespaceBuilder) error {
	if builder.Visited(n.InstanceName) {
		return nil
	}

	slog.Info(fmt.Sprintf("Instantiating MongoClient %v in %v/%v", n.InstanceName, builder.Info().Package.PackageName, builder.Info().FileName))

	return builder.DeclareConstructor(n.InstanceName, n.Spec.Constructor.AsConstructor(), []ir.IRNode{n.Addr})
}

func (node *MongoDBGoClient) ImplementsGolangNode()    {}
func (node *MongoDBGoClient) ImplementsGolangService() {}
