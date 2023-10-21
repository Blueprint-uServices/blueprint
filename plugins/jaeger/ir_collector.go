package jaeger

import (
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/blueprint"
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/core/address"
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/core/service"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/docker"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/golang/goparser"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/workflow"
)

type JaegerCollector struct {
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

func newJaegerCollector(name string, addr *address.BindConfig) (*JaegerCollector, error) {
	collector := &JaegerCollector{
		CollectorName: name,
		BindAddr:      addr,
	}
	err := collector.init(name)
	if err != nil {
		return nil, err
	}
	return collector, nil
}

func (node *JaegerCollector) init(name string) error {
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

func (node *JaegerCollector) Name() string {
	return node.CollectorName
}

func (node *JaegerCollector) String() string {
	return node.Name() + " = JaegerCollector(" + node.BindAddr.Name() + ")"
}

func (node *JaegerCollector) GetInterface(ctx blueprint.BuildContext) (service.ServiceInterface, error) {
	iface := node.Iface.ServiceInterface(ctx)
	return &JaegerInterface{Wrapped: iface}, nil
}

func (node *JaegerCollector) AddContainerArtifacts(targer docker.ContainerWorkspace) error {
	// TODO: IMplement
	return nil
}

func (node *JaegerCollector) AddContainerInstance(target docker.ContainerWorkspace) error {
	// TODO: Implement
	return nil
}
