package xtrace

import (
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/coreplugins/address"
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/coreplugins/service"
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/ir"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/docker"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/golang/goparser"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/workflow"
)

// Blueprint IR Node that represents the Xtrace container
type XTraceServerContainer struct {
	docker.Container

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
	server := &XTraceServerContainer{
		ServerName: name,
	}
	err := server.init(name)
	if err != nil {
		return nil, err
	}
	return server, nil
}

func (node *XTraceServerContainer) init(name string) error {
	workflow.Init("../../runtime")

	spec, err := workflow.GetSpec()
	if err != nil {
		return err
	}

	details, err := spec.Get("XTracerImpl")
	if err != nil {
		return err
	}

	node.Iface = details.Iface
	return nil
}

func (node *XTraceServerContainer) Name() string {
	return node.ServerName
}

func (node *XTraceServerContainer) String() string {
	return node.Name() + " = XTraceServer(" + node.BindAddr.Name() + ")"
}

func (node *XTraceServerContainer) GetInterface(ctx ir.BuildContext) (service.ServiceInterface, error) {
	iface := node.Iface.ServiceInterface(ctx)
	return &XTraceInterface{Wrapped: iface}, nil
}

func (node *XTraceServerContainer) AddContainerArtifacts(target docker.ContainerWorkspace) error {
	return nil
}

func (node *XTraceServerContainer) AddContainerInstance(target docker.ContainerWorkspace) error {
	node.BindAddr.Port = 5563
	return target.DeclarePrebuiltInstance(node.ServerName, "jonathanmace/xtrace-server:latest", node.BindAddr)
}
