package backend

import (
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/blueprint"
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/core/service"
)

type (
	Cache interface {
		blueprint.IRNode
		service.ServiceNode
	}

	NoSQLDB interface {
		blueprint.IRNode
		service.ServiceNode
	}
)
