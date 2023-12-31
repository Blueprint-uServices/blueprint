package govector

import (
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/ir"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/golang"
)

// Blueprint IR Node that wraps the client-side of a service to generate govec logs
type GovecClientWrapper struct {
	golang.Service
	golang.Instantiable
	golang.GeneratesFuncs

	InstanceName  string
	outputPackage string
	Wrapped       golang.Service
	LoggerName    *ir.IRValue
}

func (node *GovecClientWrapper) Name() string {
	return node.InstanceName
}

func (node *GovecClientWrapper) String() string {
	return node.Name() + " = GovecClientWrapper(" + node.Wrapped.Name() + ")"
}

func (node *GovecClientWrapper) ImplementsGolangNode()    {}
func (node *GovecClientWrapper) ImplementsGolangService() {}

func newGovecClientWrapper(name string, wrapped golang.Service, logger_name string) (*GovecClientWrapper, error) {
	node := &GovecClientWrapper{}
	node.InstanceName = name
	node.outputPackage = "govec"
	node.Wrapped = wrapped
	node.LoggerName = &ir.IRValue{Value: logger_name}
	return node, nil
}
