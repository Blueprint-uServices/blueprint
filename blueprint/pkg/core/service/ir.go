package service

import (
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/blueprint"
)

// Interface for IRNodes that are Call-Response Services
type ServiceNode interface {
	blueprint.IRNode
	GetInterface() *ServiceInterface
}
