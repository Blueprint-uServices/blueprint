package tests

import (
	"context"
	"testing"

	"github.com/blueprint-uservices/blueprint/examples/dsb_hotel/workflow/hotelreservation"
	"github.com/blueprint-uservices/blueprint/runtime/core/registry"
	"github.com/blueprint-uservices/blueprint/runtime/plugins/simplenosqldb"
	"github.com/stretchr/testify/assert"
)

var attractionsServiceRegistry = registry.NewServiceRegistry[hotelreservation.AttractionsService]("attractions_service")

func init() {
	attractionsServiceRegistry.Register("local", func(ctx context.Context) (hotelreservation.AttractionsService, error) {
		db, err := simplenosqldb.NewSimpleNoSQLDB(ctx)
		if err != nil {
			return nil, err
		}
		cinemas, err := db.GetCollection(ctx, "attractions-db", "cinemas")
		if err != nil {
			return nil, err
		}
		if err := cinemas.InsertOne(ctx, &hotelreservation.Cinema{
			CinemaId: "1", CinemaName: "C1", CLat: 37.7835, CLon: -122.41, Type: "independent",
		}); err != nil {
			return nil, err
		}
		return hotelreservation.NewAttractionsServiceImpl(ctx, db)
	})
}

func TestNearbyAttractions(t *testing.T) {
	ctx := context.Background()
	service, err := attractionsServiceRegistry.Get(ctx)
	assert.NoError(t, err)

	restaurants, err := service.NearbyRest(ctx, 37.7835, -122.41)
	assert.NoError(t, err)
	assert.Len(t, restaurants, 5)

	museums, err := service.NearbyMus(ctx, 37.7835, -122.41)
	assert.NoError(t, err)
	assert.Equal(t, []string{"4"}, museums)

	cinemas, err := service.NearbyCinema(ctx, 37.7835, -122.41)
	assert.NoError(t, err)
	assert.Equal(t, []string{"1"}, cinemas)
}
