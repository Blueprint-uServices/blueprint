// Package consignprice implements ts-consignprice-service from the original train ticket application
package consignprice

import (
	"context"
	"errors"
	"fmt"

	"github.com/blueprint-uservices/blueprint/runtime/core/backend"
	"go.mongodb.org/mongo-driver/bson"
)

// ConsignPriceService manages the prices of consignments
type ConsignPriceService interface {
	// Calculates the price of the consignment based on the weight and the region
	GetPriceByWeightAndRegion(ctx context.Context, weight float64, isWithinRegion bool) (float64, error)
	// Get the price configuration for calculating consignment prices as a string
	GetPriceInfo(ctx context.Context) (string, error)
	// Get the price configuration for calculating consignment prices
	GetPriceConfig(ctx context.Context) (ConsignPrice, error)
	// Creates a price config or modifies the existing price configuration
	CreateAndModifyPriceConfig(ctx context.Context, priceConfig ConsignPrice) (ConsignPrice, error)
}

type ConsignPriceServiceImpl struct {
	db backend.NoSQLDatabase
}

func NewConsignPriceServiceImpl(ctx context.Context, db backend.NoSQLDatabase) (*ConsignPriceServiceImpl, error) {
	return &ConsignPriceServiceImpl{db: db}, nil
}

func (c *ConsignPriceServiceImpl) GetPriceByWeightAndRegion(ctx context.Context, weight float64, isWithinRegion bool) (float64, error) {
	cp, err := c.GetPriceConfig(ctx)
	if err != nil {
		return 0.0, err
	}
	if weight <= cp.InitialWeight {
		return cp.InitialPrice, err
	}
	price := cp.InitialPrice
	if isWithinRegion {
		price += (weight - cp.InitialWeight) * cp.WithinPrice
	} else {
		price += (weight - cp.InitialWeight) * cp.BeyondPrice
	}
	return price, nil
}

func (c *ConsignPriceServiceImpl) GetPriceInfo(ctx context.Context) (string, error) {
	coll, err := c.db.GetCollection(ctx, "consign", "consign")
	if err != nil {
		return "", err
	}
	query := bson.D{{"index", 0}}
	res, err := coll.FindOne(ctx, query)
	if err != nil {
		return "", err
	}
	var cp ConsignPrice
	exists, err := res.One(ctx, &cp)
	if err != nil {
		return "", err
	}
	if !exists {
		return "", errors.New("Consign Price Config doesn't exist")
	}
	info := fmt.Sprintf("The price of weight within %.2f is %.2f. The price of extra weight within the region is %.2f and beyond the region is %.2f", cp.InitialWeight, cp.InitialPrice, cp.WithinPrice, cp.BeyondPrice)
	return info, nil
}

func (c *ConsignPriceServiceImpl) GetPriceConfig(ctx context.Context) (ConsignPrice, error) {
	coll, err := c.db.GetCollection(ctx, "consign", "consign")
	if err != nil {
		return ConsignPrice{}, err
	}
	query := bson.D{{"index", 0}}
	res, err := coll.FindOne(ctx, query)
	if err != nil {
		return ConsignPrice{}, err
	}
	var cp ConsignPrice
	exists, err := res.One(ctx, &cp)
	if err != nil {
		return ConsignPrice{}, err
	}
	if !exists {
		return ConsignPrice{}, errors.New("Consign Price Config doesn't exist")
	}
	return cp, nil
}

func (c *ConsignPriceServiceImpl) CreateAndModifyPriceConfig(ctx context.Context, priceConfig ConsignPrice) (ConsignPrice, error) {
	coll, err := c.db.GetCollection(ctx, "consign", "consign")
	if err != nil {
		return ConsignPrice{}, err
	}
	query := bson.D{{"index", 0}}
	res, err := coll.FindOne(ctx, query)
	if err != nil {
		return ConsignPrice{}, err
	}
	var cp ConsignPrice
	exists, err := res.One(ctx, &cp)
	if err != nil {
		return ConsignPrice{}, err
	}
	priceConfig.Index = 0
	if exists {
		ok, err := coll.Upsert(ctx, bson.D{{"index", 0}}, priceConfig)
		if err != nil {
			return priceConfig, err
		}
		if !ok {
			return priceConfig, errors.New("Failed to update consignprice")
		}
	}
	return priceConfig, coll.InsertOne(ctx, priceConfig)
}
