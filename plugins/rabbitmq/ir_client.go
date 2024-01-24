package rabbitmq

import (
	"fmt"

	"github.com/blueprint-uservices/blueprint/blueprint/pkg/coreplugins/address"
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/coreplugins/backend"
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/coreplugins/service"
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/ir"
	"github.com/blueprint-uservices/blueprint/plugins/golang"
	"github.com/blueprint-uservices/blueprint/plugins/workflow/workflowspec"
	"github.com/blueprint-uservices/blueprint/runtime/plugins/rabbitmq"
	"golang.org/x/exp/slog"
)

// Blueprint IR Node that represents the generated client for the rabbitmq container
type RabbitmqGoClient struct {
	golang.Service
	backend.Queue
	InstanceName string
	QueueName    *ir.IRValue
	Addr         *address.DialConfig
	Spec         *workflowspec.Service
}

func newRabbitmqGoClient(name string, addr *address.DialConfig, queue_name *ir.IRValue) (*RabbitmqGoClient, error) {
	spec, err := workflowspec.GetService[rabbitmq.RabbitMQ]()
	client := &RabbitmqGoClient{
		InstanceName: name,
		Addr:         addr,
		QueueName:    queue_name,
		Spec:         spec,
	}
	return client, err
}

// Implements ir.IRNode
func (n *RabbitmqGoClient) Name() string {
	return n.InstanceName
}

// Implements ir.IRNode
func (n *RabbitmqGoClient) String() string {
	return n.InstanceName + " = RabbitmqClient(" + n.Addr.Name() + ")"
}

// Implements service.ServiceNode
func (n *RabbitmqGoClient) GetInterface(ctx ir.BuildContext) (service.ServiceInterface, error) {
	return n.Spec.Iface.ServiceInterface(ctx), nil
}

// Implements golang.ProvidesModule
func (n *RabbitmqGoClient) AddToWorkspace(builder golang.WorkspaceBuilder) error {
	return n.Spec.AddToWorkspace(builder)
}

// Implements golang.ProvidesInterface
func (n *RabbitmqGoClient) AddInterfaces(builder golang.ModuleBuilder) error {
	return n.Spec.AddToModule(builder)
}

// Implements golang.Instantiable
func (n *RabbitmqGoClient) AddInstantiation(builder golang.NamespaceBuilder) error {
	if builder.Visited(n.InstanceName) {
		return nil
	}
	slog.Info(fmt.Sprintf("Instantiating RabbitmqClient %v in %v/%v", n.InstanceName, builder.Info().Package.PackageName, builder.Info().FileName))

	return builder.DeclareConstructor(n.InstanceName, n.Spec.Constructor.AsConstructor(), []ir.IRNode{n.Addr, n.QueueName})
}

func (n *RabbitmqGoClient) ImplementsGolangNode()    {}
func (n *RabbitmqGoClient) ImplementsGolangService() {}
