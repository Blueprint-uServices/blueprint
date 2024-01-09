package rabbitmq

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

// Blueprint IR Node that represents the generated client for the rabbitmq container
type RabbitmqGoClient struct {
	golang.Service
	backend.Queue
	InstanceName string
	QueueName    *ir.IRValue
	Addr         *address.DialConfig
	Iface        *goparser.ParsedInterface
	Constructor  *gocode.Constructor
}

func newRabbitmqGoClient(name string, addr *address.DialConfig, queue_name *ir.IRValue) (*RabbitmqGoClient, error) {
	client := &RabbitmqGoClient{}
	err := client.init(name)
	if err != nil {
		return nil, err
	}
	client.InstanceName = name
	client.Addr = addr
	client.QueueName = queue_name
	return client, nil
}

func (n *RabbitmqGoClient) String() string {
	return n.InstanceName + " = RabbitmqClient(" + n.Addr.Name() + ")"
}

func (n *RabbitmqGoClient) Name() string {
	return n.InstanceName
}

func (n *RabbitmqGoClient) init(name string) error {
	workflow.Init("../../runtime")

	spec, err := workflow.GetSpec()
	if err != nil {
		return err
	}

	details, err := spec.Get("RabbitMQ")
	if err != nil {
		return err
	}

	n.InstanceName = name
	n.Iface = details.Iface
	n.Constructor = details.Constructor.AsConstructor()
	return nil
}

func (n *RabbitmqGoClient) GetInterface(ctx ir.BuildContext) (service.ServiceInterface, error) {
	return n.Iface.ServiceInterface(ctx), nil
}

func (n *RabbitmqGoClient) AddToWorkspace(builder golang.WorkspaceBuilder) error {
	return golang.AddRuntimeModule(builder)
}

func (n *RabbitmqGoClient) AddInterfaces(builder golang.ModuleBuilder) error {
	return n.AddToWorkspace(builder.Workspace())
}

func (n *RabbitmqGoClient) AddInstantiation(builder golang.NamespaceBuilder) error {
	if builder.Visited(n.InstanceName) {
		return nil
	}
	slog.Info(fmt.Sprintf("Instantiating RabbitmqClient %v in %v/%v", n.InstanceName, builder.Info().Package.PackageName, builder.Info().FileName))

	return builder.DeclareConstructor(n.InstanceName, n.Constructor, []ir.IRNode{n.Addr, n.QueueName})
}

func (n *RabbitmqGoClient) ImplementsGolangNode()    {}
func (n *RabbitmqGoClient) ImplementsGolangService() {}
