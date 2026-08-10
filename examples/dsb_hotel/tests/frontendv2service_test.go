package tests

import (
	"context"
	"testing"

	"github.com/blueprint-uservices/blueprint/examples/dsb_hotel/workflow/hotelreservation"
	"github.com/blueprint-uservices/blueprint/runtime/core/registry"
	"github.com/stretchr/testify/assert"
)

var frontendV2ServiceRegistry = registry.NewServiceRegistry[hotelreservation.FrontEndV2Service]("frontend_v2_service")

func init() {
	frontendV2ServiceRegistry.Register("local", func(ctx context.Context) (hotelreservation.FrontEndV2Service, error) {
		searchService, err := searchServiceRegistry.Get(ctx)
		if err != nil {
			return nil, err
		}
		profileService, err := profileServiceRegistry.Get(ctx)
		if err != nil {
			return nil, err
		}
		recommendationService, err := recommendationServiceRegistry.Get(ctx)
		if err != nil {
			return nil, err
		}
		userService, err := userServiceRegistry.Get(ctx)
		if err != nil {
			return nil, err
		}
		reservationService, err := reservationServiceRegistry.Get(ctx)
		if err != nil {
			return nil, err
		}
		attractionsService, err := attractionsServiceRegistry.Get(ctx)
		if err != nil {
			return nil, err
		}
		reviewService, err := reviewServiceRegistry.Get(ctx)
		if err != nil {
			return nil, err
		}

		return hotelreservation.NewFrontEndV2ServiceImpl(
			ctx,
			searchService,
			profileService,
			recommendationService,
			userService,
			reservationService,
			attractionsService,
			reviewService,
		)
	})
}

func TestFrontEndV2Handlers(t *testing.T) {
	ctx := context.Background()
	service, err := frontendV2ServiceRegistry.Get(ctx)
	assert.NoError(t, err)

	reviews, err := service.ReviewHandler(ctx, "Cornell_1", "1111111111", "1")
	assert.NoError(t, err)
	assert.Len(t, reviews, 4)

	restaurants, err := service.RestaurantHandler(ctx, "Cornell_1", "1111111111", "1")
	assert.NoError(t, err)
	assert.Len(t, restaurants, 5)

	museums, err := service.MuseumHandler(ctx, "Cornell_1", "1111111111", "1")
	assert.NoError(t, err)
	assert.Len(t, museums, 1)

	cinemas, err := service.CinemaHandler(ctx, "Cornell_1", "1111111111", "1")
	assert.NoError(t, err)
	assert.Len(t, cinemas, 1)

	_, err = service.ReviewHandler(ctx, "Cornell_1", "invalid", "1")
	assert.Error(t, err)

	_, err = service.RestaurantHandler(ctx, "Cornell_1", "1111111111", "unknown")
	assert.Error(t, err)
}
