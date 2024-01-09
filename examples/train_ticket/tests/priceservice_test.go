package tests

import (
	"context"
	"fmt"
	"testing"

	"github.com/blueprint-uservices/blueprint/examples/train_ticket/workflow/price"
	"github.com/blueprint-uservices/blueprint/runtime/core/registry"
	"github.com/blueprint-uservices/blueprint/runtime/plugins/simplenosqldb"
	"github.com/stretchr/testify/require"
)

var priceServiceRegistry = registry.NewServiceRegistry[price.PriceService]("price_service")

func init() {
	priceServiceRegistry.Register("local", func(ctx context.Context) (price.PriceService, error) {
		db, err := simplenosqldb.NewSimpleNoSQLDB(ctx)
		if err != nil {
			return nil, err
		}
		return price.NewPriceServiceImpl(ctx, db)
	})
}

var trainTypes = []string{"TYPE1", "TYPE2", "TYPE3"}
var routes = []string{"ROUTE1", "ROUTE2", "ROUTE3"}

func genTestPriceData() []price.PriceConfig {
	res := []price.PriceConfig{}
	for i := 0; i < 9; i++ {
		c := price.PriceConfig{
			ID:                  fmt.Sprintf("%d", i),
			TrainType:           trainTypes[i%3],
			RouteID:             routes[i%3],
			BasicPriceRate:      0.7,
			FirstClassPriceRate: 0.9,
		}
		res = append(res, c)
	}
	return res
}

func TestPriceService(t *testing.T) {
	ctx := context.Background()
	service, err := priceServiceRegistry.Get(ctx)
	require.NoError(t, err)

	testData := genTestPriceData()

	// Check Creation
	for _, config := range testData {
		err = service.CreateNewPriceConfig(ctx, config)
		require.NoError(t, err)
	}

	// Check that all the configs made it to the database
	cs, err := service.GetAllPriceConfig(ctx)
	require.NoError(t, err)
	require.Len(t, cs, len(testData))

	// Find by ID
	for _, config := range testData {
		stored_c, err := service.FindByID(ctx, config.ID)
		require.NoError(t, err)
		requirePriceConfig(t, config, stored_c)
	}

	// Check updating
	for _, config := range testData {
		config.BasicPriceRate = 0.5
		ok, err := service.UpdatePriceConfig(ctx, config)
		require.NoError(t, err)
		require.True(t, ok)

		stored_c, err := service.FindByID(ctx, config.ID)
		require.NoError(t, err)
		requirePriceConfig(t, config, stored_c)
	}

	// FindBy RouteIDAndTrainType
	_, err = service.FindByRouteIDAndTrainType(ctx, routes[0], trainTypes[0])
	require.NoError(t, err)

	// Find a route + type combo that doesn't exist
	_, err = service.FindByRouteIDAndTrainType(ctx, routes[0], trainTypes[1])
	require.Error(t, err)

	multiroutes := []string{}
	for i := 0; i < 3; i++ {
		rt := routes[i] + ":" + trainTypes[i]
		res, err := service.FindByRouteIDsAndTrainTypes(ctx, []string{rt})
		require.NoError(t, err)
		require.Len(t, res, 1)
		multiroutes = append(multiroutes, rt)
	}
	res, err := service.FindByRouteIDsAndTrainTypes(ctx, multiroutes)
	require.NoError(t, err)
	require.Len(t, res, 3)

	mismatch_combos := []string{routes[0] + ":" + trainTypes[1]}
	res, err = service.FindByRouteIDsAndTrainTypes(ctx, mismatch_combos)
	require.NoError(t, err)
	require.Len(t, res, 0)

	// CHeck deletion
	for _, config := range testData {
		err = service.DeletePriceConfig(ctx, config.ID)
		require.NoError(t, err)
	}
}

func requirePriceConfig(t *testing.T, expected price.PriceConfig, actual price.PriceConfig) {
	require.Equal(t, expected.ID, actual.ID)
	require.Equal(t, expected.RouteID, actual.RouteID)
	require.Equal(t, expected.TrainType, actual.TrainType)
	require.Equal(t, expected.BasicPriceRate, actual.BasicPriceRate)
	require.Equal(t, expected.FirstClassPriceRate, actual.FirstClassPriceRate)
}
