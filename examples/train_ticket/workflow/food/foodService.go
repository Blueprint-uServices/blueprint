// Package food implements ts-food-service from the original train ticket application
package food

import (
	"context"
	"errors"

	"github.com/blueprint-uservices/blueprint/examples/train_ticket/workflow/common"
	"github.com/blueprint-uservices/blueprint/examples/train_ticket/workflow/stationfood"
	"github.com/blueprint-uservices/blueprint/examples/train_ticket/workflow/trainfood"
	"github.com/blueprint-uservices/blueprint/examples/train_ticket/workflow/travel"
	"github.com/blueprint-uservices/blueprint/runtime/core/backend"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"golang.org/x/exp/slices"
)

type FoodService interface {
	CreateFoodOrder(ctx context.Context, fo FoodOrder) (FoodOrder, error)
	CreateFoodOrdersInBatch(ctx context.Context, fos []FoodOrder) ([]FoodOrder, error)
	DeleteFoodOrder(ctx context.Context, orderId string) error
	FindByOrderId(ctx context.Context, orderId string) (FoodOrder, error)
	UpdateFoodOrder(ctx context.Context, fo FoodOrder) (bool, error)
	FindAllFoodOrder(ctx context.Context) ([]FoodOrder, error)
	GetAllFood(ctx context.Context, date string, from string, to string, tripid string) ([]common.Food, map[string][]stationfood.StationFoodStore, error)
}

type FoodServiceImpl struct {
	db                 backend.NoSQLDatabase
	trainfoodService   trainfood.TrainFoodService
	stationfoodService stationfood.StationFoodService
	travelService      travel.TravelService
}

func NewFoodServiceImpl(ctx context.Context, db backend.NoSQLDatabase, trainfoodService trainfood.TrainFoodService, stationfoodService stationfood.StationFoodService, travelService travel.TravelService) (*FoodServiceImpl, error) {
	return &FoodServiceImpl{db: db, trainfoodService: trainfoodService, stationfoodService: stationfoodService, travelService: travelService}, nil
}

func (f *FoodServiceImpl) CreateFoodOrder(ctx context.Context, fo FoodOrder) (FoodOrder, error) {
	coll, err := f.db.GetCollection(ctx, "food", "food")
	if err != nil {
		return fo, err
	}
	res, err := coll.FindOne(ctx, bson.D{{"orderid", fo.OrderID}})
	if err != nil {
		return fo, err
	}
	var existing FoodOrder
	ok, err := res.One(ctx, &existing)
	if err != nil {
		return fo, err
	}
	if ok {
		return fo, errors.New("Food order with orderID " + fo.OrderID + "already exists")
	}
	fo.ID = uuid.New().String()
	err = coll.InsertOne(ctx, fo)
	return fo, err
}

func (f *FoodServiceImpl) CreateFoodOrdersInBatch(ctx context.Context, fos []FoodOrder) ([]FoodOrder, error) {
	var res []FoodOrder
	for _, fo := range fos {
		var err error
		fo, err = f.CreateFoodOrder(ctx, fo)
		if err != nil {
			return res, err
		}
		res = append(res, fo)
	}
	return res, nil
}

func (f *FoodServiceImpl) DeleteFoodOrder(ctx context.Context, orderId string) error {
	coll, err := f.db.GetCollection(ctx, "food", "food")
	if err != nil {
		return err
	}
	return coll.DeleteOne(ctx, bson.D{{"orderid", orderId}})
}

func (f *FoodServiceImpl) UpdateFoodOrder(ctx context.Context, fo FoodOrder) (bool, error) {
	coll, err := f.db.GetCollection(ctx, "food", "food")
	if err != nil {
		return false, nil
	}
	return coll.Upsert(ctx, bson.D{{"orderid", fo.OrderID}}, fo)
}

func (f *FoodServiceImpl) FindByOrderId(ctx context.Context, orderId string) (FoodOrder, error) {
	coll, err := f.db.GetCollection(ctx, "food", "food")
	if err != nil {
		return FoodOrder{}, err
	}
	res, err := coll.FindOne(ctx, bson.D{{"orderid", orderId}})
	if err != nil {
		return FoodOrder{}, err
	}
	var fo FoodOrder
	ok, err := res.One(ctx, &fo)
	if err != nil {
		return fo, err
	}
	if !ok {
		return fo, errors.New("Food Order with ID " + orderId + " does not exist")
	}
	return fo, nil
}

func (f *FoodServiceImpl) FindAllFoodOrder(ctx context.Context) ([]FoodOrder, error) {
	var all_orders []FoodOrder
	coll, err := f.db.GetCollection(ctx, "food", "food")
	if err != nil {
		return all_orders, err
	}
	res, err := coll.FindMany(ctx, bson.D{})
	if err != nil {
		return all_orders, err
	}
	err = res.All(ctx, &all_orders)
	return all_orders, err
}

func (f *FoodServiceImpl) GetAllFood(ctx context.Context, date string, from string, to string, tripid string) ([]common.Food, map[string][]stationfood.StationFoodStore, error) {
	var all_foods []common.Food
	foodstores := make(map[string][]stationfood.StationFoodStore)
	if len(tripid) < 3 {
		return all_foods, foodstores, errors.New("Trip ID is not suitable")
	}
	all_foods, err := f.trainfoodService.ListTrainFoodByTripID(ctx, tripid)
	if err != nil {
		return all_foods, foodstores, err
	}

	r, err := f.travelService.GetRouteByTripId(ctx, tripid)
	if err != nil {
		return all_foods, foodstores, err
	}
	stations := r.Stations
	start_index := slices.Index(stations, from)
	end_index := slices.Index(stations, to)
	if start_index == -1 || end_index == -1 {
		return all_foods, foodstores, errors.New("Start and End stations are not on the selected route for the trip")
	}
	stations = stations[start_index : end_index+1]
	stores, err := f.stationfoodService.GetFoodStoresByStationNames(ctx, stations)
	if err != nil {
		return all_foods, foodstores, err
	}

	for _, s := range stores {
		if v, ok := foodstores[s.StationName]; ok {
			foodstores[s.StationName] = append(v, s)
		} else {
			foodstores[s.StationName] = []stationfood.StationFoodStore{s}
		}
	}

	return all_foods, foodstores, nil
}
