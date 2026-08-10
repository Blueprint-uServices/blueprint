package tests

import (
	"context"
	"testing"

	"github.com/blueprint-uservices/blueprint/examples/dsb_hotel/workflow/hotelreservation"
	"github.com/blueprint-uservices/blueprint/runtime/core/registry"
	"github.com/blueprint-uservices/blueprint/runtime/plugins/simplecache"
	"github.com/blueprint-uservices/blueprint/runtime/plugins/simplenosqldb"
	"github.com/stretchr/testify/assert"
)

var reviewServiceRegistry = registry.NewServiceRegistry[hotelreservation.ReviewService]("review_service")

func init() {
	reviewServiceRegistry.Register("local", func(ctx context.Context) (hotelreservation.ReviewService, error) {
		db, err := simplenosqldb.NewSimpleNoSQLDB(ctx)
		if err != nil {
			return nil, err
		}
		cache, err := simplecache.NewSimpleCache(ctx)
		if err != nil {
			return nil, err
		}
		return hotelreservation.NewReviewServiceImpl(ctx, cache, db)
	})
}

func TestGetReviews(t *testing.T) {
	ctx := context.Background()
	service, err := reviewServiceRegistry.Get(ctx)
	assert.NoError(t, err)

	reviews, err := service.GetReviews(ctx, "1")
	assert.NoError(t, err)
	assert.Len(t, reviews, 4)
	assert.Equal(t, "Person 1", reviews[0].Name)
	assert.Equal(t, "Person 4", reviews[3].Name)

	cachedReviews, err := service.GetReviews(ctx, "1")
	assert.NoError(t, err)
	assert.Equal(t, reviews, cachedReviews)

	reviews, err = service.GetReviews(ctx, "2")
	assert.NoError(t, err)
	assert.Len(t, reviews, 2)

	reviews, err = service.GetReviews(ctx, "unknown")
	assert.NoError(t, err)
	assert.Empty(t, reviews)
}
