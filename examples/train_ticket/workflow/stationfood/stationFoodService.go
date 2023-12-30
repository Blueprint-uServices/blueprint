// package stationfood implements ts-station-food-service from the original Train Ticket application
package stationfood

import (
	"context"
	"errors"

	"gitlab.mpi-sws.org/cld/blueprint/runtime/core/backend"
	"go.mongodb.org/mongo-driver/bson"
)

type StationFoodService interface {
	CreateFoodStore(ctx context.Context, store StationFoodStore) error
	ListFoodStores(ctx context.Context) ([]StationFoodStore, error)
	ListFoodStoresByStationName(ctx context.Context, station string) ([]StationFoodStore, error)
	GetFoodStoresByStationNames(ctx context.Context, stations []string) ([]StationFoodStore, error)
	GetFoodStoreByID(ctx context.Context, id string) (StationFoodStore, error)
	Cleanup(ctx context.Context) error
}

type StationFoodServiceImpl struct {
	db backend.NoSQLDatabase
}

func NewStationFoodServiceImpl(ctx context.Context, db backend.NoSQLDatabase) (*StationFoodServiceImpl, error) {
	return &StationFoodServiceImpl{db: db}, nil
}

func (s *StationFoodServiceImpl) CreateFoodStore(ctx context.Context, store StationFoodStore) error {
	coll, err := s.db.GetCollection(ctx, "stationfood", "stationfood")
	if err != nil {
		return err
	}
	query := bson.D{{"id", store.ID}}
	res, err := coll.FindOne(ctx, query)
	if err != nil {
		return err
	}
	var st StationFoodStore
	exists, err := res.One(ctx, &st)
	if exists {
		return errors.New("Station Food Store with id " + store.ID + " already exists")
	}
	return coll.InsertOne(ctx, store)
}

func (s *StationFoodServiceImpl) ListFoodStores(ctx context.Context) ([]StationFoodStore, error) {
	coll, err := s.db.GetCollection(ctx, "stationfood", "stationfood")
	if err != nil {
		return []StationFoodStore{}, err
	}
	res, err := coll.FindMany(ctx, bson.D{})
	if err != nil {
		return []StationFoodStore{}, err
	}
	var stores []StationFoodStore
	err = res.All(ctx, &stores)
	if err != nil {
		return []StationFoodStore{}, err
	}
	return stores, nil
}

func (s *StationFoodServiceImpl) ListFoodStoresByStationName(ctx context.Context, station string) ([]StationFoodStore, error) {
	coll, err := s.db.GetCollection(ctx, "stationfood", "stationfood")
	if err != nil {
		return []StationFoodStore{}, err
	}
	query := bson.D{{"stationname", station}}
	res, err := coll.FindMany(ctx, query)
	if err != nil {
		return []StationFoodStore{}, err
	}
	var stores []StationFoodStore
	err = res.All(ctx, &stores)
	if err != nil {
		return []StationFoodStore{}, err
	}
	return stores, nil
}

func (s *StationFoodServiceImpl) GetFoodStoresByStationNames(ctx context.Context, stations []string) ([]StationFoodStore, error) {
	coll, err := s.db.GetCollection(ctx, "stationfood", "stationfood")
	if err != nil {
		return []StationFoodStore{}, err
	}
	doc := bson.A{}
	for _, station := range stations {
		doc = append(doc, station)
	}
	query := bson.D{{"stationname", bson.D{{"$in", doc}}}}
	res, err := coll.FindMany(ctx, query)
	if err != nil {
		return []StationFoodStore{}, err
	}
	var stores []StationFoodStore
	err = res.All(ctx, &stores)
	if err != nil {
		return []StationFoodStore{}, err
	}
	return stores, nil
}

func (s *StationFoodServiceImpl) GetFoodStoreByID(ctx context.Context, id string) (StationFoodStore, error) {
	coll, err := s.db.GetCollection(ctx, "stationfood", "stationfood")
	if err != nil {
		return StationFoodStore{}, err
	}
	if err != nil {
		return StationFoodStore{}, err
	}
	res, err := coll.FindOne(ctx, bson.D{{"id", id}})
	if err != nil {
		return StationFoodStore{}, err
	}
	var store StationFoodStore
	exists, err := res.One(ctx, &store)
	if err != nil {
		return StationFoodStore{}, err
	}
	if !exists {
		return StationFoodStore{}, errors.New("Station with ID " + id + " does not exist")
	}
	return store, err
}

func (s *StationFoodServiceImpl) Cleanup(ctx context.Context) error {
	coll, err := s.db.GetCollection(ctx, "stationfood", "stationfood")
	if err != nil {
		return err
	}
	return coll.DeleteMany(ctx, bson.D{})
}
