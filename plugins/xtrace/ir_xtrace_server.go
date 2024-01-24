package xtrace

import (
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/coreplugins/address"
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/coreplugins/service"
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/ir"
	"github.com/blueprint-uservices/blueprint/plugins/docker"
	"github.com/blueprint-uservices/blueprint/plugins/golang/goparser"
	"github.com/blueprint-uservices/blueprint/plugins/workflow/workflowspec"
	"github.com/blueprint-uservices/blueprint/runtime/plugins/xtrace"
)

// Blueprint IR Node that represents the Xtrace container
type XTraceServerContainer struct {
	docker.Container
	docker.ProvidesContainerInstance

	ServerName string
	BindAddr   *address.BindConfig
	Iface      *goparser.ParsedInterface
}

// The interface exposed by the XTrace server.
type XTraceInterface struct {
	service.ServiceInterface
	Wrapped service.ServiceInterface
}

func (xt *XTraceInterface) GetName() string {
	return "xt(" + xt.Wrapped.GetName() + ")"
}

func (xt *XTraceInterface) GetMethods() []service.Method {
	return xt.Wrapped.GetMethods()
}

func newXTraceServerContainer(name string) (*XTraceServerContainer, error) {
	spec, err := workflowspec.GetService[xtrace.XTracerImpl]()
	if err != nil {
		return nil, err
	}

	server := &XTraceServerContainer{
		ServerName: name,
		Iface:      spec.Iface,
	}
	return server, nil
}

// Implements ir.IRNode
func (node *XTraceServerContainer) Name() string {
	return node.ServerName
}

// Implements ir.IRNode
func (node *XTraceServerContainer) String() string {
	return node.Name() + " = XTraceServer(" + node.BindAddr.Name() + ")"
}

// Implements service.ServiceNode
func (node *XTraceServerContainer) GetInterface(ctx ir.BuildContext) (service.ServiceInterface, error) {
	iface := node.Iface.ServiceInterface(ctx)
	return &XTraceInterface{Wrapped: iface}, nil
}

// Implements docker.ProvidesContainerInstance
func (node *XTraceServerContainer) AddContainerInstance(target docker.ContainerWorkspace) error {
	node.BindAddr.Port = 5563
	return target.DeclarePrebuiltInstance(node.ServerName, "jonathanmace/xtrace-server:latest", node.BindAddr)
}
