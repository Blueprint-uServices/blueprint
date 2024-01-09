// Package price provides an implementation of the PriceService
// PriceService uses a backend.NoSQLDatabase to store price config data
package price

import (
	"context"
	"errors"
	"strings"

	"github.com/blueprint-uservices/blueprint/runtime/core/backend"
	"go.mongodb.org/mongo-driver/bson"
)

type PriceService interface {
	FindByID(ctx context.Context, id string) (PriceConfig, error)
	CreateNewPriceConfig(ctx context.Context, config PriceConfig) error
	FindByRouteIDAndTrainType(ctx context.Context, routeID string, trainType string) (PriceConfig, error)
	FindByRouteIDsAndTrainTypes(ctx context.Context, rtsAndTypes []string) (map[string]PriceConfig, error)
	GetAllPriceConfig(ctx context.Context) ([]PriceConfig, error)
	DeletePriceConfig(ctx context.Context, id string) error
	UpdatePriceConfig(ctx context.Context, config PriceConfig) (bool, error)
}

type PriceServiceImpl struct {
	priceDB backend.NoSQLDatabase
}

func NewPriceServiceImpl(ctx context.Context, db backend.NoSQLDatabase) (*PriceServiceImpl, error) {
	return &PriceServiceImpl{priceDB: db}, nil
}

func (p *PriceServiceImpl) FindByID(ctx context.Context, id string) (PriceConfig, error) {
	coll, err := p.priceDB.GetCollection(ctx, "priceConfig", "priceConfig")
	if err != nil {
		return PriceConfig{}, err
	}
	query := bson.D{{"id", id}}
	res, err := coll.FindOne(ctx, query)
	if err != nil {
		return PriceConfig{}, err
	}
	var pc PriceConfig
	exists, err := res.One(ctx, &pc)
	if err != nil {
		return PriceConfig{}, err
	}
	if !exists {
		return PriceConfig{}, errors.New("PriceConfig with ID " + id + " does not exist")
	}
	return pc, nil
}

func (p *PriceServiceImpl) GetAllPriceConfig(ctx context.Context) ([]PriceConfig, error) {
	coll, err := p.priceDB.GetCollection(ctx, "priceConfig", "priceConfig")
	if err != nil {
		return []PriceConfig{}, err
	}
	res, err := coll.FindMany(ctx, bson.D{})
	if err != nil {
		return []PriceConfig{}, err
	}
	var pcs []PriceConfig
	err = res.All(ctx, &pcs)
	if err != nil {
		return []PriceConfig{}, err
	}
	return pcs, nil
}

func (p *PriceServiceImpl) DeletePriceConfig(ctx context.Context, id string) error {
	coll, err := p.priceDB.GetCollection(ctx, "priceConfig", "priceConfig")
	if err != nil {
		return err
	}
	query := bson.D{{"id", id}}
	return coll.DeleteOne(ctx, query)
}

func (p *PriceServiceImpl) UpdatePriceConfig(ctx context.Context, pc PriceConfig) (bool, error) {
	coll, err := p.priceDB.GetCollection(ctx, "priceConfig", "priceConfig")
	if err != nil {
		return false, err
	}
	query := bson.D{{"id", pc.ID}}
	return coll.Upsert(ctx, query, pc)
}

func (p *PriceServiceImpl) CreateNewPriceConfig(ctx context.Context, pc PriceConfig) error {

	coll, err := p.priceDB.GetCollection(ctx, "priceConfig", "priceConfig")
	if err != nil {
		return err
	}
	_, err = p.FindByID(ctx, pc.ID)
	if err != nil {
		return coll.InsertOne(ctx, pc)
	} else {
		query := bson.D{{"id", pc.ID}}
		ok, err := coll.Upsert(ctx, query, pc)
		if err != nil {
			return err
		}
		if !ok {
			return errors.New("Failed to update an existing price config")
		}
	}
	return nil
}

func (p *PriceServiceImpl) FindByRouteIDAndTrainType(ctx context.Context, routeID string, trainType string) (PriceConfig, error) {
	coll, err := p.priceDB.GetCollection(ctx, "priceConfig", "priceConfig")
	if err != nil {
		return PriceConfig{}, err
	}
	query := bson.D{{"$and", bson.A{
		bson.D{{"routeid", routeID}},
		bson.D{{"traintype", trainType}},
	}}}

	res, err := coll.FindOne(ctx, query)
	if err != nil {
		return PriceConfig{}, err
	}
	var pc PriceConfig
	exists, err := res.One(ctx, &pc)
	if err != nil {
		return PriceConfig{}, err
	}
	if !exists {
		return PriceConfig{}, errors.New("PriceConfig with routeId:trainType " + routeID + ":" + trainType + " does not exist")
	}
	return pc, nil
}

func (p *PriceServiceImpl) FindByRouteIDsAndTrainTypes(ctx context.Context, rtsAndTypes []string) (map[string]PriceConfig, error) {
	res := make(map[string]PriceConfig)
	// TODO: Maybe implement this as a single query
	for _, rt := range rtsAndTypes {
		pieces := strings.Split(rt, ":")
		routeid := pieces[0]
		trainType := pieces[1]
		pc, err := p.FindByRouteIDAndTrainType(ctx, routeid, trainType)
		// Ignore error
		if err == nil {
			res[rt] = pc
		}
	}
	return res, nil
}
