package timeouts

import (
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/ir"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/golang"
)

// Blueprint IR node representing a Timeout node
type TimeoutClient struct {
	golang.Service
	golang.GeneratesFuncs
	golang.Instantiable

	InstanceName string
	Wrapped      golang.Service

	TimeoutValue  ir.IRNode
	outputPackage string
}

func newTimeoutClient(name string, server ir.IRNode, timeout string) (*TimeoutClient, error) {
	// TODO: Implement
	return nil, nil
}
