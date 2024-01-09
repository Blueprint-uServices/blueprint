package tests

import (
	"context"
	"fmt"
	"testing"

	"github.com/blueprint-uservices/blueprint/examples/train_ticket/workflow/station"
	"github.com/blueprint-uservices/blueprint/runtime/core/registry"
	"github.com/blueprint-uservices/blueprint/runtime/plugins/simplenosqldb"
	"github.com/stretchr/testify/require"
)

var stationServiceRegistry = registry.NewServiceRegistry[station.StationService]("station_service")

func init() {
	stationServiceRegistry.Register("local", func(ctx context.Context) (station.StationService, error) {
		db, err := simplenosqldb.NewSimpleNoSQLDB(ctx)
		if err != nil {
			return nil, err
		}
		return station.NewStationServiceImpl(ctx, db)
	})
}

func genTestStationsData() []station.Station {
	res := []station.Station{}
	for i := 0; i < 10; i++ {
		s := station.Station{
			ID:       fmt.Sprintf("s%d", i),
			Name:     fmt.Sprintf("Station%d", i),
			StayTime: 1000,
		}
		res = append(res, s)
	}
	return res
}

func TestStationService(t *testing.T) {
	ctx := context.Background()
	service, err := stationServiceRegistry.Get(ctx)
	require.NoError(t, err)

	testData := genTestStationsData()
	// TestCreateStation
	for _, d := range testData {
		err = service.CreateStation(ctx, d)
		require.NoError(t, err)
	}

	// CHeck if these stations exist
	for _, d := range testData {
		ok, err := service.Exists(ctx, d.Name)
		require.NoError(t, err)
		require.True(t, ok)
	}

	// Find id one by one!
	for _, d := range testData {
		id, err := service.FindID(ctx, d.Name)
		require.NoError(t, err)
		require.Equal(t, d.ID, id)
	}

	// Find all ids directly
	all_names := []string{}
	for _, d := range testData {
		all_names = append(all_names, d.Name)
	}
	ids, err := service.FindIDs(ctx, all_names)
	require.NoError(t, err)
	for idx, d := range testData {
		require.Equal(t, d.ID, ids[idx])
	}

	// Find station using id
	all_ids := []string{}
	for _, d := range testData {
		st, err := service.FindByID(ctx, d.ID)
		require.NoError(t, err)
		requireStation(t, d, st)
		all_ids = append(all_ids, d.ID)
	}

	all_stations, err := service.FindByIDs(ctx, all_ids)
	require.NoError(t, err)
	require.Len(t, all_stations, len(testData))
	for idx, d := range testData {
		requireStation(t, d, all_stations[idx])
	}

	// Check updates
	for _, d := range testData {
		d.StayTime = 2000
		ok, err := service.UpdateStation(ctx, d)
		require.True(t, ok)
		require.NoError(t, err)
		st, err := service.FindByID(ctx, d.ID)
		require.NoError(t, err)
		requireStation(t, d, st)
	}

	// Check deletion
	for _, d := range testData {
		err = service.DeleteStation(ctx, d.ID)
		require.NoError(t, err)
	}
}

func requireStation(t *testing.T, expected station.Station, actual station.Station) {
	require.Equal(t, expected.ID, actual.ID)
	require.Equal(t, expected.Name, actual.Name)
	require.Equal(t, expected.StayTime, actual.StayTime)
}
