// Package routeplan implements ts-route-plan-service from the original TrainTicket application
package routeplan

import (
	"context"
	"math"
	"sort"
	"time"

	"github.com/blueprint-uservices/blueprint/examples/train_ticket/workflow/route"
	"github.com/blueprint-uservices/blueprint/examples/train_ticket/workflow/travel"
	"golang.org/x/exp/slices"
)

type RoutePlanService interface {
	SearchCheapestResult(ctx context.Context, info RoutePlanInfo) ([]RoutePlanResultUnit, error)
	SearchQuickestResult(ctx context.Context, info RoutePlanInfo) ([]RoutePlanResultUnit, error)
	SearchMinStopStations(ctx context.Context, info RoutePlanInfo) ([]RoutePlanResultUnit, error)
}

type RoutePlanServiceImpl struct {
	travelService  travel.TravelService
	travel2Service travel.TravelService
	routeService   route.RouteService
}

func NewRoutePlanServiceImpl(ctx context.Context, travelService travel.TravelService, travel2Service travel.TravelService, routeService route.RouteService) (*RoutePlanServiceImpl, error) {
	return &RoutePlanServiceImpl{travelService: travelService, travel2Service: travel2Service, routeService: routeService}, nil
}

func (r *RoutePlanServiceImpl) SearchCheapestResult(ctx context.Context, info RoutePlanInfo) ([]RoutePlanResultUnit, error) {
	var response []RoutePlanResultUnit
	var all_trips []travel.TripResponse
	trips1, err := r.travelService.QueryInfo(ctx, info.StartStation, info.EndStation, info.TravelDate)
	if err != nil {
		return response, err
	}
	trips2, err := r.travel2Service.QueryInfo(ctx, info.StartStation, info.EndStation, info.TravelDate)
	if err != nil {
		return response, err
	}
	all_trips = append(all_trips, trips1...)
	all_trips = append(all_trips, trips2...)
	sort.Slice(all_trips, func(i, j int) bool {
		return all_trips[i].PriceForEconomyClass < all_trips[j].PriceForEconomyClass
	})
	result_size := int(math.Min(float64(len(all_trips)), 5))
	for i := 0; i < result_size; i++ {
		var result RoutePlanResultUnit
		result.ID = all_trips[i].TripId
		result.TrainTypeName = all_trips[i].TrainTypeId
		result.StartStation = all_trips[i].StartingStation
		result.EndStation = all_trips[i].EndStation
		result.StopStations = all_trips[i].StopStations
		result.PriceForFirstClassSeat = all_trips[i].PriceForComfortClass
		result.PriceForSecondClassSeat = all_trips[i].PriceForEconomyClass
		result.StartTime = all_trips[i].StartingTime
		result.EndTime = all_trips[i].EndTime
		response = append(response, result)
	}
	return response, nil
}

func (r *RoutePlanServiceImpl) SearchQuickestResult(ctx context.Context, info RoutePlanInfo) ([]RoutePlanResultUnit, error) {
	var response []RoutePlanResultUnit
	var all_trips []travel.TripResponse
	trips1, err := r.travelService.QueryInfo(ctx, info.StartStation, info.EndStation, info.TravelDate)
	if err != nil {
		return response, err
	}
	trips2, err := r.travel2Service.QueryInfo(ctx, info.StartStation, info.EndStation, info.TravelDate)
	if err != nil {
		return response, err
	}
	all_trips = append(all_trips, trips1...)
	all_trips = append(all_trips, trips2...)
	sort.Slice(all_trips, func(i, j int) bool {
		dur1, _ := time.ParseDuration(all_trips[i].Duration)
		dur2, _ := time.ParseDuration(all_trips[j].Duration)
		return dur1 < dur2
	})
	result_size := int(math.Min(float64(len(all_trips)), 5))
	for i := 0; i < result_size; i++ {
		var result RoutePlanResultUnit
		result.ID = all_trips[i].TripId
		result.TrainTypeName = all_trips[i].TrainTypeId
		result.StartStation = all_trips[i].StartingStation
		result.EndStation = all_trips[i].EndStation
		result.StopStations = all_trips[i].StopStations
		result.PriceForFirstClassSeat = all_trips[i].PriceForComfortClass
		result.PriceForSecondClassSeat = all_trips[i].PriceForEconomyClass
		result.StartTime = all_trips[i].StartingTime
		result.EndTime = all_trips[i].EndTime
		response = append(response, result)
	}
	return response, nil
}

func (r *RoutePlanServiceImpl) SearchMinStopStations(ctx context.Context, info RoutePlanInfo) ([]RoutePlanResultUnit, error) {
	var response []RoutePlanResultUnit
	all_routes, err := r.routeService.GetAllRoutes(ctx)
	if err != nil {
		return response, err
	}
	lowest := math.MaxInt32
	var lowest_route route.Route
	for _, r := range all_routes {
		start_index := slices.Index(r.Stations, info.StartStation)
		end_index := slices.Index(r.Stations, info.EndStation)
		if start_index == -1 || end_index == -1 || end_index <= start_index {
			continue
		}
		numStops := end_index - start_index
		if numStops < lowest {
			lowest = numStops
			lowest_route = r
		}
	}
	trips1, err := r.travelService.GetTripsByRouteId(ctx, []string{lowest_route.ID})
	if err != nil {
		return response, err
	}
	trips2, err := r.travel2Service.GetTripsByRouteId(ctx, []string{lowest_route.ID})
	all_trips := append(trips1, trips2...)
	result_size := int(math.Min(float64(len(all_trips)), 5))
	for i := 0; i < result_size; i++ {
		var result RoutePlanResultUnit
		result.ID = all_trips[i].ID
		result.TrainTypeName = all_trips[i].TrainTypeName
		result.StartStation = all_trips[i].StartStationName
		result.EndStation = all_trips[i].TerminalStationName
		result.StopStations = lowest_route.Stations
		var resp travel.TripResponse
		if all_trips[i].ID[0:1] == "G" || all_trips[i].ID[0:1] == "D" {
			_, resp, err = r.travelService.GetTripAllDetailInfo(ctx, all_trips[i].ID, all_trips[i].StartStationName, all_trips[i].TerminalStationName, all_trips[i].StartTime)
		} else {
			_, resp, err = r.travel2Service.GetTripAllDetailInfo(ctx, all_trips[i].ID, all_trips[i].StartStationName, all_trips[i].TerminalStationName, all_trips[i].StartTime)
		}
		result.PriceForFirstClassSeat = resp.PriceForComfortClass
		result.PriceForSecondClassSeat = resp.PriceForEconomyClass
		result.StartTime = all_trips[i].StartTime
		result.EndTime = all_trips[i].EndTime
		response = append(response, result)
	}
	return response, nil
}
