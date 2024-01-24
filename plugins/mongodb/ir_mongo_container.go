package mongodb

import (
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/coreplugins/address"
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/coreplugins/backend"
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/coreplugins/service"
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/ir"
	"github.com/blueprint-uservices/blueprint/plugins/docker"
	"github.com/blueprint-uservices/blueprint/plugins/golang/goparser"
	"github.com/blueprint-uservices/blueprint/plugins/workflow/workflowspec"
	"github.com/blueprint-uservices/blueprint/runtime/plugins/mongodb"
)

// Blueprint IR Node that represents the server side docker container
type MongoDBContainer struct {
	docker.Container
	docker.ProvidesContainerInstance
	backend.NoSQLDB

	InstanceName string
	BindAddr     *address.BindConfig
	Iface        *goparser.ParsedInterface
}

// MongoDB interface exposed by the docker container.
type MongoInterface struct {
	service.ServiceInterface
	Wrapped service.ServiceInterface
}

func (m *MongoInterface) GetName() string {
	return "mongo(" + m.Wrapped.GetName() + ")"
}

func (m *MongoInterface) GetMethods() []service.Method {
	return m.Wrapped.GetMethods()
}

func newMongoDBContainer(name string) (*MongoDBContainer, error) {
	spec, err := workflowspec.GetService[mongodb.MongoDB]()
	if err != nil {
		return nil, err
	}

	proc := &MongoDBContainer{
		InstanceName: name,
		Iface:        spec.Iface,
	}
	return proc, nil
}

// Implements ir.IRNode
func (m *MongoDBContainer) String() string {
	return m.InstanceName + " = MongoDBProcess(" + m.BindAddr.Name() + ")"
}

// Implements ir.IRNode
func (m *MongoDBContainer) Name() string {
	return m.InstanceName
}

// Implements service.ServiceNode
func (m *MongoDBContainer) GetInterface(ctx ir.BuildContext) (service.ServiceInterface, error) {
	iface := m.Iface.ServiceInterface(ctx)
	return &MongoInterface{Wrapped: iface}, nil
}

// Implements docker.ProvidesContainerInstance
func (node *MongoDBContainer) AddContainerInstance(target docker.ContainerWorkspace) error {
	node.BindAddr.Port = 27017
	return target.DeclarePrebuiltInstance(node.InstanceName, "mongo", node.BindAddr)
}
