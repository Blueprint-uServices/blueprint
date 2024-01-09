// package route implements ts-route-service from the original train ticket application
package route

import (
	"context"
	"errors"
	"strconv"
	"strings"

	"github.com/blueprint-uservices/blueprint/runtime/core/backend"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
)

// RouteService manages all the routes in the application
type RouteService interface {
	// Get a route based on the `start` point and `end` point
	GetRouteByStartAndEnd(ctx context.Context, start string, end string) (Route, error)
	// Gets all routes
	GetAllRoutes(ctx context.Context) ([]Route, error)
	// Get a route by ID
	GetRouteById(ctx context.Context, id string) (Route, error)
	// Get multiple routes based on ids
	GetRouteByIds(ctx context.Context, ids []string) ([]Route, error)
	// Delete a route by `id`
	DeleteRoute(ctx context.Context, id string) error
	// Create a new route or modify an existing route based on provided `info` for the route
	CreateAndModify(ctx context.Context, info RouteInfo) (Route, error)
}

type RouteServiceImpl struct {
	db backend.NoSQLDatabase
}

func NewRouteServiceImpl(ctx context.Context, db backend.NoSQLDatabase) (*RouteServiceImpl, error) {
	return &RouteServiceImpl{db: db}, nil
}

func (r *RouteServiceImpl) DeleteRoute(ctx context.Context, id string) error {
	coll, err := r.db.GetCollection(ctx, "route", "route")
	if err != nil {
		return err
	}
	return coll.DeleteOne(ctx, bson.D{{"id", id}})
}

func (r *RouteServiceImpl) GetAllRoutes(ctx context.Context) ([]Route, error) {
	var routes []Route
	var err error
	coll, err := r.db.GetCollection(ctx, "route", "route")
	if err != nil {
		return routes, err
	}
	res, err := coll.FindMany(ctx, bson.D{})
	if err != nil {
		return routes, err
	}
	err = res.All(ctx, &routes)
	if len(routes) == 0 {
		return routes, errors.New("No content found")
	}
	return routes, err
}

func (r *RouteServiceImpl) GetRouteById(ctx context.Context, id string) (Route, error) {
	coll, err := r.db.GetCollection(ctx, "route", "route")
	if err != nil {
		return Route{}, err
	}
	res, err := coll.FindOne(ctx, bson.D{{"id", id}})
	if err != nil {
		return Route{}, err
	}
	var route Route
	exists, err := res.One(ctx, &route)
	if err != nil {
		return Route{}, err
	}
	if !exists {
		return Route{}, errors.New("Route with ID " + id + " does not exist")
	}
	return route, nil
}

func (r *RouteServiceImpl) GetRouteByIds(ctx context.Context, ids []string) ([]Route, error) {
	var routes []Route
	for _, id := range ids {
		route, err := r.GetRouteById(ctx, id)
		if err == nil {
			routes = append(routes, route)
		} else {
			routes = append(routes, Route{})
		}
	}
	return routes, nil
}

func (r *RouteServiceImpl) GetRouteByStartAndEnd(ctx context.Context, start string, end string) (Route, error) {
	coll, err := r.db.GetCollection(ctx, "route", "route")
	if err != nil {
		return Route{}, err
	}
	query := bson.D{{"$and", bson.A{
		bson.D{{"startstation", start}},
		bson.D{{"endstation", end}},
	}}}
	res, err := coll.FindOne(ctx, query)
	if err != nil {
		return Route{}, err
	}
	var route Route
	exists, err := res.One(ctx, &route)
	if err != nil {
		return route, err
	}
	if !exists {
		return route, errors.New("Route with start station " + start + " and end station " + end + " does not exist.")
	}
	return route, nil
}

func (r *RouteServiceImpl) CreateAndModify(ctx context.Context, info RouteInfo) (Route, error) {
	coll, err := r.db.GetCollection(ctx, "route", "route")
	if err != nil {
		return Route{}, err
	}
	var distances []int64
	stations := strings.Split(info.StationList, ",")
	dist_pieces := strings.Split(info.DistanceList, ",")
	for _, piece := range dist_pieces {
		converted_distance, err := strconv.ParseInt(piece, 10, 64)
		if err != nil {
			return Route{}, err
		}
		distances = append(distances, converted_distance)
	}

	if len(stations) != len(distances) {
		return Route{}, errors.New("Length of stations and distances do not match")
	}
	route := Route{}
	const MAXIDLEN = 32
	old_exists := false
	if info.ID == "" || len(info.ID) < MAXIDLEN {
		route.ID = uuid.New().String()
	} else {
		res, err := coll.FindOne(ctx, bson.D{{"id", info.ID}})
		if err != nil {
			return route, err
		}
		var old_route Route
		exists, err := res.One(ctx, &old_route)
		if err != nil {
			return route, err
		}
		if exists {
			route.ID = old_route.ID
		} else {
			route.ID = info.ID
		}
		old_exists = exists
	}
	route.Stations = stations
	route.Distances = distances
	route.StartStation = info.StartStation
	route.EndStation = info.EndStation
	if old_exists {
		ok, err := coll.Upsert(ctx, bson.D{{"id", info.ID}}, route)
		if err != nil {
			return route, err
		}
		if !ok {
			return route, errors.New("Failed to update route")
		}
		return route, nil
	}

	return route, coll.InsertOne(ctx, route)
}
