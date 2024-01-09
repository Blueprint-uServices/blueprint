// Package assurance implements the ts-assurance service from the original TrainTicket application
package assurance

import (
	"context"
	"errors"

	"github.com/blueprint-uservices/blueprint/runtime/core/backend"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
)

// AssuranceService manages assurances provided to customers for trips
type AssuranceService interface {
	// Find an assurance by ID of the assurance
	FindAssuranceById(ctx context.Context, id string) (Assurance, error)
	// Find an assurance by Order ID
	FindAssuranceByOrderId(ctx context.Context, orderId string) (Assurance, error)
	// Creates a new Assurance
	Create(ctx context.Context, typeindex int64, orderId string) (Assurance, error)
	// Deletes the assurance with ID `id`
	DeleteById(ctx context.Context, id string) error
	// Delete the assurance associated with order that has id `orderId`
	DeleteByOrderId(ctx context.Context, orderId string) error
	// Modify an existing an assurance with provided Assurance `a`
	Modify(ctx context.Context, a Assurance) (Assurance, error)
	// Return all assurances
	GetAllAssurances(ctx context.Context) ([]Assurance, error)
	// Return all types of assurances
	GetAllAssuranceTypes(ctx context.Context) ([]AssuranceType, error)
}

// Implementation of an AssuranceService
type AssuranceServiceImpl struct {
	db backend.NoSQLDatabase
}

// Constructs an AssuranceService object
func NewAssuranceServiceImpl(ctx context.Context, db backend.NoSQLDatabase) (*AssuranceServiceImpl, error) {
	return &AssuranceServiceImpl{db: db}, nil
}

func (a *AssuranceServiceImpl) GetAllAssuranceTypes(ctx context.Context) ([]AssuranceType, error) {
	return ALL_ASSURANCES, nil
}

func (a *AssuranceServiceImpl) GetAllAssurances(ctx context.Context) ([]Assurance, error) {
	coll, err := a.db.GetCollection(ctx, "assurance", "assurance")
	if err != nil {
		return []Assurance{}, err
	}
	res, err := coll.FindMany(ctx, bson.D{})
	if err != nil {
		return []Assurance{}, err
	}
	var assurances []Assurance
	err = res.All(ctx, &assurances)
	return assurances, err
}

func (a *AssuranceServiceImpl) FindAssuranceById(ctx context.Context, id string) (Assurance, error) {
	coll, err := a.db.GetCollection(ctx, "assurance", "assurance")
	if err != nil {
		return Assurance{}, err
	}
	query := bson.D{{"id", id}}
	res, err := coll.FindOne(ctx, query)
	if err != nil {
		return Assurance{}, err
	}
	var ass Assurance
	exists, err := res.One(ctx, &ass)
	if err != nil {
		return ass, err
	}
	if !exists {
		return ass, errors.New("Assurance with id " + id + " does not exist")
	}
	return ass, nil
}

func (a *AssuranceServiceImpl) FindAssuranceByOrderId(ctx context.Context, order_id string) (Assurance, error) {
	coll, err := a.db.GetCollection(ctx, "assurance", "assurance")
	if err != nil {
		return Assurance{}, err
	}
	query := bson.D{{"orderid", order_id}}
	res, err := coll.FindOne(ctx, query)
	if err != nil {
		return Assurance{}, err
	}
	var ass Assurance
	exists, err := res.One(ctx, &ass)
	if err != nil {
		return ass, err
	}
	if !exists {
		return ass, errors.New("Assurance with order id " + order_id + " does not exist")
	}
	return ass, nil
}

func (a *AssuranceServiceImpl) DeleteById(ctx context.Context, id string) error {
	coll, err := a.db.GetCollection(ctx, "assurance", "assurance")
	if err != nil {
		return err
	}
	query := bson.D{{"id", id}}
	return coll.DeleteOne(ctx, query)
}

func (a *AssuranceServiceImpl) DeleteByOrderId(ctx context.Context, order_id string) error {
	coll, err := a.db.GetCollection(ctx, "assurance", "assurance")
	if err != nil {
		return err
	}
	query := bson.D{{"orderid", order_id}}
	return coll.DeleteOne(ctx, query)
}

func (a *AssuranceServiceImpl) Modify(ctx context.Context, assurance Assurance) (Assurance, error) {
	_, err := a.FindAssuranceById(ctx, assurance.ID)
	if err != nil {
		return assurance, err
	}
	coll, err := a.db.GetCollection(ctx, "assurance", "assurance")
	if err != nil {
		return assurance, err
	}
	query := bson.D{{"id", assurance.ID}}
	ok, err := coll.Upsert(ctx, query, assurance)
	if err != nil {
		return assurance, err
	}
	if !ok {
		return assurance, errors.New("Failed to update assurance with ID " + assurance.ID)
	}
	return assurance, nil
}

func (a *AssuranceServiceImpl) Create(ctx context.Context, typeindex int64, orderid string) (Assurance, error) {
	at, err := getAssuranceType(ctx, typeindex)
	if err != nil {
		return Assurance{}, err
	}
	id := uuid.New().String()
	var assurance Assurance
	assurance.ID = id
	assurance.OrderID = orderid
	assurance.AT = at
	coll, err := a.db.GetCollection(ctx, "assurance", "assurance")
	if err != nil {
		return Assurance{}, err
	}
	return assurance, coll.InsertOne(ctx, assurance)
}
