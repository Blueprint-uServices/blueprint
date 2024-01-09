package service

import "github.com/blueprint-uservices/blueprint/blueprint/pkg/ir"

/*
Interface for IRNodes that are Call-Response Services

At build time, services need to be able to provide information about the interface that they implement
*/

// Any IR node that represents a callable service should implement this interface.
type ServiceNode interface {

	// Returns the interface of this service
	GetInterface(ctx ir.BuildContext) (ServiceInterface, error)
}
