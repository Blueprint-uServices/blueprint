package mysql

import (
	"fmt"

	"github.com/blueprint-uservices/blueprint/blueprint/pkg/coreplugins/address"
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/coreplugins/backend"
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/coreplugins/service"
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/ir"
	"github.com/blueprint-uservices/blueprint/plugins/golang"
	"github.com/blueprint-uservices/blueprint/plugins/workflow/workflowspec"
	"github.com/blueprint-uservices/blueprint/runtime/plugins/mysql"
	"golang.org/x/exp/slog"
)

// Blueprint IR Node that represents the generated client for the mysql container
type MySQLDBGoClient struct {
	golang.Service
	backend.RelDB

	InstanceName string
	Username     *ir.IRValue
	Password     *ir.IRValue
	DBVal        *ir.IRValue
	Addr         *address.DialConfig

	Spec *workflowspec.Service
}

func newMySQLDBGoClient(name string, addr *address.DialConfig, username *ir.IRValue, password *ir.IRValue, dbname *ir.IRValue) (*MySQLDBGoClient, error) {
	spec, err := workflowspec.GetService[mysql.MySqlDB]()
	client := &MySQLDBGoClient{
		InstanceName: name,
		Username:     username,
		Password:     password,
		DBVal:        dbname,
		Addr:         addr,
		Spec:         spec,
	}
	return client, err
}

// Implements ir.IRNode
func (m *MySQLDBGoClient) Name() string {
	return m.InstanceName
}

// Implements ir.IRNode
func (m *MySQLDBGoClient) String() string {
	return m.InstanceName + " = MySqlClient(" + m.Addr.Name() + ")"
}

// Implements service.ServiceNode
func (m *MySQLDBGoClient) GetInterface(ctx ir.BuildContext) (service.ServiceInterface, error) {
	return m.Spec.Iface.ServiceInterface(ctx), nil
}

// Implements golang.ProvidesModule
func (m *MySQLDBGoClient) AddToWorkspace(builder golang.WorkspaceBuilder) error {
	return m.Spec.AddToWorkspace(builder)
}

// Implements golang.ProvidesInterface
func (n *MySQLDBGoClient) AddInterfaces(builder golang.ModuleBuilder) error {
	return n.Spec.AddToModule(builder)
}

// Implements golang.Instantiable
func (m *MySQLDBGoClient) AddInstantiation(builder golang.NamespaceBuilder) error {
	if builder.Visited(m.InstanceName) {
		return nil
	}

	slog.Info(fmt.Sprintf("Instantiating MySqlClient %v in %v/%v", m.InstanceName, builder.Info().Package.PackageName, builder.Info().FileName))

	return builder.DeclareConstructor(m.InstanceName, m.Spec.Constructor.AsConstructor(), []ir.IRNode{m.Addr, m.DBVal, m.Username, m.Password})
}

func (node *MySQLDBGoClient) ImplementsGolangNode()    {}
func (node *MySQLDBGoClient) ImplementsGolangService() {}
