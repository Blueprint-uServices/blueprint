package latencyinjector

import (
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/ir"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/golang"
)

// Blueprint IR Node representing a server side latency injector
type LatencyInjector struct {
	golang.Service
	golang.GeneratesFuncs
	golang.Instantiable

	InstanceName  string
	Wrapped       golang.Service
	outputPackage string
	LatencyValue  *ir.IRValue
}

func newLatencyInjector(name string, server ir.IRNode, latency string) (*LatencyInjector, error) {
	return nil, nil
}
