package backend

import (
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/blueprint"
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/core/service"
)

type Cache struct {
	blueprint.IRNode
	service.ServiceNode
}

func (c *Cache) GetInterface() *service.ServiceInterface {
	// TODO: return the cache interface.  might need to change serviceinterface a bit
	return nil
}
