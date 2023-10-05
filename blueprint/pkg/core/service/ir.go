package service

import "gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/blueprint"

/*
Interface for IRNodes that are Call-Response Services

At build time, services need to be able to provide information about the interface that they implement
*/
type ServiceNode interface {
	GetInterface(ctx blueprint.BuildContext) (ServiceInterface, error)
}
