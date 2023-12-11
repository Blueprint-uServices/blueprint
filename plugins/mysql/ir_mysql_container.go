package mysql

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
type MySQLDBContainer struct {
	docker.Container
	backend.RelDB

	InstanceName string
	BindAddr     *address.BindConfig
	Iface        *goparser.ParsedInterface

	password string
}

// MySQL interface exposed by the docker container.
type MySQLInterface struct {
	service.ServiceInterface
	Wrapped service.ServiceInterface
}

func (m *MySQLInterface) GetName() string {
	return "mysql(" + m.Wrapped.GetName() + ")"
}

func (m *MySQLInterface) GetMethods() []service.Method {
	return m.Wrapped.GetMethods()
}

func newMySQLDBContainer(name string, addr *address.BindConfig, username string, password string) (*MySQLDBContainer, error) {
	cntr := &MySQLDBContainer{}
	cntr.InstanceName = name
	cntr.BindAddr = addr
	cntr.password = password
	err := cntr.init(name)
	if err != nil {
		return nil, err
	}
	return cntr, nil
}

func (node *MySQLDBContainer) init(name string) error {
	workflow.Init("../../runtime")

	spec, err := workflow.GetSpec()
	if err != nil {
		return err
	}

	details, err := spec.Get("MySqlDB")
	if err != nil {
		return err
	}

	node.Iface = details.Iface
	return nil
}

func (m *MySQLDBContainer) String() string {
	return m.InstanceName + " = MySqlDBContainer(" + m.BindAddr.Name() + ")"
}

func (m *MySQLDBContainer) Name() string {
	return m.InstanceName
}

func (m *MySQLDBContainer) GetInterface(ctx ir.BuildContext) (service.ServiceInterface, error) {
	iface := m.Iface.ServiceInterface(ctx)
	return &MySQLInterface{Wrapped: iface}, nil
}

func (m *MySQLDBContainer) GenerateArtifacts(outdir string) error {
	return nil
}

func (m *MySQLDBContainer) AddContainerArtifacts(target docker.ContainerWorkspace) error {
	return nil
}

func (m *MySQLDBContainer) AddContainerInstance(target docker.ContainerWorkspace) error {
	m.BindAddr.Port = 3306
	err := target.DeclarePrebuiltInstance(m.InstanceName, "mysql/mysql-server", m.BindAddr)
	if err != nil {
		return err
	}
	// Set necessary environment variables
	err = target.SetEnvironmentVariable(m.InstanceName, "MYSQL_ROOT_HOST", "%")
	if err != nil {
		return err
	}

	return target.SetEnvironmentVariable(m.InstanceName, "MYSQL_ROOT_PASSWORD", m.password)
}
