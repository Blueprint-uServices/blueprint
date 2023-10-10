package xtrace

import "gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/core/process"

type XTraceServer struct {
	process.ProcessNode
	process.ArtifactGenerator

	ServerName string
	Addr       *GolangXTraceAddress
}

func newXTraceServer(name string, addr *GolangXTraceAddress) (*XTraceServer, error) {
	return &XTraceServer{
		ServerName: name,
		Addr:       addr,
	}, nil
}

func (node *XTraceServer) Name() string {
	return node.ServerName
}

func (node *XTraceServer) String() string {
	return node.Name() + " = XTraceServer(" + node.Addr.Name() + ")"
}

func (node *XTraceServer) GenerateArtifacts(outputDir string) error {
	// TODO: generate artifacts for the XTraceServer process
	return nil
}
