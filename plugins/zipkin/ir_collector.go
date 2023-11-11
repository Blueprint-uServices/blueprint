package zipkin

import (
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/coreplugins/address"
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/coreplugins/service"
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/ir"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/docker"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/golang/goparser"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/workflow"
)

type ZipkinCollectorContainer struct {
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

func newZipkinCollectorContainer(name string, addr *address.BindConfig) (*ZipkinCollectorContainer, error) {
	collector := &ZipkinCollectorContainer{
		CollectorName: name,
		BindAddr:      addr,
	}
	err := collector.init(name)
	if err != nil {
		return nil, err
	}
	return collector, nil
}

func (node *ZipkinCollectorContainer) init(name string) error {
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

func (node *ZipkinCollectorContainer) Name() string {
	return node.CollectorName
}

func (node *ZipkinCollectorContainer) String() string {
	return node.Name() + " = ZipkinCollector(" + node.BindAddr.Name() + ")"
}

func (node *ZipkinCollectorContainer) GetInterface(ctx ir.BuildContext) (service.ServiceInterface, error) {
	iface := node.Iface.ServiceInterface(ctx)
	return &ZipkinInterface{Wrapped: iface}, nil
}

func (node *ZipkinCollectorContainer) AddContainerArtifacts(targer docker.ContainerWorkspace) error {
	return nil
}

func (node *ZipkinCollectorContainer) AddContainerInstance(target docker.ContainerWorkspace) error {
	node.BindAddr.Port = 9411
	return target.DeclarePrebuiltInstance(node.CollectorName, "openzipkin/zipkin", node.BindAddr)
}
