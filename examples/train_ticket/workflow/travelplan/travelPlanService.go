// Package travelplan implements ts-travel-plan service from the original Train Ticket application
package travelplan

import (
	"context"
	"errors"

	"github.com/blueprint-uservices/blueprint/examples/train_ticket/workflow/common"
	"github.com/blueprint-uservices/blueprint/examples/train_ticket/workflow/routeplan"
	"github.com/blueprint-uservices/blueprint/examples/train_ticket/workflow/seat"
	"github.com/blueprint-uservices/blueprint/examples/train_ticket/workflow/train"
	"github.com/blueprint-uservices/blueprint/examples/train_ticket/workflow/travel"
)

type TravelPlanService interface {
	GetTransferSearch(ctx context.Context, info TransferTravelInfo) (TransferTravelResult, error)
	GetCheapest(ctx context.Context, trip common.Trip) ([]TravelAdvanceResult, error)
	GetQuickest(ctx context.Context, trip common.Trip) ([]TravelAdvanceResult, error)
	GetMinStations(ctx context.Context, trip common.Trip) ([]TravelAdvanceResult, error)
	GetTravelResults(ctx context.Context, rps []routeplan.RoutePlanResultUnit) ([]TravelAdvanceResult, error)
}

type TravelPlanServiceImpl struct {
	seatService      seat.SeatService
	routePlanService routeplan.RoutePlanService
	travelService    travel.TravelService
	travel2Service   travel.TravelService
	trainService     train.TrainService
}

func NewTravelServiceImpl(ctx context.Context, seatService seat.SeatService, routePlanService routeplan.RoutePlanService, travelService travel.TravelService, travel2Service travel.TravelService, trainService train.TrainService) (*TravelPlanServiceImpl, error) {
	return &TravelPlanServiceImpl{seatService, routePlanService, travelService, travel2Service, trainService}, nil
}

func (t *TravelPlanServiceImpl) GetTransferSearch(ctx context.Context, info TransferTravelInfo) (TransferTravelResult, error) {
	var res TransferTravelResult
	highspeed_trips, err := t.travelService.QueryInfo(ctx, info.StartStation, info.EndStation, info.TravelDate)
	if err != nil {
		return res, err
	}
	normal_trips, err := t.travel2Service.QueryInfo(ctx, info.StartStation, info.EndStation, info.TravelDate)
	if err != nil {
		return res, err
	}
	res.FirstSection = append(res.FirstSection, highspeed_trips...)
	res.FirstSection = append(res.FirstSection, normal_trips...)

	highspeed_trips, err = t.travelService.QueryInfo(ctx, info.ViaStation, info.EndStation, info.TravelDate)
	if err != nil {
		return res, err
	}
	normal_trips, err = t.travel2Service.QueryInfo(ctx, info.ViaStation, info.EndStation, info.TravelDate)
	if err != nil {
		return res, err
	}
	res.SecondSection = append(res.SecondSection, highspeed_trips...)
	res.SecondSection = append(res.SecondSection, highspeed_trips...)

	return res, nil
}

func (t *TravelPlanServiceImpl) GetTravelResults(ctx context.Context, rps []routeplan.RoutePlanResultUnit) ([]TravelAdvanceResult, error) {
	var res []TravelAdvanceResult
	if len(rps) == 0 {
		return res, errors.New("No trips available for selected criteria")
	}

	for _, r := range rps {
		tr_unit := TravelAdvanceResult{}
		tr_unit.TripID = r.ID
		tr_unit.EndStation = r.EndStation
		tr_unit.StartStation = r.StartStation
		tr_unit.TrainTypeId = r.TrainTypeName
		tr_unit.StopStations = r.StopStations
		tr_unit.PriceForFirstClassSeat = r.PriceForFirstClassSeat
		tr_unit.PriceForSecondClassSeat = r.PriceForSecondClassSeat
		tr_unit.StartTime = r.StartTime
		tr_unit.EndTime = r.EndTime

		tt, err := t.trainService.RetrieveByName(ctx, r.TrainTypeName)
		if err != nil {
			return res, err
		}

		var s seat.Seat
		s.StartStation = r.StartStation
		s.DstStation = r.EndStation
		s.Stations = r.StopStations
		s.SeatType = seat.FIRSTCLASS
		s.TrainNumber = r.ID
		s.TotalNum = tt.ComfortClass
		s.TravelDate = r.StartTime
		first, err := t.seatService.GetLeftTicketOfInterval(ctx, s)
		if err != nil {
			return res, err
		}
		s.SeatType = seat.SECONDCLASS
		s.TotalNum = tt.EconomyClass
		second, err := t.seatService.GetLeftTicketOfInterval(ctx, s)
		if err != nil {
			return res, err
		}
		tr_unit.RemainingFirstClassTix = first
		tr_unit.RemainingSecondClassTix = second

		res = append(res, tr_unit)
	}

	return res, nil
}

func (t *TravelPlanServiceImpl) GetCheapest(ctx context.Context, trip common.Trip) ([]TravelAdvanceResult, error) {
	var res []TravelAdvanceResult
	var routeInfo routeplan.RoutePlanInfo
	routeInfo.StartStation = trip.StartStationName
	routeInfo.EndStation = trip.TerminalStationName
	routeInfo.TravelDate = trip.StartTime
	routeInfo.Num = 5
	rps, err := t.routePlanService.SearchCheapestResult(ctx, routeInfo)
	if err != nil {
		return res, err
	}
	return t.GetTravelResults(ctx, rps)
}

func (t *TravelPlanServiceImpl) GetQuickest(ctx context.Context, trip common.Trip) ([]TravelAdvanceResult, error) {
	var res []TravelAdvanceResult
	var routeInfo routeplan.RoutePlanInfo
	routeInfo.StartStation = trip.StartStationName
	routeInfo.EndStation = trip.TerminalStationName
	routeInfo.TravelDate = trip.StartTime
	routeInfo.Num = 5
	rps, err := t.routePlanService.SearchQuickestResult(ctx, routeInfo)
	if err != nil {
		return res, err
	}
	return t.GetTravelResults(ctx, rps)
}

func (t *TravelPlanServiceImpl) GetMinStations(ctx context.Context, trip common.Trip) ([]TravelAdvanceResult, error) {
	var res []TravelAdvanceResult
	var routeInfo routeplan.RoutePlanInfo
	routeInfo.StartStation = trip.StartStationName
	routeInfo.EndStation = trip.TerminalStationName
	routeInfo.TravelDate = trip.StartTime
	routeInfo.Num = 5
	rps, err := t.routePlanService.SearchMinStopStations(ctx, routeInfo)
	if err != nil {
		return res, err
	}
	return t.GetTravelResults(ctx, rps)
}
