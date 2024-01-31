// Package fooddelivery implements ts-food-delivery-service from the original application
package fooddelivery

import (
	"context"
	"errors"

	"github.com/blueprint-uservices/blueprint/examples/train_ticket/workflow/stationfood"
	"github.com/blueprint-uservices/blueprint/runtime/core/backend"
	"go.mongodb.org/mongo-driver/bson"
)

type FoodDeliveryService interface {
	CreateFoodDeliveryOrder(ctx context.Context, o FoodDeliveryOrder) (FoodDeliveryOrder, error)
	DeleteFoodDeliveryOrder(ctx context.Context, id string) error
	GetFoodDeliveryOrderByID(ctx context.Context, id string) (FoodDeliveryOrder, error)
	GetAllOrders(ctx context.Context) ([]FoodDeliveryOrder, error)
	GetOrdersByStoreID(ctx context.Context, store_id string) ([]FoodDeliveryOrder, error)
	UpdateTripID(ctx context.Context, info TripOrderInfo) error
	UpdateSeatNum(ctx context.Context, info SeatInfo) error
	UpdateDeliveryTime(ctx context.Context, info DeliveryInfo) error
}

type FoodDeliveryServiceImpl struct {
	db                 backend.NoSQLDatabase
	stationFoodService stationfood.StationFoodService
}

func NewFoodDeliveryServiceImpl(ctx context.Context, db backend.NoSQLDatabase, stationFoodService stationfood.StationFoodService) (*FoodDeliveryServiceImpl, error) {
	return &FoodDeliveryServiceImpl{db, stationFoodService}, nil
}

func (f *FoodDeliveryServiceImpl) CreateFoodDeliveryOrder(ctx context.Context, o FoodDeliveryOrder) (FoodDeliveryOrder, error) {
	store, err := f.stationFoodService.GetFoodStoreByID(ctx, o.StationFoodStoreID)
	if err != nil {
		return o, err
	}
	priceMap := make(map[string]float64)
	for _, fd := range store.Foods {
		priceMap[fd.Name] = fd.Price
	}
	var deliveryFee float64
	for _, fd := range o.FoodList {
		if v, ok := priceMap[fd]; ok {
			deliveryFee += v
		} else {
			return o, errors.New("Food not in list")
		}
	}
	o.DeliveryFee = deliveryFee

	coll, err := f.db.GetCollection(ctx, "fooddel", "fooddel")
	if err != nil {
		return o, err
	}
	err = coll.InsertOne(ctx, o)
	return o, err
}

func (f *FoodDeliveryServiceImpl) DeleteFoodDeliveryOrder(ctx context.Context, id string) error {
	coll, err := f.db.GetCollection(ctx, "fooddel", "fooddel")
	if err != nil {
		return err
	}
	return coll.DeleteOne(ctx, bson.D{{"id", id}})
}

func (f *FoodDeliveryServiceImpl) GetFoodDeliveryOrderByID(ctx context.Context, id string) (FoodDeliveryOrder, error) {
	coll, err := f.db.GetCollection(ctx, "fooddel", "fooddel")
	if err != nil {
		return FoodDeliveryOrder{}, err
	}
	res, err := coll.FindOne(ctx, bson.D{{"id", id}})
	if err != nil {
		return FoodDeliveryOrder{}, err
	}
	var retval FoodDeliveryOrder
	ok, err := res.One(ctx, &retval)
	if err != nil {
		return retval, err
	}
	if !ok {
		return retval, errors.New("Order with id " + id + " does not exist")
	}
	return retval, nil
}

func (f *FoodDeliveryServiceImpl) GetAllOrders(ctx context.Context) ([]FoodDeliveryOrder, error) {
	var all_orders []FoodDeliveryOrder
	coll, err := f.db.GetCollection(ctx, "fooddel", "fooddel")
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

func (f *FoodDeliveryServiceImpl) GetOrdersByStoreID(ctx context.Context, store_id string) ([]FoodDeliveryOrder, error) {
	var all_orders []FoodDeliveryOrder
	coll, err := f.db.GetCollection(ctx, "fooddel", "fooddel")
	if err != nil {
		return all_orders, err
	}
	res, err := coll.FindMany(ctx, bson.D{{"stationfoodstoreid", store_id}})
	if err != nil {
		return all_orders, err
	}
	err = res.All(ctx, &all_orders)
	return all_orders, err
}

func (f *FoodDeliveryServiceImpl) UpdateTripID(ctx context.Context, info TripOrderInfo) error {
	coll, err := f.db.GetCollection(ctx, "fooddel", "fooddel")
	if err != nil {
		return err
	}
	c, err := coll.UpdateOne(ctx, bson.D{{"id", info.OrderID}}, bson.D{{"$set", bson.D{{"tripid", info.TripID}}}})
	if err != nil {
		return err
	}
	if c != 1 {
		return errors.New("Failed to update trip ID")
	}
	return nil
}

func (f *FoodDeliveryServiceImpl) UpdateSeatNum(ctx context.Context, info SeatInfo) error {
	coll, err := f.db.GetCollection(ctx, "fooddel", "fooddel")
	if err != nil {
		return err
	}
	c, err := coll.UpdateOne(ctx, bson.D{{"id", info.OrderID}}, bson.D{{"$set", bson.D{{"seatnum", info.SeatNum}}}})
	if err != nil {
		return err
	}
	if c != 1 {
		return errors.New("Failed to update trip ID")
	}
	return nil
}

func (f *FoodDeliveryServiceImpl) UpdateDeliveryTime(ctx context.Context, info DeliveryInfo) error {
	coll, err := f.db.GetCollection(ctx, "fooddel", "fooddel")
	if err != nil {
		return err
	}
	c, err := coll.UpdateOne(ctx, bson.D{{"id", info.OrderID}}, bson.D{{"$set", bson.D{{"deliverytime", info.DeliveryTime}}}})
	if err != nil {
		return err
	}
	if c != 1 {
		return errors.New("Failed to update trip ID")
	}
	return nil
}
