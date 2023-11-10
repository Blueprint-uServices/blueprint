package service

import "gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/ir"

/*
Interface for IRNodes that are Call-Response Services

At build time, services need to be able to provide information about the interface that they implement
*/
type ServiceNode interface {
	GetInterface(ctx ir.BuildContext) (ServiceInterface, error)
}
