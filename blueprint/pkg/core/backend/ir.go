package backend

import (
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/core/service"
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/ir"
)

type (
	Cache interface {
		ir.IRNode
		service.ServiceNode
	}

	NoSQLDB interface {
		ir.IRNode
		service.ServiceNode
	}
)
