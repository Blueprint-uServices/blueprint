package jaeger

import (
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/coreplugins/address"
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/coreplugins/service"
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/ir"
	"github.com/blueprint-uservices/blueprint/plugins/docker"
	"github.com/blueprint-uservices/blueprint/plugins/golang/goparser"
	"github.com/blueprint-uservices/blueprint/plugins/workflow/workflowspec"
	"github.com/blueprint-uservices/blueprint/runtime/plugins/jaeger"
)

// Blueprint IR node that represents the Jaeger container
type JaegerCollectorContainer struct {
	docker.Container

	CollectorName string
	BindAddr      *address.BindConfig
	UIBindAddr    *address.BindConfig

	Iface *goparser.ParsedInterface
}

// Jaeger interface exposed to the application.
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

func newJaegerCollectorContainer(name string) (*JaegerCollectorContainer, error) {
	spec, err := workflowspec.GetService[jaeger.JaegerTracer]()
	if err != nil {
		return nil, err
	}

	collector := &JaegerCollectorContainer{
		CollectorName: name,
		Iface:         spec.Iface,
	}
	return collector, nil
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
	node.UIBindAddr.Port = 16686
	node.BindAddr.Port = 14268
	return target.DeclarePrebuiltInstance(node.CollectorName, "jaegertracing/all-in-one:latest", node.BindAddr, node.UIBindAddr)
}
