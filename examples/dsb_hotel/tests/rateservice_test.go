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

var rateServiceRegistry = registry.NewServiceRegistry[hotelreservation.RateService]("rate_service")

func init() {
	rateServiceRegistry.Register("local", func(ctx context.Context) (hotelreservation.RateService, error) {
		db, err := simplenosqldb.NewSimpleNoSQLDB(ctx)
		if err != nil {
			return nil, err
		}
		cache, err := simplecache.NewSimpleCache(ctx)
		if err != nil {
			return nil, err
		}
		return hotelreservation.NewRateServiceImpl(ctx, cache, db)
	})
}

func TestGetRates(t *testing.T) {
	ctx := context.Background()
	service, err := rateServiceRegistry.Get(ctx)
	assert.NoError(t, err)

	plans, err := service.GetRates(ctx, []string{"1", "2", "3", "12", "9"}, "2015-04-09", "2015-04-10")
	assert.NoError(t, err)
	assert.Len(t, plans, 5)
}
