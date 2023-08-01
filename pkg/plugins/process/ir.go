package process

import (
	"gitlab.mpi-sws.org/cld/blueprint/pkg/blueprint"
)

// This Node represents a Golang process that internally will instantiate a number of application-level services
type ProcessNode interface {
	blueprint.IRNode
}
