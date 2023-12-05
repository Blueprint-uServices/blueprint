package mysql

import (
	"fmt"

	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/coreplugins/address"
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/coreplugins/backend"
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/coreplugins/service"
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/ir"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/golang"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/golang/gocode"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/golang/goparser"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/workflow"
	"golang.org/x/exp/slog"
)

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

func (m *MySQLDBGoClient) String() string {
	return m.InstanceName + " = MySqlClient(" + m.Addr.Name() + ")"
}

func (m *MySQLDBGoClient) Name() string {
	return m.InstanceName
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

func (m *MySQLDBGoClient) GetInterface(ctx ir.BuildContext) (service.ServiceInterface, error) {
	return m.Iface.ServiceInterface(ctx), nil
}

func (m *MySQLDBGoClient) AddToWorkspace(builder golang.WorkspaceBuilder) error {
	return golang.AddRuntimeModule(builder)
}

func (m *MySQLDBGoClient) AddInstantiation(builder golang.NamespaceBuilder) error {
	if builder.Visited(m.InstanceName) {
		return nil
	}

	slog.Info(fmt.Sprintf("Instantiating MongoClient %v in %v/%v", m.InstanceName, builder.Info().Package.PackageName, builder.Info().FileName))

	return builder.DeclareConstructor(m.InstanceName, m.Constructor, []ir.IRNode{m.Addr, m.DBVal, m.Username, m.Password})
}

func (node *MySQLDBGoClient) ImplementsGolangNode()    {}
func (node *MySQLDBGoClient) ImplementsGolangService() {}
