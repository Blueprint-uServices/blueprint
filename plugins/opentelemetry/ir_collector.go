package opentelemetry

import (
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/core/address"
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/core/service"
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/ir"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/docker"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/golang/goparser"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/workflow"
)

type OpenTelemetryCollector struct {
	docker.Container

	CollectorName string
	BindAddr      *address.BindConfig
	Iface         *goparser.ParsedInterface
}

type OTCollectorInterface struct {
	service.ServiceInterface
	Wrapped service.ServiceInterface
}

func (xt *OTCollectorInterface) GetName() string {
	return "xt(" + xt.Wrapped.GetName() + ")"
}

func (xt *OTCollectorInterface) GetMethods() []service.Method {
	return xt.Wrapped.GetMethods()
}

func newOpenTelemetryCollector(name string, addr *address.BindConfig) (*OpenTelemetryCollector, error) {
	collector := &OpenTelemetryCollector{
		CollectorName: name,
		BindAddr:      addr,
	}
	err := collector.init(name)
	if err != nil {
		return nil, err
	}
	return collector, nil
}

func (node *OpenTelemetryCollector) init(name string) error {
	workflow.Init("../../runtime")

	spec, err := workflow.GetSpec()
	if err != nil {
		return err
	}

	details, err := spec.Get("StdoutTracer")
	if err != nil {
		return err
	}

	node.Iface = details.Iface
	return nil
}

func (node *OpenTelemetryCollector) Name() string {
	return node.CollectorName
}

func (node *OpenTelemetryCollector) String() string {
	return node.Name() + " = OTCollector(" + node.BindAddr.Name() + ")"
}

func (node *OpenTelemetryCollector) GetInterface(ctx ir.BuildContext) (service.ServiceInterface, error) {
	iface := node.Iface.ServiceInterface(ctx)
	return &OTCollectorInterface{Wrapped: iface}, nil
}

func (node *OpenTelemetryCollector) AddContainerArtifacts(target docker.ContainerWorkspace) error {
	// OpenTelemetryCollector doesn't have any artifacts to add
	return nil
}

func (node *OpenTelemetryCollector) AddContainerInstance(target docker.ContainerWorkspace) error {
	// OpenTelemetryCollector doesn't have any instances to add
	return nil
}
