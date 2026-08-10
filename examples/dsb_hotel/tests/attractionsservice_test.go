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
	assert.ElementsMatch(t, []hotelreservation.Restaurant{
		{RestaurantId: "1", RLat: 37.7867, RLon: -122.4112, RestaurantName: "R1", Rating: 3.5, Type: "fusion"},
		{RestaurantId: "2", RLat: 37.7857, RLon: -122.4012, RestaurantName: "R2", Rating: 3.9, Type: "italian"},
		{RestaurantId: "3", RLat: 37.7847, RLon: -122.3912, RestaurantName: "R3", Rating: 4.5, Type: "sushi"},
		{RestaurantId: "4", RLat: 37.7862, RLon: -122.4212, RestaurantName: "R4", Rating: 3.2, Type: "sushi"},
		{RestaurantId: "5", RLat: 37.7839, RLon: -122.4052, RestaurantName: "R5", Rating: 4.9, Type: "fusion"},
	}, restaurants)

	museums, err := service.NearbyMus(ctx, 37.7835, -122.41)
	assert.NoError(t, err)
	assert.Equal(t, []hotelreservation.Museum{
		{MuseumId: "4", MLat: 37.7867, MLon: -122.4912, MuseumName: "M4", Type: "nature"},
	}, museums)

	cinemas, err := service.NearbyCinema(ctx, 37.7835, -122.41)
	assert.NoError(t, err)
	assert.Equal(t, []hotelreservation.Cinema{
		{CinemaId: "1", CLat: 37.7835, CLon: -122.41, CinemaName: "C1", Type: "independent"},
	}, cinemas)
}
