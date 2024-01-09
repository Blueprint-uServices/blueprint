package tests

import (
	"context"
	"fmt"
	"testing"

	"github.com/blueprint-uservices/blueprint/examples/train_ticket/workflow/trainfood"
	"github.com/blueprint-uservices/blueprint/runtime/core/registry"
	"github.com/blueprint-uservices/blueprint/runtime/plugins/simplenosqldb"
	"github.com/stretchr/testify/require"
)

var trainfoodServiceRegistry = registry.NewServiceRegistry[trainfood.TrainFoodService]("trainfood_service")

func init() {
	trainfoodServiceRegistry.Register("local", func(ctx context.Context) (trainfood.TrainFoodService, error) {
		db, err := simplenosqldb.NewSimpleNoSQLDB(ctx)
		if err != nil {
			return nil, err
		}
		return trainfood.NewTrainFoodServiceImpl(ctx, db)
	})
}

func genTestTrainFoodData() []trainfood.TrainFood {
	res := []trainfood.TrainFood{}
	foods := genTestFoodData()
	for i := 0; i < 10; i++ {
		f := trainfood.TrainFood{
			ID:     fmt.Sprintf("ID%d", i),
			TripID: fmt.Sprintf("TripID%d", i),
			Foods:  foods[:i],
		}
		res = append(res, f)
	}
	return res
}

func TestTrainFoodService(t *testing.T) {
	ctx := context.Background()
	service, err := trainfoodServiceRegistry.Get(ctx)
	require.NoError(t, err)

	testData := genTestTrainFoodData()
	// Test Create
	for _, d := range testData {
		tf, err := service.CreateTrainFood(ctx, d)
		require.NoError(t, err)
		require.Equal(t, d, tf)
	}

	// Test ListTrainFood
	foods, err := service.ListTrainFood(ctx)
	require.NoError(t, err)
	require.Len(t, foods, len(testData))

	// Test ListTrainFoodByTripID
	for _, d := range testData {
		foods, err := service.ListTrainFoodByTripID(ctx, d.TripID)
		require.NoError(t, err)
		for i := 0; i < len(foods); i++ {
			requireFood(t, d.Foods[i], foods[i])
		}
	}

	// Cleanup
	err = service.Cleanup(ctx)
	require.NoError(t, err)
}

func requireTrainFood(t *testing.T, expected trainfood.TrainFood, actual trainfood.TrainFood) {
	require.Equal(t, expected.ID, actual.ID)
	require.Equal(t, expected.TripID, actual.TripID)
	require.Equal(t, len(expected.Foods), len(actual.Foods))
	for i := 0; i < len(expected.Foods); i++ {
		requireFood(t, expected.Foods[i], actual.Foods[i])
	}
}
