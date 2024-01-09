package tests

import (
	"context"
	"fmt"
	"testing"

	"github.com/blueprint-uservices/blueprint/examples/train_ticket/workflow/assurance"
	"github.com/blueprint-uservices/blueprint/runtime/core/registry"
	"github.com/blueprint-uservices/blueprint/runtime/plugins/simplenosqldb"
	"github.com/stretchr/testify/require"
)

var assuranceServiceRegistry = registry.NewServiceRegistry[assurance.AssuranceService]("assurance_service")

func init() {
	assuranceServiceRegistry.Register("local", func(ctx context.Context) (assurance.AssuranceService, error) {
		db, err := simplenosqldb.NewSimpleNoSQLDB(ctx)
		if err != nil {
			return nil, err
		}
		return assurance.NewAssuranceServiceImpl(ctx, db)
	})
}

func genTestAssuranceData() []assurance.Assurance {
	res := []assurance.Assurance{}
	for i := 0; i < 10; i++ {
		a := assurance.Assurance{
			OrderID: fmt.Sprintf("Order%d", i),
			AT:      assurance.TRAFFIC_ACCIDENT,
		}
		res = append(res, a)
	}
	return res
}

func TestAssuranceService(t *testing.T) {
	ctx := context.Background()
	service, err := assuranceServiceRegistry.Get(ctx)
	require.NoError(t, err)

	// Test All Assurance types
	ats, err := service.GetAllAssuranceTypes(ctx)
	require.NoError(t, err)
	require.Len(t, ats, len(assurance.ALL_ASSURANCES))

	testData := genTestAssuranceData()
	// Test Create
	for idx, d := range testData {
		ass, err := service.Create(ctx, d.AT.Index, d.OrderID)
		require.NoError(t, err)
		// IDs are generated internally so we have to set the ID in the test data to that of the generated data
		d.ID = ass.ID
		testData[idx] = d
		requireAssurance(t, d, ass)
	}

	// Find by ID
	for _, d := range testData {
		ass, err := service.FindAssuranceById(ctx, d.ID)
		require.NoError(t, err)
		requireAssurance(t, d, ass)
	}

	// Find by Order ID
	for _, d := range testData {
		ass, err := service.FindAssuranceByOrderId(ctx, d.OrderID)
		require.NoError(t, err)
		requireAssurance(t, d, ass)
	}

	// Test GetAllAssurances
	assurances, err := service.GetAllAssurances(ctx)
	require.NoError(t, err)
	require.Len(t, assurances, len(testData))

	// Test Modify
	for i, d := range testData {
		d.OrderID = fmt.Sprintf("ID%d", i)
		ass, err := service.Modify(ctx, d)
		require.NoError(t, err)
		requireAssurance(t, d, ass)

		ass, err = service.FindAssuranceById(ctx, d.ID)
		require.NoError(t, err)
		requireAssurance(t, d, ass)
	}

	// Test Delete by ID
	for _, d := range testData {
		err = service.DeleteById(ctx, d.ID)
		require.NoError(t, err)
	}

	// Test Delete by OrderID
	testData = genTestAssuranceData()

	for _, d := range testData {
		_, err = service.Create(ctx, d.AT.Index, d.OrderID)
		err = service.DeleteByOrderId(ctx, d.OrderID)
		require.NoError(t, err)
	}
}

func requireAssuranceType(t *testing.T, expected assurance.AssuranceType, actual assurance.AssuranceType) {
	require.Equal(t, expected.Index, actual.Index)
	require.Equal(t, expected.Name, actual.Name)
	require.Equal(t, expected.Price, actual.Price)
}

func requireAssurance(t *testing.T, expected assurance.Assurance, actual assurance.Assurance) {
	require.Equal(t, expected.ID, actual.ID)
	require.Equal(t, expected.OrderID, actual.OrderID)
	requireAssuranceType(t, expected.AT, actual.AT)
}
