// Package basic implements ts-basic-service from original TrainTicketApplication
package basic

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"gitlab.mpi-sws.org/cld/blueprint/examples/train_ticket/workflow/common"
	"gitlab.mpi-sws.org/cld/blueprint/examples/train_ticket/workflow/price"
	"gitlab.mpi-sws.org/cld/blueprint/examples/train_ticket/workflow/route"
	"gitlab.mpi-sws.org/cld/blueprint/examples/train_ticket/workflow/station"
	"gitlab.mpi-sws.org/cld/blueprint/examples/train_ticket/workflow/train"
	"golang.org/x/exp/slices"
)

type BasicService interface {
	QueryForTravel(ctx context.Context, info common.Travel) (common.TravelResult, error)
	QueryForTravels(ctx context.Context, infos []common.Travel) (map[string]common.TravelResult, error)
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
	wg.Add(5)
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

	// Check if that stations are on the obtained route

	startIndex := -1
	endIndex := -1
	for idx, station := range route.Stations {
		if station == startstationID {
			startIndex = idx
		} else if station == endstationID {
			endIndex = idx
		}
	}

	if startIndex == -1 || endIndex == -1 {
		return res, errors.New("Selected start and/or end stations are not available on this route")
	}

	res.Prices = make(map[string]string)
	distance := route.Distances[endIndex] - route.Distances[startIndex]
	basicRate := float64(distance) * pc.BasicPriceRate
	comfortRate := float64(distance) * pc.FirstClassPriceRate
	res.Prices["economyClass"] = fmt.Sprintf("%v", basicRate)
	res.Prices["comfortClass"] = fmt.Sprintf("%v", comfortRate)

	res.TType = trainType
	res.Route = route
	res.Percent = 1.0
	res.Status = true

	return res, nil
}

