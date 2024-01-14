package admin

import (
	"context"
	"errors"
	"sync"

	"github.com/blueprint-uservices/blueprint/examples/train_ticket/workflow/common"
	"github.com/blueprint-uservices/blueprint/examples/train_ticket/workflow/travel"
)

type AdminTravelService interface {
	GetAllTravels(ctx context.Context) ([]common.AdminTrip, error)
	AddTravel(ctx context.Context, trip common.Trip) (common.Trip, error)
	UpdateTravel(ctx context.Context, trip common.Trip) (common.Trip, error)
	DeleteTravel(ctx context.Context, tripId string) (string, error)
}

type AdminTravelServiceImpl struct {
	travelService  travel.TravelService
	travel2Service travel.TravelService
}

func NewAdminTravelServiceImpl(ctx context.Context, travelService travel.TravelService, travel2Service travel.TravelService) (*AdminTravelServiceImpl, error) {
	return &AdminTravelServiceImpl{travelService: travelService, travel2Service: travel2Service}, nil
}

func (atsi *AdminTravelServiceImpl) GetAllTravels(ctx context.Context) ([]common.AdminTrip, error) {
	var err1, err2 error
	var trips1, trips2 []common.AdminTrip
	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		trips1, err1 = atsi.travelService.AdminQueryAll(ctx)
	}()

	go func() {
		defer wg.Done()
		trips2, err2 = atsi.travel2Service.AdminQueryAll(ctx)
	}()

	wg.Wait()

	if err1 != nil {
		return trips1, err1
	}
	if err2 != nil {
		return trips2, err2
	}

	trips := append(trips1, trips2...)

	if len(trips) == 0 {
		return trips, errors.New("No trips found")
	}

	return trips, nil
}

func (atsi *AdminTravelServiceImpl) AddTravel(ctx context.Context, trip common.Trip) (common.Trip, error) {
	var err error
	if trip.ID[0:1] == "D" || trip.ID[0:1] == "G" {
		trip, err = atsi.travelService.CreateTrip(ctx, trip)
	} else {
		trip, err = atsi.travel2Service.CreateTrip(ctx, trip)
	}

	if err != nil {
		return common.Trip{}, err
	}

	return trip, nil
}

func (atsi *AdminTravelServiceImpl) UpdateTravel(ctx context.Context, trip common.Trip) (common.Trip, error) {
	var err error
	if trip.ID[0:1] == "D" || trip.ID[0:1] == "G" {
		trip, err = atsi.travelService.UpdateTrip(ctx, trip)
	} else {
		trip, err = atsi.travel2Service.UpdateTrip(ctx, trip)
	}

	if err != nil {
		return common.Trip{}, err
	}

	return trip, nil
}

func (atsi *AdminTravelServiceImpl) DeleteRoute(ctx context.Context, tripId string) (string, error) {
	var err error
	if tripId[0:1] == "D" || tripId[0:1] == "G" {
		_, err = atsi.travelService.DeleteTrip(ctx, tripId)
	} else {
		_, err = atsi.travel2Service.DeleteTrip(ctx, tripId)
	}
	if err != nil {
		return "", err
	}

	return "Trip deleted.", nil
}
