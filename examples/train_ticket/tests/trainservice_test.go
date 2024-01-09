package tests

import (
	"context"
	"fmt"
	"testing"

	"github.com/blueprint-uservices/blueprint/examples/train_ticket/workflow/train"
	"github.com/blueprint-uservices/blueprint/runtime/core/registry"
	"github.com/blueprint-uservices/blueprint/runtime/plugins/simplenosqldb"
	"github.com/stretchr/testify/require"
)

var trainServiceRegistry = registry.NewServiceRegistry[train.TrainService]("train_service")

func init() {
	trainServiceRegistry.Register("local", func(ctx context.Context) (train.TrainService, error) {
		db, err := simplenosqldb.NewSimpleNoSQLDB(ctx)
		if err != nil {
			return nil, err
		}
		return train.NewTrainServiceImpl(ctx, db)
	})
}

func genTestTrainsData() []train.TrainType {
	res := []train.TrainType{}
	for i := 0; i < 10; i++ {
		t := train.TrainType{
			ID:           fmt.Sprintf("t%d", i),
			Name:         fmt.Sprintf("Train%d", i),
			EconomyClass: int64(i),
			ComfortClass: int64(i),
			AvgSpeed:     int64((10 - i + 1) * 100),
		}
		res = append(res, t)
	}
	return res
}

func TestTrainService(t *testing.T) {
	ctx := context.Background()
	service, err := trainServiceRegistry.Get(ctx)
	require.NoError(t, err)

	testData := genTestTrainsData()
	// TestCreate TrainType
	for _, d := range testData {
		ok, err := service.Create(ctx, d)
		require.NoError(t, err)
		require.True(t, ok)
	}

	// Find train type one by one by id!
	for _, d := range testData {
		tt, err := service.Retrieve(ctx, d.ID)
		require.NoError(t, err)
		requireTrainType(t, d, tt)
	}

	// Find train type one by one by name!
	for _, d := range testData {
		tt, err := service.RetrieveByName(ctx, d.Name)
		require.NoError(t, err)
		requireTrainType(t, d, tt)
	}

	// Test all trains
	trains, err := service.AllTrains(ctx)
	require.NoError(t, err)
	require.Len(t, trains, len(testData))

	// Test RetrieveByNames
	all_names := []string{}
	for _, d := range testData {
		all_names = append(all_names, d.Name)
	}
	trains, err = service.RetrieveByNames(ctx, all_names)
	for idx, _ := range testData {
		requireTrainType(t, testData[idx], trains[idx])
	}

	// Test update
	for _, d := range testData {
		d.AvgSpeed = 1000
		ok, err := service.Update(ctx, d)
		require.NoError(t, err)
		require.True(t, ok)
		tt, err := service.Retrieve(ctx, d.ID)
		require.NoError(t, err)
		requireTrainType(t, d, tt)
	}

	// Delete all trains
	for _, d := range testData {
		ok, err := service.Delete(ctx, d.ID)
		require.NoError(t, err)
		require.True(t, ok)
	}
}

func requireTrainType(t *testing.T, expected train.TrainType, actual train.TrainType) {
	require.Equal(t, expected.ID, actual.ID)
	require.Equal(t, expected.Name, actual.Name)
	require.Equal(t, expected.EconomyClass, actual.EconomyClass)
	require.Equal(t, expected.ComfortClass, actual.ComfortClass)
	require.Equal(t, expected.AvgSpeed, actual.AvgSpeed)
}
