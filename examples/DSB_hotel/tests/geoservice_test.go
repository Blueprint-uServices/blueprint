package tests

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"gitlab.mpi-sws.org/cld/blueprint/examples/DSB_hotel/workflow/hotelreservation"
	"gitlab.mpi-sws.org/cld/blueprint/runtime/core/registry"
	"gitlab.mpi-sws.org/cld/blueprint/runtime/plugins/simplenosqldb"
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
}
