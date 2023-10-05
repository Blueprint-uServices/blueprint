package service

import (
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/blueprint"
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/core/irutil"
)

// Interface for IRNodes that are Call-Response Services
type ServiceNode interface {
	blueprint.IRNode
	GetInterface(visitor irutil.BuildContext) ServiceInterface
}
