package tests

import (
	"context"
	"testing"

	"github.com/blueprint-uservices/blueprint/examples/dsb_hotel/workflow/hotelreservation"
	"github.com/blueprint-uservices/blueprint/runtime/core/registry"
	"github.com/blueprint-uservices/blueprint/runtime/plugins/simplenosqldb"
	"github.com/stretchr/testify/assert"
)

var recommendationServiceRegistry = registry.NewServiceRegistry[hotelreservation.RecommendationService]("recommendation_service")

func init() {

	recommendationServiceRegistry.Register("local", func(ctx context.Context) (hotelreservation.RecommendationService, error) {
		db, err := simplenosqldb.NewSimpleNoSQLDB(ctx)
		if err != nil {
			return nil, err
		}

		return hotelreservation.NewRecommendationServiceImpl(ctx, db)
	})
}

func TestGetRecommendations(t *testing.T) {
	ctx := context.Background()
	service, err := recommendationServiceRegistry.Get(ctx)
	assert.NoError(t, err)

	// Check require = "dis"
	dis_hotels, err := service.GetRecommendations(ctx, "dis", 37.7835, -122.41)
	assert.NoError(t, err)
	assert.True(t, len(dis_hotels) > 0)
	// Check require = "rate"
	rate_hotels, err := service.GetRecommendations(ctx, "rate", 37.7835, -122.41)
	assert.NoError(t, err)
	assert.True(t, len(rate_hotels) > 0)
	// Check require = "price"
	price_hotels, err := service.GetRecommendations(ctx, "price", 37.7835, -122.41)
	assert.NoError(t, err)
	assert.True(t, len(price_hotels) > 0)
}
