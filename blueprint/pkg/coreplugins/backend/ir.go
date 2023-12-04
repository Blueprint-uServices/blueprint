// Package backend provides IR node interfaces for common backend components.
package backend

import (
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/coreplugins/service"
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

	Queue interface {
		ir.IRNode
		service.ServiceNode
	}

	RelDB interface {
		ir.IRNode
		service.ServiceNode
	}
)
