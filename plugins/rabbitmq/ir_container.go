package rabbitmq

import (
	"github.com/Blueprint-uServices/blueprint/blueprint/pkg/coreplugins/address"
	"github.com/Blueprint-uServices/blueprint/blueprint/pkg/coreplugins/backend"
	"github.com/Blueprint-uServices/blueprint/blueprint/pkg/coreplugins/service"
	"github.com/Blueprint-uServices/blueprint/blueprint/pkg/ir"
	"github.com/Blueprint-uServices/blueprint/plugins/docker"
	"github.com/Blueprint-uServices/blueprint/plugins/golang/goparser"
	"github.com/Blueprint-uServices/blueprint/plugins/workflow"
)

// Blueprint IR Node that represents the server side docker container
type RabbitmqContainer struct {
	docker.Container
	backend.Queue

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
	cntr := &RabbitmqContainer{}
	cntr.InstanceName = name
	err := cntr.init(name)
	if err != nil {
		return nil, err
	}
	return cntr, nil
}

func (node *RabbitmqContainer) init(name string) error {
	workflow.Init("../../runtime")

	spec, err := workflow.GetSpec()
	if err != nil {
		return err
	}

	details, err := spec.Get("RabbitMQ")
	if err != nil {
		return err
	}
	node.Iface = details.Iface

	return nil
}

func (n *RabbitmqContainer) String() string {
	return n.InstanceName + " = RabbitmqContainer(" + n.BindAddr.Name() + ")"
}

func (n *RabbitmqContainer) Name() string {
	return n.InstanceName
}

func (n *RabbitmqContainer) GetInterface(ctx ir.BuildContext) (service.ServiceInterface, error) {
	iface := n.Iface.ServiceInterface(ctx)
	return &RabbitmqInterface{Wrapped: iface}, nil
}

func (n *RabbitmqContainer) GenerateArtifacts(outdir string) error {
	return nil
}

func (n *RabbitmqContainer) AddContainerArtifacts(target docker.ContainerWorkspace) error {
	return nil
}

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
