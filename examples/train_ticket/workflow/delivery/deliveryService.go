//Package delivery implements ts-delivery service from the original train ticket application
package delivery

import (
	"context"
	"fmt"

	"github.com/blueprint-uservices/blueprint/runtime/core/backend"
	"golang.org/x/exp/slog"
)

// DeliveryService implements the Delivery microservice.
//
// It is not a service that can be called; instead it pulls deliveries from
// the delivery queue
type DeliveryService interface {
	Run(ctx context.Context) error
}

type DeliveryServiceImpl struct {
	db   backend.NoSQLDatabase
	delQ backend.Queue
}

func NewDeliveryServiceImpl(ctx context.Context, queue backend.Queue, db backend.NoSQLDatabase) (*DeliveryServiceImpl, error) {
	return &DeliveryServiceImpl{db: db, delQ: queue}, nil
}

func (d *DeliveryServiceImpl) Run(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return nil
		default:
			var delivery Delivery
			didpop, err := d.delQ.Pop(ctx, &delivery)
			if err != nil {
				slog.Error(fmt.Sprintf("DeliveryService unable to pull delivery info from deliver queue due to %v", err))
			}
			if didpop {
				coll, err := d.db.GetCollection(ctx, "delivery", "delivery")
				if err != nil {
					slog.Error(fmt.Sprintf("DeliveryService unable to obtain a collection to delivery database due to %v", err))
				}
				err = coll.InsertOne(ctx, delivery)
				if err != nil {
					slog.Error(fmt.Sprintf("DeliveryService unable to add a delivery to the database due to %v", err))
				}
			}
		}
	}
}
