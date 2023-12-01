package tests

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"gitlab.mpi-sws.org/cld/blueprint/examples/dsb_hotel/workflow/hotelreservation"
	"gitlab.mpi-sws.org/cld/blueprint/runtime/core/registry"
	"gitlab.mpi-sws.org/cld/blueprint/runtime/plugins/simplecache"
	"gitlab.mpi-sws.org/cld/blueprint/runtime/plugins/simplenosqldb"
)

var reservationServiceRegistry = registry.NewServiceRegistry[hotelreservation.ReservationService]("reservation_service")

func init() {
	reservationServiceRegistry.Register("local", func(ctx context.Context) (hotelreservation.ReservationService, error) {
		db, err := simplenosqldb.NewSimpleNoSQLDB(ctx)
		if err != nil {
			return nil, err
		}
		cache, err := simplecache.NewSimpleCache(ctx)
		if err != nil {
			return nil, err
		}

		return hotelreservation.NewReservationServiceImpl(ctx, cache, db)
	})
}

func TestCheckAvailability(t *testing.T) {
	ctx := context.Background()
	service, err := reservationServiceRegistry.Get(ctx)
	assert.NoError(t, err)

	hotels, err := service.CheckAvailability(ctx, "Vaastav", []string{"1"}, "2015-04-09", "2015-04-10", 1)
	assert.NoError(t, err)
	assert.Len(t, hotels, 1)
	assert.Equal(t, hotels[0], "1")
}

func TestMakeReservation(t *testing.T) {
	ctx := context.Background()
	service, err := reservationServiceRegistry.Get(ctx)
	assert.NoError(t, err)

	hotels, err := service.MakeReservation(ctx, "Vaastav", []string{"1"}, "2015-04-09", "2015-04-10", 1)
	assert.NoError(t, err)
	assert.Len(t, hotels, 1)
	assert.Equal(t, hotels[0], "1")
}
