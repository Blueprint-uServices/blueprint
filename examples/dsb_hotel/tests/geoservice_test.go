package tests

import (
	"context"
	"testing"

	"github.com/blueprint-uservices/blueprint/examples/dsb_hotel/workflow/hotelreservation"
	"github.com/blueprint-uservices/blueprint/runtime/core/registry"
	"github.com/blueprint-uservices/blueprint/runtime/plugins/simplenosqldb"
	"github.com/stretchr/testify/assert"
)

var geoServiceRegistry = registry.NewServiceRegistry[hotelreservation.GeoService]("geo_service")

func init() {

	geoServiceRegistry.Register("local", func(ctx context.Context) (hotelreservation.GeoService, error) {
		db, err := simplenosqldb.NewSimpleNoSQLDB(ctx)
		if err != nil {
			return nil, err
		}

		return hotelreservation.NewGeoServiceImpl(ctx, db)
	})
}

func TestNearby(t *testing.T) {
	ctx := context.Background()
	service, err := geoServiceRegistry.Get(ctx)
	assert.NoError(t, err)

	hotels, err := service.Nearby(ctx, 37.7835, -122.41)
	assert.NoError(t, err)
	assert.Len(t, hotels, 5)
	t.Log(hotels)
}
