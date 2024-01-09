package tests

import (
	"context"
	"fmt"
	"testing"

	"github.com/blueprint-uservices/blueprint/examples/train_ticket/workflow/consignprice"
	"github.com/blueprint-uservices/blueprint/runtime/core/registry"
	"github.com/blueprint-uservices/blueprint/runtime/plugins/simplenosqldb"
	"github.com/stretchr/testify/require"
)

var consignPriceServiceRegistry = registry.NewServiceRegistry[consignprice.ConsignPriceService]("consignprice_service")

func init() {
	consignPriceServiceRegistry.Register("local", func(ctx context.Context) (consignprice.ConsignPriceService, error) {
		db, err := simplenosqldb.NewSimpleNoSQLDB(ctx)
		if err != nil {
			return nil, err
		}
		return consignprice.NewConsignPriceServiceImpl(ctx, db)
	})
}

func genTestConsignPriceData() []consignprice.ConsignPrice {
	res := []consignprice.ConsignPrice{}
	for i := 0; i < 1; i++ {
		c := consignprice.ConsignPrice{
			ID:            fmt.Sprintf("CP%d", i),
			Index:         0,
			InitialWeight: 100,
			InitialPrice:  1000,
			WithinPrice:   50,
			BeyondPrice:   100,
		}
		res = append(res, c)
	}
	return res
}

func TestConsignPriceService(t *testing.T) {
	ctx := context.Background()
	service, err := consignPriceServiceRegistry.Get(ctx)
	require.NoError(t, err)

	testData := genTestConsignPriceData()
	test_conf := testData[0]

	// Test Create
	saved_conf, err := service.CreateAndModifyPriceConfig(ctx, test_conf)
	require.NoError(t, err)
	requireConsignPrice(t, test_conf, saved_conf)

	// Test GetPriceConfig
	saved_conf, err = service.GetPriceConfig(ctx)
	require.NoError(t, err)
	requireConsignPrice(t, test_conf, saved_conf)

	// Test GetPriceInfo
	str, err := service.GetPriceInfo(ctx)
	require.NoError(t, err)
	expected_str := fmt.Sprintf("The price of weight within %.2f is %.2f. The price of extra weight within the region is %.2f and beyond the region is %.2f", test_conf.InitialWeight, test_conf.InitialPrice, test_conf.WithinPrice, test_conf.BeyondPrice)
	require.Equal(t, expected_str, str)

	// Test GetPriceByWeightAndRegion
	// weight lower than initial weight
	price, err := service.GetPriceByWeightAndRegion(ctx, 50, false)
	require.NoError(t, err)
	require.Equal(t, test_conf.InitialPrice, price)
	price, err = service.GetPriceByWeightAndRegion(ctx, 50, true)
	require.NoError(t, err)
	require.Equal(t, test_conf.InitialPrice, price)

	// weight greater than initial weight and within region
	price, err = service.GetPriceByWeightAndRegion(ctx, 200, true)
	require.NoError(t, err)
	require.Equal(t, 100*test_conf.WithinPrice+test_conf.InitialPrice, price)
	// weight greater than initial weight and beyond region
	price, err = service.GetPriceByWeightAndRegion(ctx, 200, false)
	require.NoError(t, err)
	require.Equal(t, 100*test_conf.BeyondPrice+test_conf.InitialPrice, price)

	// Test Modify
	test_conf.BeyondPrice = 150
	saved_conf, err = service.CreateAndModifyPriceConfig(ctx, test_conf)
	require.NoError(t, err)
	requireConsignPrice(t, test_conf, saved_conf)
	saved_conf, err = service.GetPriceConfig(ctx)
	require.NoError(t, err)
	requireConsignPrice(t, test_conf, saved_conf)
}

func requireConsignPrice(t *testing.T, expected consignprice.ConsignPrice, actual consignprice.ConsignPrice) {
	require.Equal(t, expected.ID, actual.ID)
	require.Equal(t, expected.Index, actual.Index)
	require.Equal(t, expected.InitialWeight, actual.InitialWeight)
	require.Equal(t, expected.InitialPrice, actual.InitialPrice)
	require.Equal(t, expected.WithinPrice, actual.WithinPrice)
	require.Equal(t, expected.BeyondPrice, actual.BeyondPrice)
}
