package zipkin

import (
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/blueprint"
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/core/address"
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/core/service"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/docker"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/golang/goparser"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/workflow"
)

type ZipkinCollector struct {
	docker.Container

	CollectorName string
	BindAddr      *address.BindConfig
	Iface         *goparser.ParsedInterface
}

type ZipkinInterface struct {
	service.ServiceInterface
	Wrapped service.ServiceInterface
}

func (j *ZipkinInterface) GetName() string {
	return "j(" + j.Wrapped.GetName() + ")"
}

func (j *ZipkinInterface) GetMethods() []service.Method {
	return j.Wrapped.GetMethods()
}

func newZipkinCollector(name string, addr *address.BindConfig) (*ZipkinCollector, error) {
	collector := &ZipkinCollector{
		CollectorName: name,
		BindAddr:      addr,
	}
	err := collector.init(name)
	if err != nil {
		return nil, err
	}
	return collector, nil
}

func (node *ZipkinCollector) init(name string) error {
	workflow.Init("../../runtime")

	spec, err := workflow.GetSpec()
	if err != nil {
		return err
	}

	details, err := spec.Get("ZipkinTracer")
	if err != nil {
		return err
	}

	node.Iface = details.Iface
	return nil
}

func (node *ZipkinCollector) Name() string {
	return node.CollectorName
}

func (node *ZipkinCollector) String() string {
	return node.Name() + " = ZipkinCollector(" + node.BindAddr.Name() + ")"
}

func (node *ZipkinCollector) GetInterface(ctx blueprint.BuildContext) (service.ServiceInterface, error) {
	iface := node.Iface.ServiceInterface(ctx)
	return &ZipkinInterface{Wrapped: iface}, nil
}

func (node *ZipkinCollector) AddContainerArtifacts(targer docker.ContainerWorkspace) error {
	return nil
}

func (node *ZipkinCollector) AddContainerInstance(target docker.ContainerWorkspace) error {
	instanceName := blueprint.CleanName(node.CollectorName)

	node.BindAddr.Hostname = instanceName
	node.BindAddr.Port = 9411

	return target.DeclarePrebuiltInstance(instanceName, "openzipkin/zipkin", node.BindAddr)
}
