package mongodb

import (
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/blueprint"
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/core/address"
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/core/backend"
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/core/service"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/docker"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/golang/goparser"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/workflow"
)

type MongoDBProcess struct {
	docker.Container
	backend.NoSQLDB

	InstanceName string
	BindAddr     *address.BindConfig
	Iface        *goparser.ParsedInterface
}

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

func newMongoDBProcess(name string, addr *address.BindConfig) (*MongoDBProcess, error) {
	proc := &MongoDBProcess{}
	proc.InstanceName = name
	proc.BindAddr = addr
	err := proc.init(name)
	if err != nil {
		return nil, err
	}
	return proc, nil
}

func (node *MongoDBProcess) init(name string) error {
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

func (m *MongoDBProcess) String() string {
	return m.InstanceName + " = MongoDBProcess(" + m.BindAddr.Name() + ")"
}

func (m *MongoDBProcess) Name() string {
	return m.InstanceName
}

func (m *MongoDBProcess) GetInterface(ctx blueprint.BuildContext) (service.ServiceInterface, error) {
	iface := m.Iface.ServiceInterface(ctx)
	return &MongoInterface{Wrapped: iface}, nil
}

func (m *MongoDBProcess) GenerateArtifacts(outdir string) error {
	return nil
}

func (node *MongoDBProcess) AddContainerArtifacts(targer docker.ContainerWorkspace) error {
	return nil
}

func (node *MongoDBProcess) AddContainerInstance(target docker.ContainerWorkspace) error {
	return target.DeclarePrebuiltInstance(node.InstanceName, "mongo", node.BindAddr)
}
