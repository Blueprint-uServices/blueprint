package zipkin

import (
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/coreplugins/address"
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/coreplugins/service"
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/ir"
	"github.com/blueprint-uservices/blueprint/plugins/docker"
	"github.com/blueprint-uservices/blueprint/plugins/golang/goparser"
	"github.com/blueprint-uservices/blueprint/plugins/workflow/workflowspec"
	"github.com/blueprint-uservices/blueprint/runtime/plugins/zipkin"
)

// Blueprint IR node that represents the Zipkin container
type ZipkinCollectorContainer struct {
	docker.Container
	docker.ProvidesContainerInstance

	CollectorName string
	BindAddr      *address.BindConfig
	Iface         *goparser.ParsedInterface
}

// Zipkin interface exposed to the application.
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

func newZipkinCollectorContainer(name string) (*ZipkinCollectorContainer, error) {
	spec, err := workflowspec.GetService[zipkin.ZipkinTracer]()
	if err != nil {
		return nil, err
	}

	collector := &ZipkinCollectorContainer{
		CollectorName: name,
		Iface:         spec.Iface,
	}
	return collector, nil
}

// Implements ir.IRNode
func (node *ZipkinCollectorContainer) Name() string {
	return node.CollectorName
}

// Implements ir.IRNode
func (node *ZipkinCollectorContainer) String() string {
	return node.Name() + " = ZipkinCollector(" + node.BindAddr.Name() + ")"
}

// Implements service.ServiceNode
func (node *ZipkinCollectorContainer) GetInterface(ctx ir.BuildContext) (service.ServiceInterface, error) {
	iface := node.Iface.ServiceInterface(ctx)
	return &ZipkinInterface{Wrapped: iface}, nil
}

// Implements docker.ProvidesContainerInstance
func (node *ZipkinCollectorContainer) AddContainerInstance(target docker.ContainerWorkspace) error {
	node.BindAddr.Port = 9411
	return target.DeclarePrebuiltInstance(node.CollectorName, "openzipkin/zipkin", node.BindAddr)
}
