package tests

import (
	"context"
	"fmt"
	"testing"

	"github.com/blueprint-uservices/blueprint/examples/train_ticket/workflow/stationfood"
	"github.com/blueprint-uservices/blueprint/runtime/core/registry"
	"github.com/blueprint-uservices/blueprint/runtime/plugins/simplenosqldb"
	"github.com/stretchr/testify/require"
)

var stationfoodServiceRegistry = registry.NewServiceRegistry[stationfood.StationFoodService]("stationfood_service")

func init() {
	stationfoodServiceRegistry.Register("local", func(ctx context.Context) (stationfood.StationFoodService, error) {
		db, err := simplenosqldb.NewSimpleNoSQLDB(ctx)
		if err != nil {
			return nil, err
		}
		return stationfood.NewStationFoodServiceImpl(ctx, db)
	})
}

func genTestStationFoodData() []stationfood.StationFoodStore {
	res := []stationfood.StationFoodStore{}
	foods := genTestFoodData()
	for i := 0; i < 10; i++ {
		s := stationfood.StationFoodStore{
			ID:           fmt.Sprintf("ID%d", i),
			StationName:  fmt.Sprintf("Station%d", i%2),
			StoreName:    fmt.Sprintf("Store%d", i),
			Telephone:    "5550123",
			BusinessTime: "0900",
			DeliveryFee:  float64(i*100) + 100,
			Foods:        foods[:i],
		}
		res = append(res, s)
	}
	return res
}

func TestStationFoodService(t *testing.T) {
	ctx := context.Background()
	service, err := stationfoodServiceRegistry.Get(ctx)
	require.NoError(t, err)

	testData := genTestStationFoodData()

	// Test CreateFoodStore
	for _, d := range testData {
		err = service.CreateFoodStore(ctx, d)
		require.NoError(t, err)
	}

	// Test ListFoodStores
	stores, err := service.ListFoodStores(ctx)
	require.Len(t, stores, len(testData))

	// Test GetFoodStoreByID
	for _, d := range testData {
		store, err := service.GetFoodStoreByID(ctx, d.ID)
		require.NoError(t, err)
		requireStationFoodStore(t, d, store)
	}

	// Test ListFoodStoresByStationName
	for _, d := range testData {
		stores, err = service.ListFoodStoresByStationName(ctx, d.StationName)
		require.NoError(t, err)
		require.Len(t, stores, int(len(testData)/2))
	}

	// Test GetFoodStoresByStationName
	stations := []string{"Station0", "Station1"}
	stores, err = service.GetFoodStoresByStationNames(ctx, stations)
	require.NoError(t, err)
	require.Len(t, stores, len(testData))

	// CLeanup
	err = service.Cleanup(ctx)
	require.NoError(t, err)
}

func requireStationFoodStore(t *testing.T, expected stationfood.StationFoodStore, actual stationfood.StationFoodStore) {
	require.Equal(t, expected.ID, actual.ID)
	require.Equal(t, expected.StationName, actual.StationName)
	require.Equal(t, expected.StoreName, actual.StoreName)
	require.Equal(t, expected.Telephone, actual.Telephone)
	require.Equal(t, expected.BusinessTime, actual.BusinessTime)
	require.Equal(t, expected.DeliveryFee, actual.DeliveryFee)
	require.Equal(t, len(expected.Foods), len(actual.Foods))
	for i := 0; i < len(expected.Foods); i++ {
		requireFood(t, expected.Foods[i], actual.Foods[i])
	}
}
