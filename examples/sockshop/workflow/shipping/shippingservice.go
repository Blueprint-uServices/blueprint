// Package shipping implements the SockShop shipping microservice.
//
// All the shipping microservice does is push the shipment to a queue.
// The queue-master service pulls shipments from the queue and "processes"
// them.
package shipping

import (
	"context"
	"fmt"

	"github.com/blueprint-uservices/blueprint/runtime/core/backend"
	"go.mongodb.org/mongo-driver/bson"
)

// ShippingService implements the SockShop shipping microservice
type ShippingService interface {
	// Submit a shipment to be shipped.  The actual handling of the
	// shipment will happen asynchronously by the queue-master service.
	//
	// Returns the submitted shipment or an error
	PostShipping(ctx context.Context, shipment Shipment) (Shipment, error)

	// Get a shipment's status
	GetShipment(ctx context.Context, id string) (Shipment, error)

	// Update a shipment's status; called by the queue master
	UpdateStatus(ctx context.Context, id, status string) error
}

// Represents a shipment for an order
type Shipment struct {
	ID     string
	Name   string
	Status string
}

// Instantiates a shipping service that submits all shipments to a queue for asynchronous background processing
func NewShippingService(ctx context.Context, queue backend.Queue, db backend.NoSQLDatabase) (ShippingService, error) {
	c, err := db.GetCollection(ctx, "shipping_service", "shipments")
	return &shippingImpl{
		q:  queue,
		db: c,
	}, err
}

type shippingImpl struct {
	q  backend.Queue
	db backend.NoSQLCollection
}

// PostShipping implements ShippingService.
func (service *shippingImpl) PostShipping(ctx context.Context, shipment Shipment) (Shipment, error) {
	// Push to the queue to be shipped
	shipped, err := service.q.Push(ctx, shipment)
	if err != nil {
		return shipment, err
	} else if !shipped {
		return shipment, fmt.Errorf("Unable to submit shipment %v %v to the shipping queue", shipment.ID, shipment.Name)
	}

	// Insert into the shipment DB
	return shipment, service.db.InsertOne(ctx, shipment)
}

// GetShipment implements ShippingService.
func (s *shippingImpl) GetShipment(ctx context.Context, id string) (Shipment, error) {
	cursor, err := s.db.FindOne(ctx, bson.D{{"id", id}})
	if err != nil {
		return Shipment{}, err
	}
	var shipment Shipment
	shipmentExists, err := cursor.One(ctx, &shipment)
	if err != nil {
		return Shipment{}, err
	} else if !shipmentExists {
		return Shipment{}, fmt.Errorf("unknown shipment %v", id)
	}
	return shipment, nil
}

// UpdateStatus implements ShippingService.
func (s *shippingImpl) UpdateStatus(ctx context.Context, id string, status string) error {
	updated, err := s.db.UpdateOne(ctx, bson.D{{"id", id}}, bson.D{{"$set", bson.D{{"status", status}}}})
	if err != nil {
		return err
	} else if updated == 0 {
		return fmt.Errorf("unknown shipment %v", id)
	}
	return nil
}
