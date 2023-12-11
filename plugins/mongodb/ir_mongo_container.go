package mongodb

import (
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/coreplugins/address"
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/coreplugins/backend"
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/coreplugins/service"
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/ir"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/docker"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/golang/goparser"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/workflow"
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

func newMongoDBContainer(name string, addr *address.BindConfig) (*MongoDBContainer, error) {
	proc := &MongoDBContainer{}
	proc.InstanceName = name
	proc.BindAddr = addr
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
