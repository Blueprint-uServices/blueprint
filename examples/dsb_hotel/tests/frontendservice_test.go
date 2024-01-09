package tests

import (
	"context"
	"testing"

	"github.com/blueprint-uservices/blueprint/examples/dsb_hotel/workflow/hotelreservation"
	"github.com/blueprint-uservices/blueprint/runtime/core/registry"
	"github.com/stretchr/testify/assert"
)

var frontendServiceRegistry = registry.NewServiceRegistry[hotelreservation.FrontEndService]("frontend_service")

func init() {
	frontendServiceRegistry.Register("local", func(ctx context.Context) (hotelreservation.FrontEndService, error) {
		searchService, err := searchServiceRegistry.Get(ctx)
		if err != nil {
			return nil, err
		}
		profileService, err := profileServiceRegistry.Get(ctx)
		if err != nil {
			return nil, err
		}
		recomdService, err := recommendationServiceRegistry.Get(ctx)
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

		return hotelreservation.NewFrontEndServiceImpl(ctx, searchService, profileService, recomdService, userService, reservationService)
	})
}

func TestSearchHandler(t *testing.T) {
	ctx := context.Background()
	service, err := frontendServiceRegistry.Get(ctx)
	assert.NoError(t, err)

	profiles, err := service.SearchHandler(ctx, "Vaastav", "2015-04-09", "2015-04-10", 37.7835, -122.41, "en")
	assert.NoError(t, err)
	assert.True(t, len(profiles) > 0)
}

func TestUserHandler(t *testing.T) {
	ctx := context.Background()
	service, err := frontendServiceRegistry.Get(ctx)
	assert.NoError(t, err)

	// Check valid user
	resp, err := service.UserHandler(ctx, "Cornell_1", "1111111111")
	assert.NoError(t, err)
	assert.Equal(t, resp, "Login successful")

	// Check Invalid pwd
	resp, err = service.UserHandler(ctx, "Cornell_1", "bleh")
	assert.Error(t, err)

	// Check invalid user
	resp, err = service.UserHandler(ctx, "Vaastav", "blueprint")
	assert.Error(t, err)
}

func TestRecommendHandler(t *testing.T) {
	ctx := context.Background()
	service, err := frontendServiceRegistry.Get(ctx)
	assert.NoError(t, err)

	profiles, err := service.RecommendHandler(ctx, 37.7835, -122.41, "dis", "en")
	assert.NoError(t, err)
	assert.True(t, len(profiles) > 0)

	profiles, err = service.RecommendHandler(ctx, 37.7835, -122.41, "rate", "en")
	assert.NoError(t, err)
	assert.True(t, len(profiles) > 0)

	profiles, err = service.RecommendHandler(ctx, 37.7835, -122.41, "price", "en")
	assert.NoError(t, err)
	assert.True(t, len(profiles) > 0)
}

func TestReservationHandler(t *testing.T) {
	ctx := context.Background()
	service, err := frontendServiceRegistry.Get(ctx)
	assert.NoError(t, err)

	status, err := service.ReservationHandler(ctx, "2015-04-09", "2015-04-10", "1", "Cornell User 1", "Cornell_1", "1111111111", 1)

	assert.NoError(t, err)
	assert.Equal(t, "Reservation successful", status)
}
