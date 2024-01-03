// Package basic implements ts-basic-service from original TrainTicketApplication
package basic

import (
	"context"
	"sync"

	"gitlab.mpi-sws.org/cld/blueprint/examples/train_ticket/workflow/common"
	"gitlab.mpi-sws.org/cld/blueprint/examples/train_ticket/workflow/price"
	"gitlab.mpi-sws.org/cld/blueprint/examples/train_ticket/workflow/route"
	"gitlab.mpi-sws.org/cld/blueprint/examples/train_ticket/workflow/station"
	"gitlab.mpi-sws.org/cld/blueprint/examples/train_ticket/workflow/train"
)

type BasicService interface {
	QueryForTravel(ctx context.Context, info common.Travel) (common.TravelResult, error)
	QueryForTravels(ctx context.Context, infos []common.Travel) ([]common.TravelResult, error)
	QueryForStationID(ctx context.Context, name string) (string, error)
}

type BasicServiceImpl struct {
	trainService   train.TrainService
	stationService station.StationService
	routeService   route.RouteService
	priceService   price.PriceService
}

func NewBasicServiceImpl(ctx context.Context, trainService train.TrainService, stationService station.StationService, routeService route.RouteService, priceService price.PriceService) (*BasicServiceImpl, error) {
	return &BasicServiceImpl{trainService: trainService, stationService: stationService, routeService: routeService, priceService: priceService}, nil
}

func (b *BasicServiceImpl) QueryForTravel(ctx context.Context, info common.Travel) (common.TravelResult, error) {
	var res common.TravelResult
	var wg sync.WaitGroup
	wg.Add(4)
	var trainType train.TrainType
	var startstationID, endstationID string
	var route route.Route
	var pc price.PriceConfig
	var err1, err2, err3, err4, err5 error
	go func() {
		defer wg.Done()
		trainType, err1 = b.trainService.Retrieve(ctx, info.T.TrainTypeName)
	}()
	go func() {
		defer wg.Done()
		startstationID, err2 = b.stationService.FindID(ctx, info.StartPlace)
	}()
	go func() {
		defer wg.Done()
		endstationID, err3 = b.stationService.FindID(ctx, info.EndPlace)
	}()
	go func() {
		defer wg.Done()
		route, err4 = b.routeService.GetRouteById(ctx, info.T.RouteID)
	}()
	go func() {
		defer wg.Done()
		pc, err5 = b.priceService.FindByRouteIDAndTrainType(ctx, info.T.RouteID, info.T.TrainTypeName)
	}()

	wg.Wait()
	if err1 != nil {
		return res, err1
	}
	if err2 != nil {
		return res, err2
	}
	if err3 != nil {
		return res, err3
	}
	if err4 != nil {
		return res, err4
	}
	if err5 != nil {
		return res, err5
	}

	return res, nil
}

func (b *BasicServiceImpl) QueryForTravels(ctx context.Context, infos []common.Travel) ([]common.TravelResult, error) {
	var results []common.TravelResult
	//

	return results, nil
}

func (b *BasicServiceImpl) QueryForStationID(ctx context.Context, name string) (string, error) {
	return b.stationService.FindID(ctx, name)
}
