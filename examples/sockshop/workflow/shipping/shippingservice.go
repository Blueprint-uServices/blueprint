// Package shipping implements the SockShop shipping microservice.
//
// All the shipping microservice does is push the shipment to a queue.
// The queue-master service pulls shipments from the queue and "processes"
// them.
package shipping

import (
	"context"

	"gitlab.mpi-sws.org/cld/blueprint/runtime/core/backend"
)

// ShippingService implements the SockShop shipping microservice
type ShippingService interface {
	// Submit a shipment to be shipped.  The actual handling of the
	// shipment will happen asynchronously by the queue-master service.
	//
	// Returns the submitted shipment or an error
	PostShipping(ctx context.Context, shipment Shipment) (Shipment, error)
}

// Represents a shipment for an order
type Shipment struct {
	ID   string
	Name string
}

// Instantiates a shipping service that submits all shipments to a queue for asynchronous background processing
func NewShippingService(ctx context.Context, queue backend.Queue) (ShippingService, error) {
	return &shippingImpl{
		q: queue,
	}, nil
}

type shippingImpl struct {
	q backend.Queue
}

// PostShipping implements ShippingService.
func (service *shippingImpl) PostShipping(ctx context.Context, shipment Shipment) (Shipment, error) {
	return shipment, service.q.Push(ctx, shipment)
}
