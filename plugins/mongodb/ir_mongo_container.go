package mongodb

import (
	"github.com/Blueprint-uServices/blueprint/blueprint/pkg/coreplugins/address"
	"github.com/Blueprint-uServices/blueprint/blueprint/pkg/coreplugins/backend"
	"github.com/Blueprint-uServices/blueprint/blueprint/pkg/coreplugins/service"
	"github.com/Blueprint-uServices/blueprint/blueprint/pkg/ir"
	"github.com/Blueprint-uServices/blueprint/plugins/docker"
	"github.com/Blueprint-uServices/blueprint/plugins/golang/goparser"
	"github.com/Blueprint-uServices/blueprint/plugins/workflow"
)

// Blueprint IR Node that represents the server side docker container
type MongoDBContainer struct {
	docker.Container
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
	proc := &MongoDBContainer{}
	proc.InstanceName = name
	err := proc.init(name)
	if err != nil {
		return nil, err
	}
	return proc, nil
}

func (node *MongoDBContainer) init(name string) error {
	workflow.Init("../../runtime")

	spec, err := workflow.GetSpec()
	if err != nil {
		return err
	}

	details, err := spec.Get("MongoDB")
	if err != nil {
		return err
	}
	node.Iface = details.Iface
	return nil
}

func (m *MongoDBContainer) String() string {
	return m.InstanceName + " = MongoDBProcess(" + m.BindAddr.Name() + ")"
}

func (m *MongoDBContainer) Name() string {
	return m.InstanceName
}

func (m *MongoDBContainer) GetInterface(ctx ir.BuildContext) (service.ServiceInterface, error) {
	iface := m.Iface.ServiceInterface(ctx)
	return &MongoInterface{Wrapped: iface}, nil
}

func (m *MongoDBContainer) GenerateArtifacts(outdir string) error {
	return nil
}

func (node *MongoDBContainer) AddContainerArtifacts(targer docker.ContainerWorkspace) error {
	return nil
}

func (node *MongoDBContainer) AddContainerInstance(target docker.ContainerWorkspace) error {
	node.BindAddr.Port = 27017
	return target.DeclarePrebuiltInstance(node.InstanceName, "mongo", node.BindAddr)
}
