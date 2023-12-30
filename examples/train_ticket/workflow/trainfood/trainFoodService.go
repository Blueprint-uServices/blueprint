// package trainfood implements ts-train-food-service from the original train ticket application
package trainfood

import (
	"context"
	"errors"

	"gitlab.mpi-sws.org/cld/blueprint/examples/train_ticket/workflow/food"
	"gitlab.mpi-sws.org/cld/blueprint/runtime/core/backend"
	"go.mongodb.org/mongo-driver/bson"
)

type TrainFoodService interface {
	CreateTrainFood(ctx context.Context, tf TrainFood) (TrainFood, error)
	ListTrainFood(ctx context.Context) ([]TrainFood, error)
	ListTrainFoodByTripID(ctx context.Context, tripid string) ([]food.Food, error)
	Cleanup(ctx context.Context) error
}

type TrainFoodServiceImpl struct {
	db backend.NoSQLDatabase
}

func NewTrainFoodServiceImpl(ctx context.Context, db backend.NoSQLDatabase) (*TrainFoodServiceImpl, error) {
	return &TrainFoodServiceImpl{db: db}, nil
}

func (t *TrainFoodServiceImpl) ListTrainFood(ctx context.Context) ([]TrainFood, error) {
	coll, err := t.db.GetCollection(ctx, "trainfood", "trainfood")
	if err != nil {
		return []TrainFood{}, err
	}
	res, err := coll.FindMany(ctx, bson.D{})
	if err != nil {
		return []TrainFood{}, err
	}
	var all_foods []TrainFood
	err = res.All(ctx, &all_foods)
	if err != nil {
		return []TrainFood{}, err
	}
	return all_foods, nil
}

func (t *TrainFoodServiceImpl) ListTrainFoodByTripID(ctx context.Context, tripid string) ([]food.Food, error) {
	coll, err := t.db.GetCollection(ctx, "trainfood", "trainfood")
	if err != nil {
		return []food.Food{}, err
	}
	res, err := coll.FindOne(ctx, bson.D{{"tripid", tripid}})
	if err != nil {
		return []food.Food{}, err
	}
	var tf TrainFood
	exists, err := res.One(ctx, &tf)
	if err != nil {
		return []food.Food{}, err
	}
	if !exists {
		return []food.Food{}, errors.New("Trip with Trip ID " + tripid + " does not exist")
	}
	return tf.Foods, nil
}

func (t *TrainFoodServiceImpl) CreateTrainFood(ctx context.Context, tf TrainFood) (TrainFood, error) {
	coll, err := t.db.GetCollection(ctx, "trainfood", "trainfood")
	if err != nil {
		return TrainFood{}, err
	}
	query := bson.D{{"tripid", tf.TripID}}
	res, err := coll.FindOne(ctx, query)
	if err != nil {
		return TrainFood{}, err
	}
	var stored_tf TrainFood
	exists, err := res.One(ctx, &stored_tf)
	if err != nil {
		return TrainFood{}, err
	}
	if !exists {
		return tf, coll.InsertOne(ctx, tf)
	}
	ok, err := coll.Upsert(ctx, query, tf)
	if err != nil {
		return TrainFood{}, err
	}
	if !ok {
		return TrainFood{}, errors.New("Failed to set the train food")
	}
	return tf, err
}

func (t *TrainFoodServiceImpl) Cleanup(ctx context.Context) error {
	coll, err := t.db.GetCollection(ctx, "trainfood", "trainfood")
	if err != nil {
		return err
	}
	return coll.DeleteMany(ctx, bson.D{})
}
