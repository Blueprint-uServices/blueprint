package xtrace

import (
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/blueprint"
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/core/process"
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/core/service"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/golang/goparser"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/workflow"
)

type XTraceServer struct {
	process.ProcessNode
	process.ArtifactGenerator

	ServerName string
	Addr       *GolangXTraceAddress
	Iface      *goparser.ParsedInterface
}

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

func newXTraceServer(name string, addr *GolangXTraceAddress) (*XTraceServer, error) {
	server := &XTraceServer{
		ServerName: name,
		Addr:       addr,
	}
	err := server.init(name)
	if err != nil {
		return nil, err
	}
	return server, nil
}

func (node *XTraceServer) init(name string) error {
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

func (node *XTraceServer) Name() string {
	return node.ServerName
}

func (node *XTraceServer) String() string {
	return node.Name() + " = XTraceServer(" + node.Addr.Name() + ")"
}

func (node *XTraceServer) GetInterface(ctx blueprint.BuildContext) (service.ServiceInterface, error) {
	iface := node.Iface.ServiceInterface(ctx)
	return &XTraceInterface{Wrapped: iface}, nil
}

func (node *XTraceServer) GenerateArtifacts(outputDir string) error {
	// TODO: generate artifacts for the XTraceServer process
	return nil
}
