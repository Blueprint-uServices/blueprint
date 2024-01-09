package tests

import (
	"context"
	"testing"

	"github.com/blueprint-uservices/blueprint/examples/dsb_hotel/workflow/hotelreservation"
	"github.com/blueprint-uservices/blueprint/runtime/core/registry"
	"github.com/stretchr/testify/assert"
)

var searchServiceRegistry = registry.NewServiceRegistry[hotelreservation.SearchService]("search_service")

func init() {

	searchServiceRegistry.Register("local", func(ctx context.Context) (hotelreservation.SearchService, error) {
		geoService, err := geoServiceRegistry.Get(ctx)
		if err != nil {
			return nil, err
		}

		rateService, err := rateServiceRegistry.Get(ctx)
		if err != nil {
			return nil, err
		}

		return hotelreservation.NewSearchServiceImpl(ctx, geoService, rateService)
	})
}

func TestSearchNearby(t *testing.T) {
	ctx := context.Background()
	service, err := searchServiceRegistry.Get(ctx)
	assert.NoError(t, err)

	hotels, err := service.Nearby(ctx, 37.7835, -122.41, "2015-04-09", "2015-04-10")
	assert.NoError(t, err)
	assert.True(t, len(hotels) > 0)
}
