package train

import (
	"context"
	"errors"

	"gitlab.mpi-sws.org/cld/blueprint/runtime/core/backend"
	"go.mongodb.org/mongo-driver/bson"
)

type TrainService interface {
	Create(ctx context.Context, ttype TrainType) (bool, error)
	Retrieve(ctx context.Context, id string) (TrainType, error)
	RetrieveByName(ctx context.Context, name string) (TrainType, error)
	RetrieveByNames(ctx context.Context, names []string) ([]TrainType, error)
	Update(ctx context.Context, ttype TrainType) (bool, error)
	Delete(ctx context.Context, id string) (bool, error)
	AllTrains(ctx context.Context) ([]TrainType, error)
}

type TrainServiceImpl struct {
	db backend.NoSQLDatabase
}

func NewTrainServiceImpl(ctx context.Context, db backend.NoSQLDatabase) (*TrainServiceImpl, error) {
	return &TrainServiceImpl{db: db}, nil
}

func (ts *TrainServiceImpl) Create(ctx context.Context, tt TrainType) (bool, error) {
	coll, err := ts.db.GetCollection(ctx, "train", "train")
	if err != nil {
		return false, err
	}
	query := bson.D{{"name", tt.Name}}
	res, err := coll.FindOne(ctx, query)
	if err != nil {
		return false, err
	}
	var saved_tt TrainType
	exists, err := res.One(ctx, &saved_tt)
	if err != nil {
		return false, err
	}
	if exists {
		return false, errors.New("TrainType already exists")
	}

	err = coll.InsertOne(ctx, tt)
	if err != nil {
		return false, err
	}
	return true, nil
}

func (ts *TrainServiceImpl) Retrieve(ctx context.Context, id string) (TrainType, error) {
	coll, err := ts.db.GetCollection(ctx, "train", "train")
	if err != nil {
		return TrainType{}, err
	}
	query := bson.D{{"id", id}}
	res, err := coll.FindOne(ctx, query)
	if err != nil {
		return TrainType{}, err
	}
	var tt TrainType
	exists, err := res.One(ctx, &tt)
	if err != nil {
		return TrainType{}, err
	}
	if !exists {
		return TrainType{}, errors.New("TrainType with ID " + id + "does not exist")
	}
	return tt, nil
}

func (ts *TrainServiceImpl) RetrieveByName(ctx context.Context, name string) (TrainType, error) {
	coll, err := ts.db.GetCollection(ctx, "train", "train")
	if err != nil {
		return TrainType{}, err
	}
	query := bson.D{{"name", name}}
	res, err := coll.FindOne(ctx, query)
	if err != nil {
		return TrainType{}, err
	}
	var tt TrainType
	exists, err := res.One(ctx, &tt)
	if err != nil {
		return TrainType{}, err
	}
	if !exists {
		return TrainType{}, errors.New("TrainType with name " + name + "does not exist")
	}
	return tt, nil
}

func (ts *TrainServiceImpl) RetrieveByNames(ctx context.Context, names []string) ([]TrainType, error) {
	var trainTypes []TrainType
	for _, name := range names {
		tt, err := ts.RetrieveByName(ctx, name)
		if err == nil {
			trainTypes = append(trainTypes, tt)
		} else {
			trainTypes = append(trainTypes, TrainType{})
		}
	}
	return trainTypes, nil
}

func (ts *TrainServiceImpl) Update(ctx context.Context, ttype TrainType) (bool, error) {
	coll, err := ts.db.GetCollection(ctx, "train", "train")
	if err != nil {
		return false, err
	}
	query := bson.D{{"id", ttype.ID}}
	return coll.Upsert(ctx, query, ttype)
}

func (ts *TrainServiceImpl) Delete(ctx context.Context, id string) (bool, error) {
	coll, err := ts.db.GetCollection(ctx, "train", "train")
	if err != nil {
		return false, err
	}
	err = coll.DeleteOne(ctx, bson.D{{"id", id}})
	if err != nil {
		return false, err
	}
	return true, nil
}

func (ts *TrainServiceImpl) AllTrains(ctx context.Context) ([]TrainType, error) {
	var trains []TrainType
	coll, err := ts.db.GetCollection(ctx, "train", "train")
	if err != nil {
		return trains, err
	}
	res, err := coll.FindMany(ctx, bson.D{})
	if err != nil {
		return trains, err
	}
	err = res.All(ctx, &trains)
	if len(trains) == 0 {
		return trains, errors.New("No trains found")
	}
	return trains, err
}
