// package station implements ts-station-service from the original TrainTicket application
package station

import (
	"context"
	"errors"

	"github.com/blueprint-uservices/blueprint/runtime/core/backend"
	"go.mongodb.org/mongo-driver/bson"
)

// StationService manages all stations
type StationService interface {
	// Creates a new station
	CreateStation(ctx context.Context, station Station) error
	// Check if a station exists
	Exists(ctx context.Context, name string) (bool, error)
	// Updates an existing station
	UpdateStation(ctx context.Context, station Station) (bool, error)
	// Deletes an existing station based on `id`
	DeleteStation(ctx context.Context, id string) error
	// Find a station based on `id`
	FindByID(ctx context.Context, id string) (Station, error)
	// Find all stations based on `ids`
	FindByIDs(ctx context.Context, ids []string) ([]Station, error)
	// Find the station `id` for the station with Name `name`
	FindID(ctx context.Context, name string) (string, error)
	// Find the station `ids` for stations with Names `names`
	FindIDs(ctx context.Context, names []string) ([]string, error)
}

// Implementation of the StationService
type StationServiceImpl struct {
	stationDB backend.NoSQLDatabase
}

// Returns a new StationService object
func NewStationServiceImpl(ctx context.Context, db backend.NoSQLDatabase) (*StationServiceImpl, error) {
	return &StationServiceImpl{stationDB: db}, nil
}

func (s *StationServiceImpl) CreateStation(ctx context.Context, station Station) error {
	coll, err := s.stationDB.GetCollection(ctx, "station", "station")
	if err != nil {
		return err
	}
	query := bson.D{{"id", station.ID}}
	res, err := coll.FindOne(ctx, query)
	if err != nil {
		return err
	}
	var st Station
	exists, err := res.One(ctx, &st)
	if err != nil {
		return err
	}
	if exists {
		return errors.New("Station with station id " + station.ID + " already exists")
	}

	return coll.InsertOne(ctx, station)
}

func (s *StationServiceImpl) Exists(ctx context.Context, name string) (bool, error) {
	coll, err := s.stationDB.GetCollection(ctx, "station", "station")
	if err != nil {
		return false, err
	}
	query := bson.D{{"name", name}}
	res, err := coll.FindOne(ctx, query)
	if err != nil {
		return false, err
	}
	var st Station
	exists, err := res.One(ctx, &st)
	return exists, err
}

func (s *StationServiceImpl) UpdateStation(ctx context.Context, station Station) (bool, error) {
	coll, err := s.stationDB.GetCollection(ctx, "station", "station")
	if err != nil {
		return false, err
	}
	query := bson.D{{"id", station.ID}}
	return coll.Upsert(ctx, query, station)
}

func (s *StationServiceImpl) DeleteStation(ctx context.Context, id string) error {
	coll, err := s.stationDB.GetCollection(ctx, "station", "station")
	if err != nil {
		return err
	}
	query := bson.D{{"id", id}}
	return coll.DeleteOne(ctx, query)
}

func (s *StationServiceImpl) FindID(ctx context.Context, name string) (string, error) {
	coll, err := s.stationDB.GetCollection(ctx, "station", "station")
	if err != nil {
		return "", err
	}
	query := bson.D{{"name", name}}
	res, err := coll.FindOne(ctx, query)
	if err != nil {
		return "", err
	}
	var st Station
	exists, err := res.One(ctx, &st)
	if err != nil {
		return "", err
	}
	if !exists {
		return "", errors.New("Station with name " + name + "does not exist")
	}
	return st.ID, nil
}

func (s *StationServiceImpl) FindIDs(ctx context.Context, names []string) ([]string, error) {
	var ids []string
	for _, name := range names {
		id, err := s.FindID(ctx, name)
		if err != nil {
			// Attach an empty string to indicate that the ID was not found for a given station
			ids = append(ids, "")
		}
		ids = append(ids, id)
	}
	return ids, nil
}

func (s *StationServiceImpl) FindByID(ctx context.Context, id string) (Station, error) {
	coll, err := s.stationDB.GetCollection(ctx, "station", "station")
	if err != nil {
		return Station{}, err
	}
	query := bson.D{{"id", id}}
	res, err := coll.FindOne(ctx, query)
	if err != nil {
		return Station{}, err
	}
	var st Station
	exists, err := res.One(ctx, &st)
	if err != nil {
		return Station{}, err
	}
	if !exists {
		return Station{}, errors.New("Station with id " + id + "does not exist")
	}
	return st, nil
}

func (s *StationServiceImpl) FindByIDs(ctx context.Context, ids []string) ([]Station, error) {
	var stations []Station
	for _, id := range ids {
		st, err := s.FindByID(ctx, id)
		if err != nil {
			// Attach an empty Station object to indicate that the ID was not found for a given station
			stations = append(stations, Station{})
		}
		stations = append(stations, st)
	}
	return stations, nil
}
