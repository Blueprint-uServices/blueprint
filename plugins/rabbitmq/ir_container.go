package rabbitmq

import (
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/coreplugins/address"
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/coreplugins/backend"
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/coreplugins/service"
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/ir"
	"github.com/blueprint-uservices/blueprint/plugins/docker"
	"github.com/blueprint-uservices/blueprint/plugins/golang/goparser"
	"github.com/blueprint-uservices/blueprint/plugins/workflow/workflowspec"
	"github.com/blueprint-uservices/blueprint/runtime/plugins/rabbitmq"
)

// Blueprint IR Node that represents the server side docker container
type RabbitmqContainer struct {
	backend.Queue
	docker.Container
	docker.ProvidesContainerInstance

	InstanceName string
	BindAddr     *address.BindConfig
	Iface        *goparser.ParsedInterface
}

// RabbitMQ interface exposed by the docker container.
type RabbitmqInterface struct {
	service.ServiceInterface
	Wrapped service.ServiceInterface
}

func (r *RabbitmqInterface) GetName() string {
	return "rabbitmq(" + r.Wrapped.GetName() + ")"
}

func (r *RabbitmqInterface) GetMethods() []service.Method {
	return r.Wrapped.GetMethods()
}

func newRabbitmqContainer(name string) (*RabbitmqContainer, error) {
	spec, err := workflowspec.GetService[rabbitmq.RabbitMQ]()
	if err != nil {
		return nil, err
	}
	cntr := &RabbitmqContainer{
		InstanceName: name,
		Iface:        spec.Iface,
	}
	return cntr, nil
}

// Implements ir.IRNode
func (n *RabbitmqContainer) String() string {
	return n.InstanceName + " = RabbitmqContainer(" + n.BindAddr.Name() + ")"
}

// Implements ir.IRNode
func (n *RabbitmqContainer) Name() string {
	return n.InstanceName
}

// Implements service.ServiceNode
func (n *RabbitmqContainer) GetInterface(ctx ir.BuildContext) (service.ServiceInterface, error) {
	iface := n.Iface.ServiceInterface(ctx)
	return &RabbitmqInterface{Wrapped: iface}, nil
}

// Implements docker.ProvidesContainerInstance
func (n *RabbitmqContainer) AddContainerInstance(target docker.ContainerWorkspace) error {
	n.BindAddr.Port = 5672
	err := target.DeclarePrebuiltInstance(n.InstanceName, "rabbitmq:3.8", n.BindAddr)
	if err != nil {
		return err
	}
	err = target.SetEnvironmentVariable(n.InstanceName, "RABBITMQ_DEFAULT_HOST", "/")
	if err != nil {
		return err
	}
	return target.SetEnvironmentVariable(n.InstanceName, "RABBITMQ_ERLANG_COOKIE", n.InstanceName+"-RABBITMQ")
}
