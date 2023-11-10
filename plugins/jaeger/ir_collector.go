package jaeger

import (
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/core/address"
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/core/service"
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/ir"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/docker"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/golang/goparser"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/workflow"
)

type JaegerCollectorContainer struct {
	docker.Container

	CollectorName string
	BindAddr      *address.BindConfig
	Iface         *goparser.ParsedInterface
}

type JaegerInterface struct {
	service.ServiceInterface
	Wrapped service.ServiceInterface
}

func (j *JaegerInterface) GetName() string {
	return "j(" + j.Wrapped.GetName() + ")"
}

func (j *JaegerInterface) GetMethods() []service.Method {
	return j.Wrapped.GetMethods()
}

func newJaegerCollectorContainer(name string, addr *address.BindConfig) (*JaegerCollectorContainer, error) {
	collector := &JaegerCollectorContainer{
		CollectorName: name,
		BindAddr:      addr,
	}
	err := collector.init(name)
	if err != nil {
		return nil, err
	}
	return collector, nil
}

func (node *JaegerCollectorContainer) init(name string) error {
	workflow.Init("../../runtime")

	spec, err := workflow.GetSpec()
	if err != nil {
		return err
	}

	details, err := spec.Get("JaegerTracer")
	if err != nil {
		return err
	}

	node.Iface = details.Iface
	return nil
}

func (node *JaegerCollectorContainer) Name() string {
	return node.CollectorName
}

func (node *JaegerCollectorContainer) String() string {
	return node.Name() + " = JaegerCollector(" + node.BindAddr.Name() + ")"
}

func (node *JaegerCollectorContainer) GetInterface(ctx ir.BuildContext) (service.ServiceInterface, error) {
	iface := node.Iface.ServiceInterface(ctx)
	return &JaegerInterface{Wrapped: iface}, nil
}

func (node *JaegerCollectorContainer) AddContainerArtifacts(targer docker.ContainerWorkspace) error {
	return nil
}

func (node *JaegerCollectorContainer) AddContainerInstance(target docker.ContainerWorkspace) error {
	node.BindAddr.Port = 14268
	return target.DeclarePrebuiltInstance(node.CollectorName, "jaegertracing/all-in-one:latest", node.BindAddr)
}
