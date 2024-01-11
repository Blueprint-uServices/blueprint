// Package consign implements ts-consign-service from the original Train Ticket application
package consign

import (
	"context"
	"errors"

	"github.com/blueprint-uservices/blueprint/examples/train_ticket/workflow/consignprice"
	"github.com/blueprint-uservices/blueprint/runtime/core/backend"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
)

// Service managers the consignments in te application
type ConsignService interface {
	// Insert a consignment
	InsertConsign(ctx context.Context, c Consign) (Consign, error)
	// Update a consignment
	UpdateConsign(ctx context.Context, c Consign) (Consign, error)
	// Find all consignments by account ID
	FindByAccountId(ctx context.Context, accountId string) ([]Consign, error)
	// Find all consignments by order ID
	FindByOrderId(ctx context.Context, orderId string) ([]Consign, error)
	// Find all consignments by consignee
	FindByConsignee(ctx context.Context, consignee string) ([]Consign, error)
}

// Implementation of Consign Service
type ConsignServiceImpl struct {
	cps consignprice.ConsignPriceService
	db  backend.NoSQLDatabase
}

// Returns a new object of consign service
func NewConsignServiceImpl(ctx context.Context, cps consignprice.ConsignPriceService, db backend.NoSQLDatabase) (*ConsignServiceImpl, error) {
	return &ConsignServiceImpl{cps: cps, db: db}, nil
}

// Implementation of ConsignService
func (csi *ConsignServiceImpl) InsertConsign(ctx context.Context, c Consign) (Consign, error) {
	price, err := csi.cps.GetPriceByWeightAndRegion(ctx, c.Weight, c.Within)
	if err != nil {
		return Consign{}, err
	}

	c.Price = price
	c.Id = uuid.New().String()

	collection, err := csi.db.GetCollection(ctx, "consign", "consign")
	if err != nil {
		return Consign{}, err
	}
	err = collection.InsertOne(ctx, c)
	if err != nil {
		return Consign{}, err
	}
	return c, nil
}

// Implementation of ConsignService
func (csi *ConsignServiceImpl) UpdateConsign(ctx context.Context, c Consign) (Consign, error) {
	collection, err := csi.db.GetCollection(ctx, "consign", "consign")
	if err != nil {
		return Consign{}, err
	}

	query := bson.D{{"id", c.Id}}

	ok, err := collection.Upsert(ctx, query, c)
	if err != nil {
		return Consign{}, err
	}
	if !ok {
		return Consign{}, errors.New("Failed to update consign")
	}
	return c, nil
}

// Implementation of ConsignService
func (csi *ConsignServiceImpl) FindByAccountId(ctx context.Context, accountId string) ([]Consign, error) {
	var consigns []Consign
	collection, err := csi.db.GetCollection(ctx, "consign", "consign")
	if err != nil {
		return consigns, err
	}
	query := bson.D{{"accountid", accountId}}

	result, err := collection.FindMany(ctx, query)
	if err != nil {
		return nil, err
	}

	err = result.All(ctx, &consigns)
	if err != nil {
		return consigns, err
	}

	return consigns, nil
}

// Implementation of ConsignService
func (csi *ConsignServiceImpl) FindByOrderId(ctx context.Context, orderId string) ([]Consign, error) {
	var consigns []Consign
	collection, err := csi.db.GetCollection(ctx, "consign", "consign")
	if err != nil {
		return consigns, err
	}
	query := bson.D{{"orderid", orderId}}

	result, err := collection.FindMany(ctx, query)
	if err != nil {
		return consigns, err
	}
	if err != nil {
		return consigns, err
	}

	err = result.All(ctx, &consigns)
	if err != nil {
		return consigns, err
	}

	return consigns, nil
}

// Implementation of ConsignService
func (csi *ConsignServiceImpl) FindByConsignee(ctx context.Context, consignee string) ([]Consign, error) {
	var consigns []Consign
	collection, err := csi.db.GetCollection(ctx, "consign", "consign")
	if err != nil {
		return consigns, err
	}
	query := bson.D{{"consignee", consignee}}

	result, err := collection.FindMany(ctx, query)
	if err != nil {
		return consigns, err
	}

	err = result.All(ctx, &consigns)
	if err != nil {
		return consigns, err
	}

	return consigns, nil
}
