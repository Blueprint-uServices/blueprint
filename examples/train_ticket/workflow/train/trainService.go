package train

import (
	"context"
	"errors"

	"github.com/blueprint-uservices/blueprint/runtime/core/backend"
	"go.mongodb.org/mongo-driver/bson"
)

// TrainService manages the different types of trains in the application
type TrainService interface {
	// Creates a new type of train
	Create(ctx context.Context, ttype TrainType) (bool, error)
	// Retrieves the type of train using its `id`
	Retrieve(ctx context.Context, id string) (TrainType, error)
	// Retrieves the type of train using its `name`
	RetrieveByName(ctx context.Context, name string) (TrainType, error)
	// Retrieves all train types using its `names`
	RetrieveByNames(ctx context.Context, names []string) ([]TrainType, error)
	// Updates an existing train type
	Update(ctx context.Context, ttype TrainType) (bool, error)
	// Delete an existing train type
	Delete(ctx context.Context, id string) (bool, error)
	// Returns all types of trains
	AllTrains(ctx context.Context) ([]TrainType, error)
}

// Implementation of TrainService
type TrainServiceImpl struct {
	db backend.NoSQLDatabase
}

// Creates a new TrainService object
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
