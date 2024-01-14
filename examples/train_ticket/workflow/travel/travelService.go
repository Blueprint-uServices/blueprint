// Package travel implements ts-travel and ts-travel2 services from the original TrainTicket application
package travel

import (
	"context"
	"errors"
	"strconv"
	"time"

	"github.com/blueprint-uservices/blueprint/examples/train_ticket/workflow/admin"
	"github.com/blueprint-uservices/blueprint/examples/train_ticket/workflow/basic"
	"github.com/blueprint-uservices/blueprint/examples/train_ticket/workflow/common"
	"github.com/blueprint-uservices/blueprint/examples/train_ticket/workflow/route"
	"github.com/blueprint-uservices/blueprint/examples/train_ticket/workflow/seat"
	"github.com/blueprint-uservices/blueprint/examples/train_ticket/workflow/train"
	"github.com/blueprint-uservices/blueprint/runtime/core/backend"
	"go.mongodb.org/mongo-driver/bson"
	"golang.org/x/exp/slices"
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
	QueryAll(ctx context.Context) ([]common.Trip, error)
	AdminQueryAll(ctx context.Context) ([]admin.AdminTrip, error)
}

type TravelServiceImpl struct {
	db           backend.NoSQLDatabase
	basicService basic.BasicService
	trainService train.TrainService
	routeService route.RouteService
	seatService  seat.SeatService
}

func NewTravelServiceImpl(ctx context.Context, db backend.NoSQLDatabase, basicService basic.BasicService, trainService train.TrainService, routeService route.RouteService, seatService seat.SeatService) (*TravelServiceImpl, error) {
	return &TravelServiceImpl{db: db, basicService: basicService, trainService: trainService, routeService: routeService, seatService: seatService}, nil
}

func (t *TravelServiceImpl) Retrieve(ctx context.Context, tripId string) (common.Trip, error) {
	coll, err := t.db.GetCollection(ctx, "trips", "trips")
	if err != nil {
		return common.Trip{}, err
	}
	query := bson.D{{"id", tripId}}
	res, err := coll.FindOne(ctx, query)
	if err != nil {
		return common.Trip{}, err
	}
	var trip common.Trip
	ok, err := res.One(ctx, &trip)
	if err != nil {
		return common.Trip{}, err
	}
	if !ok {
		return common.Trip{}, errors.New("Trip with ID " + tripId + " not found")
	}
	return trip, nil
}

func (t *TravelServiceImpl) GetTrainTypeByTripId(ctx context.Context, tripId string) (train.TrainType, error) {
	var tt train.TrainType
	trip, err := t.Retrieve(ctx, tripId)
	if err != nil {
		return tt, err
	}
	return t.trainService.RetrieveByName(ctx, trip.TrainTypeName)
}

func (t *TravelServiceImpl) GetRouteByTripId(ctx context.Context, tripId string) (route.Route, error) {
	var r route.Route
	trip, err := t.Retrieve(ctx, tripId)
	if err != nil {
		return r, err
	}
	return t.routeService.GetRouteById(ctx, trip.RouteID)
}

func (t *TravelServiceImpl) GetTripsByRouteId(ctx context.Context, routeIds []string) ([]common.Trip, error) {
	coll, err := t.db.GetCollection(ctx, "trips", "trips")
	if err != nil {
		return []common.Trip{}, err
	}
	var all_trips []common.Trip

	for _, rid := range routeIds {
		var specific_trips []common.Trip
		query := bson.D{{"routeid", rid}}
		res, err := coll.FindMany(ctx, query)
		if err != nil {
			continue
		}
		err = res.All(ctx, &specific_trips)
		if err != nil {
			continue
		}
		all_trips = append(all_trips, specific_trips...)
	}
	return all_trips, nil
}

func (t *TravelServiceImpl) UpdateTrip(ctx context.Context, trip common.Trip) (common.Trip, error) {
	coll, err := t.db.GetCollection(ctx, "trips", "trips")
	if err != nil {
		return common.Trip{}, err
	}
	query := bson.D{{"id", trip.ID}}
	ok, err := coll.Upsert(ctx, query, trip)
	if err != nil {
		return common.Trip{}, err
	}
	if !ok {
		return common.Trip{}, errors.New("Failed to update trip")
	}
	return trip, nil
}

func (t *TravelServiceImpl) CreateTrip(ctx context.Context, trip common.Trip) (common.Trip, error) {
	coll, err := t.db.GetCollection(ctx, "trips", "trips")
	if err != nil {
		return common.Trip{}, err
	}
	err = coll.InsertOne(ctx, trip)
	return trip, err
}

func (t *TravelServiceImpl) DeleteTrip(ctx context.Context, tripId string) (string, error) {
	coll, err := t.db.GetCollection(ctx, "trips", "trips")
	if err != nil {
		return "", err
	}
	query := bson.D{{"id", tripId}}
	err = coll.DeleteOne(ctx, query)
	if err != nil {
		return "", err
	}
	return "Deletion successful", nil
}