func (b *BasicServiceImpl) QueryForTravels(ctx context.Context, infos []common.Travel) (map[string]common.TravelResult, error) {
	results := make(map[string]common.TravelResult)

	startTrips := make(map[string][]string)
	endTrips := make(map[string][]string)
	routeTrips := make(map[string][]string)
	typeTrips := make(map[string][]string)
	stations := make(map[string]bool)
	trainTypes := make(map[string]bool)
	routeIds := make(map[string]bool)
	avaTrips := make(map[string]bool)
	tripInfos := make(map[string]common.Travel)
	for _, info := range infos {
		stations[info.StartPlace] = true
		stations[info.EndPlace] = true
		avaTrips[info.T.ID] = true
		trainTypes[info.T.TrainTypeName] = true
		routeIds[info.T.RouteID] = true
		tripInfos[info.T.ID] = info

		if v, ok := startTrips[info.StartPlace]; !ok {
			startTrips[info.StartPlace] = []string{info.T.ID}
		} else {
			startTrips[info.StartPlace] = append(v, info.T.ID)
		}

		if v, ok := endTrips[info.EndPlace]; !ok {
			endTrips[info.EndPlace] = []string{info.T.ID}
		} else {
			endTrips[info.EndPlace] = append(v, info.T.ID)
		}

		if v, ok := routeTrips[info.T.RouteID]; !ok {
			routeTrips[info.T.RouteID] = []string{info.T.ID}
		} else {
			routeTrips[info.T.RouteID] = append(v, info.T.ID)
		}

		if v, ok := typeTrips[info.T.TrainTypeName]; !ok {
			typeTrips[info.T.TrainTypeName] = []string{info.T.ID}
		} else {
			typeTrips[info.T.TrainTypeName] = append(v, info.T.ID)
		}
	}

	var all_stations []string
	for station := range stations {
		all_stations = append(all_stations, station)
	}
	station_ids, err := b.stationService.FindIDs(ctx, all_stations)
	if err != nil {
		return results, err
	}

	for idx, sid := range station_ids {
		if sid == "" {
			// Station doesn't exist so we should remove all the trips
			station_name := all_stations[idx]
			for _, t := range startTrips[station_name] {
				delete(avaTrips, t)
			}
			for _, t := range endTrips[station_name] {
				delete(avaTrips, t)
			}
		}
	}

	if len(avaTrips) == 0 {
		return results, errors.New("No travel info available")
	}

	var all_train_type_names []string
	for tt_name := range trainTypes {
		all_train_type_names = append(all_train_type_names, tt_name)
	}
	train_types, err := b.trainService.RetrieveByNames(ctx, all_train_type_names)
	if err != nil {
		return results, err
	}

	ttypeMap := make(map[string]train.TrainType)
	for idx, tt := range train_types {
		tt_name := all_train_type_names[idx]
		if tt == (train.TrainType{}) {
			for _, t := range typeTrips[tt_name] {
				delete(avaTrips, t)
			}
		} else {
			ttypeMap[tt_name] = tt
		}
	}

	if len(avaTrips) == 0 {
		return results, errors.New("No travel info exists")
	}

	var route_ids []string
	for r := range routeIds {
		route_ids = append(route_ids, r)
	}

	routes, err := b.routeService.GetRouteByIds(ctx, route_ids)
	if err != nil {
		return results, err
	}

	routeMap := make(map[string]route.Route)
	for idx, r := range routes {
		r_id := route_ids[idx]
		if isRouteEmpty(r) {
			for _, t := range routeTrips[r_id] {
				delete(avaTrips, t)
			}
		} else {
			for _, t := range routeTrips[r_id] {
				tripInfo := tripInfos[t]
				start_index := slices.Index(r.Stations, tripInfo.StartPlace)
				end_index := slices.Index(r.Stations, tripInfo.EndPlace)
				if start_index == -1 || end_index == -1 || start_index >= end_index {
					delete(avaTrips, t)
				}
			}
			routeMap[r_id] = r
		}
	}

	if len(avaTrips) == 0 {
		return results, errors.New("No travel info exists")
	}

	var routeInfosAndTripTypes []string
	for t := range avaTrips {
		info := tripInfos[t]
		routeInfosAndTripTypes = append(routeInfosAndTripTypes, info.T.RouteID+":"+info.T.TrainTypeName)
	}

	pcs, err := b.priceService.FindByRouteIDsAndTrainTypes(ctx, routeInfosAndTripTypes)
	if err != nil {
		return results, err
	}

	for t := range avaTrips {
		var result common.TravelResult
		info := tripInfos[t]
		pc_key := info.T.RouteID + ":" + info.T.TrainTypeName
		route := routeMap[info.T.RouteID]
		ttype := ttypeMap[info.T.TrainTypeName]
		basicPriceRate := 0.75
		firstPriceRate := 1.0
		if pc, ok := pcs[pc_key]; ok {
			basicPriceRate = pc.BasicPriceRate
			firstPriceRate = pc.FirstClassPriceRate
		}
		startIndex := slices.Index(route.Stations, info.StartPlace)
		endIndex := slices.Index(route.Stations, info.EndPlace)
		distance := route.Distances[endIndex] - route.Distances[startIndex]
		priceMap := make(map[string]string)
		econPrice := basicPriceRate * float64(distance)
		firstPrice := firstPriceRate * float64(distance)
		priceMap["economyClass"] = fmt.Sprintf("%v", econPrice)
		priceMap["comfortClass"] = fmt.Sprintf("%v", firstPrice)
		result.Percent = 1.0
		result.Prices = priceMap
		result.Status = true
		result.Route = route
		result.TType = ttype
	}

	return results, nil
}

func isRouteEmpty(r route.Route) bool {
	if r.ID != "" {
		return false
	}
	if len(r.Stations) != 0 {
		return false
	}
	if len(r.Distances) != 0 {
		return false
	}
	if r.StartStation != "" {
		return false
	}
	if r.EndStation != "" {
		return false
	}

	return true
}

func (b *BasicServiceImpl) QueryForStationID(ctx context.Context, name string) (string, error) {
	return b.stationService.FindID(ctx, name)
}
