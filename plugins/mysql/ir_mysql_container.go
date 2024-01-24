package mysql

import (
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/coreplugins/address"
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/coreplugins/backend"
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/coreplugins/service"
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/ir"
	"github.com/blueprint-uservices/blueprint/plugins/docker"
	"github.com/blueprint-uservices/blueprint/plugins/golang/goparser"
	"github.com/blueprint-uservices/blueprint/plugins/workflow/workflowspec"
	"github.com/blueprint-uservices/blueprint/runtime/plugins/mysql"
)

// Blueprint IR Node that represents the server side docker container
type MySQLDBContainer struct {
	backend.RelDB
	docker.Container
	docker.ProvidesContainerInstance

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

func newMySQLDBContainer(name, root_password string) (*MySQLDBContainer, error) {
	spec, err := workflowspec.GetService[mysql.MySqlDB]()
	if err != nil {
		return nil, err
	}

	cntr := &MySQLDBContainer{
		InstanceName: name,
		Iface:        spec.Iface,
		password:     root_password,
	}
	return cntr, nil
}

// Implements ir.IRNode
func (m *MySQLDBContainer) String() string {
	return m.InstanceName + " = MySqlDBContainer(" + m.BindAddr.Name() + ")"
}

// Implements ir.IRNode
func (m *MySQLDBContainer) Name() string {
	return m.InstanceName
}

// Implements service.ServiceNode
func (m *MySQLDBContainer) GetInterface(ctx ir.BuildContext) (service.ServiceInterface, error) {
	iface := m.Iface.ServiceInterface(ctx)
	return &MySQLInterface{Wrapped: iface}, nil
}

// Implements docker.ProvidesContainerInstance
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