func (t *TravelServiceImpl) getTickets(ctx context.Context, trip common.Trip, start string, end string, departureTime string) (TripResponse, error) {
	tr := common.Travel{}
	tr.StartPlace = start
	tr.EndPlace = end
	tr.DepartureTime = departureTime
	tr.T = trip
	res, err := t.basicService.QueryForTravel(ctx, tr)
	if err != nil {
		return TripResponse{}, err
	}
	tr_resp := TripResponse{}
	tr_resp.TripId = trip.ID
	tr_resp.StartingStation = start
	tr_resp.EndStation = end
	tr_resp.TrainTypeId = res.TType.ID
	price_com, _ := strconv.ParseFloat(res.Prices["ComfortClass"], 64)
	tr_resp.PriceForComfortClass = price_com
	price_econ, _ := strconv.ParseFloat(res.Prices["EconomyClass"], 64)
	tr_resp.PriceForEconomyClass = price_econ
	tr_resp.StartingTime = departureTime

	start_index := slices.Index(res.Route.Stations, start)
	end_index := slices.Index(res.Route.Stations, end)
	distance := res.Route.Distances[end_index] - res.Route.Distances[start_index]
	// Calculated in hours
	time_taken := distance / res.TType.AvgSpeed
	dateFormat := "Sat Jul 26 00:00:00 2025"

	depart, _ := time.Parse(dateFormat, departureTime)
	end_time := depart.Add(time.Duration(time_taken) * time.Hour)
	tr_resp.EndTime = end_time.Format(dateFormat)

	s := seat.Seat{}
	s.DstStation = end
	s.StartStation = start
	s.Stations = res.Route.Stations
	s.TrainNumber = trip.ID
	s.TravelDate = departureTime
	s.TotalNum = res.TType.EconomyClass
	s.SeatType = seat.SECONDCLASS
	val, err := t.seatService.GetLeftTicketOfInterval(ctx, s)
	if err != nil {
		return tr_resp, err
	}
	tr_resp.EconomyClass = val
	s.SeatType = seat.FIRSTCLASS
	val, err = t.seatService.GetLeftTicketOfInterval(ctx, s)
	if err != nil {
		return tr_resp, err
	}
	tr_resp.ComfortClass = val
	return tr_resp, nil
}

func (tsi *TravelServiceImpl) QueryInfo(ctx context.Context, startingPlace string, endPlace string, departureTime string) ([]TripResponse, error) {
	var response []TripResponse

	collection, err := tsi.db.GetCollection(ctx, "trips", "trips")
	if err != nil {
		return response, err
	}

	result, err := collection.FindMany(ctx, bson.D{})
	if err != nil {
		return response, err
	}
	var trips []common.Trip
	err = result.All(ctx, &trips)
	if err != nil {
		return response, err
	}

	for _, trip := range trips {
		route, err := tsi.routeService.GetRouteById(ctx, trip.RouteID)
		if err != nil {
			continue
		}

		foundLeft := false
		for _, station := range route.Stations {
			if startingPlace == station {
				foundLeft = true
				continue
			}

			if endPlace == station {
				if foundLeft {
					tripDetails, err := tsi.getTickets(ctx, trip, startingPlace, endPlace, departureTime)
					if err != nil {
						break
					}
					response = append(response, tripDetails)
				} else {
					break
				}
			}
		}

	}
	return response, nil
}

func (tsi *TravelServiceImpl) QueryAll(ctx context.Context) ([]common.Trip, error) {
	var response []common.Trip

	collection, err := tsi.db.GetCollection(ctx, "trips", "trips")
	if err != nil {
		return response, err
	}
	result, err := collection.FindMany(ctx, bson.D{})
	if err != nil {
		return nil, err
	}
	err = result.All(ctx, &response)
	if err != nil {
		return response, err
	}
	return response, nil
}

func (t *TravelServiceImpl) AdminQueryAll(ctx context.Context) ([]admin.AdminTrip, error) {
	var admin_trips []admin.AdminTrip
	all_trips, err := t.QueryAll(ctx)
	if err != nil {
		return admin_trips, err
	}
	for _, trip := range all_trips {
		route, err := t.routeService.GetRouteById(ctx, trip.RouteID)
		if err != nil {
			continue
		}
		ttype, err := t.trainService.RetrieveByName(ctx, trip.TrainTypeName)
		if err != nil {
			continue
		}
		admin_trips = append(admin_trips, admin.AdminTrip{T: trip, R: route, TT: ttype})
	}
	return admin_trips, nil
}

func (t *TravelServiceImpl) GetTripAllDetailInfo(ctx context.Context, id string, from string, to string, departureTime string) (common.Trip, TripResponse, error) {
	trip, err := t.Retrieve(ctx, id)
	if err != nil {
		return trip, TripResponse{}, err
	}
	resp, err := t.getTickets(ctx, trip, from, to, departureTime)
	return trip, resp, err
}
