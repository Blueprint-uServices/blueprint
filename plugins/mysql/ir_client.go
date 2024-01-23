package mysql

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

// Blueprint IR Node that represents the generated client for the mysql container
type MySQLDBGoClient struct {
	golang.Service
	backend.RelDB
	InstanceName string
	Username     *ir.IRValue
	Password     *ir.IRValue
	DBVal        *ir.IRValue
	Addr         *address.DialConfig
	Iface        *goparser.ParsedInterface
	Constructor  *gocode.Constructor
}

func newMySQLDBGoClient(name string, addr *address.DialConfig, username *ir.IRValue, password *ir.IRValue, dbname *ir.IRValue) (*MySQLDBGoClient, error) {
	client := &MySQLDBGoClient{}
	err := client.init(name)
	if err != nil {
		return nil, err
	}
	client.InstanceName = name
	client.Addr = addr
	client.Username = username
	client.Password = password
	client.DBVal = dbname
	return client, nil
}

// Implements ir.IRNode
func (m *MySQLDBGoClient) Name() string {
	return m.InstanceName
}

// Implements ir.IRNode
func (m *MySQLDBGoClient) String() string {
	return m.InstanceName + " = MySqlClient(" + m.Addr.Name() + ")"
}

func (m *MySQLDBGoClient) init(name string) error {
	workflow.Init("../../runtime")

	spec, err := workflow.GetSpec()
	if err != nil {
		return err
	}

	details, err := spec.Get("MySqlDB")
	if err != nil {
		return err
	}

	m.InstanceName = name
	m.Iface = details.Iface
	m.Constructor = details.Constructor.AsConstructor()

	return nil
}

// Implements service.ServiceNode
func (m *MySQLDBGoClient) GetInterface(ctx ir.BuildContext) (service.ServiceInterface, error) {
	return m.Iface.ServiceInterface(ctx), nil
}

// Implements golang.ProvidesModule
func (m *MySQLDBGoClient) AddToWorkspace(builder golang.WorkspaceBuilder) error {
	// TODO: move runtime implementation into this package and out of Blueprint runtime package
	//       afterwards, need to add interfaces from node.Iface and node.Constructor
	return fmt.Errorf("not implemented")
	// return golang.AddRuntimeModule(builder)
}

// Implements golang.ProvidesInterface
func (n *MySQLDBGoClient) AddInterfaces(builder golang.ModuleBuilder) error {
	// TODO: move runtime implementation into this package and out of Blueprint runtime package
	//       afterwards, need to add interfaces from node.Iface and node.Constructor
	return fmt.Errorf("not implemented")
	// return n.AddToWorkspace(builder.Workspace())
}

// Implements golang.Instantiable
func (m *MySQLDBGoClient) AddInstantiation(builder golang.NamespaceBuilder) error {
	if builder.Visited(m.InstanceName) {
		return nil
	}

	slog.Info(fmt.Sprintf("Instantiating MySqlClient %v in %v/%v", m.InstanceName, builder.Info().Package.PackageName, builder.Info().FileName))

	return builder.DeclareConstructor(m.InstanceName, m.Constructor, []ir.IRNode{m.Addr, m.DBVal, m.Username, m.Password})
}

func (node *MySQLDBGoClient) ImplementsGolangNode()    {}
func (node *MySQLDBGoClient) ImplementsGolangService() {}
