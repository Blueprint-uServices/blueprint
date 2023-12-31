package govector

import (
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/ir"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/golang"
)

// Blueprint IR node that wraps the server-side of a service to generate govec compatible logs
type GovecServerWrapper struct {
	golang.Service
	golang.Instantiable
	golang.GeneratesFuncs

	InstanceName  string
	outputPackage string
	Wrapped       golang.Service
	LoggerName    *ir.IRValue
}

func (node *GovecServerWrapper) Name() string {
	return node.InstanceName
}

func (node *GovecServerWrapper) String() string {
	return node.Name() + " = GovecServerWrapper(" + node.Wrapped.Name() + ")"
}

func (node *GovecServerWrapper) ImplementsGolangNode()    {}
func (node *GovecServerWrapper) ImplementsGolangService() {}

func newGovecServerWrapper(name string, wrapped golang.Service, logger_name string) (*GovecServerWrapper, error) {
	node := &GovecServerWrapper{}
	node.InstanceName = name
	node.outputPackage = "govec"
	node.Wrapped = wrapped
	node.LoggerName = &ir.IRValue{Value: logger_name}
	return node, nil
}
