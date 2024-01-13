// Package travel implements ts-travel and ts-travel2 services from the original TrainTicket application
package travel

import (
	"context"

	"github.com/blueprint-uservices/blueprint/examples/train_ticket/workflow/common"
	"github.com/blueprint-uservices/blueprint/examples/train_ticket/workflow/route"
	"github.com/blueprint-uservices/blueprint/examples/train_ticket/workflow/train"
	"github.com/blueprint-uservices/blueprint/runtime/core/backend"
)

type TravelService interface {
	GetTrainTypeByTripId(ctx context.Context, tripId string) (train.TrainType, error)
	GetRouteByTripId(ctx context.Context, tripId string) (route.Route, error)
	GetTripsByRouteId(ctx context.Context, routeIds []string) ([]common.Trip, error)
	UpdateTrip(ctx context.Context, trip common.Trip) (common.Trip, error)
	Retrieve(ctx context.Context, tripId string) (common.Trip, error)
	CreateTrip(ctx context.Context, trip common.Trip) (common.Trip, error)
	DeleteTrip(ctx context.Context, tripId string) (string, error)
	QueryInfo(ctx context.Context, startingPlace string, endPlace string, departureTime string) ([]TripResponse, error)
	GetTripAllDetailInfo(ctx context.Context, id string, from string, to string, travelDate string) (common.Trip, TripResponse, error)
	GetTickets(ctx context.Context, id string, from string, to string, travelDate string) (TripResponse, error)
	QueryAll(ctx context.Context) ([]common.Trip, error)
	AdminQueryAll(ctx context.Context) ([]common.Trip, []train.TrainType, []route.Route, error)
}

type TravelServiceImpl struct {
	db backend.NoSQLDatabase
}
