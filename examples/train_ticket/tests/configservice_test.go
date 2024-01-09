package tests

import (
	"context"
	"fmt"
	"testing"

	"github.com/blueprint-uservices/blueprint/examples/train_ticket/workflow/config"
	"github.com/blueprint-uservices/blueprint/runtime/core/registry"
	"github.com/blueprint-uservices/blueprint/runtime/plugins/simplenosqldb"
	"github.com/stretchr/testify/require"
)

var configServiceRegistry = registry.NewServiceRegistry[config.ConfigService]("config_service")

func init() {
	configServiceRegistry.Register("local", func(ctx context.Context) (config.ConfigService, error) {
		db, err := simplenosqldb.NewSimpleNoSQLDB(ctx)
		if err != nil {
			return nil, err
		}
		return config.NewConfigServiceImpl(ctx, db)
	})
}

func genTestConfigsData() []config.Config {
	res := []config.Config{}
	for i := 0; i < 10; i++ {
		c := config.Config{
			Name:        fmt.Sprintf("ConfigVarName%d", i),
			Value:       fmt.Sprintf("%d", i),
			Description: "This is a config object",
		}
		res = append(res, c)
	}
	return res
}

func TestConfigService(t *testing.T) {
	ctx := context.Background()
	service, err := configServiceRegistry.Get(ctx)
	require.NoError(t, err)

	testData := genTestConfigsData()

	// Test Create
	for _, d := range testData {
		err = service.Create(ctx, d)
		require.NoError(t, err)
	}

	// TestFindALl
	all, err := service.FindAll(ctx)
	require.NoError(t, err)
	require.Len(t, all, len(testData))

	// Test Find
	for _, d := range testData {
		conf, err := service.Find(ctx, d.Name)
		require.NoError(t, err)
		requireConfig(t, d, conf)
	}

	// Test Update
	for _, d := range testData {
		d.Description = "Updated description"
		ok, err := service.Update(ctx, d)
		require.NoError(t, err)
		require.True(t, ok)
		conf, err := service.Find(ctx, d.Name)
		require.NoError(t, err)
		requireConfig(t, d, conf)
	}

	// Test Delete
	for _, d := range testData {
		err = service.Delete(ctx, d.Name)
		require.NoError(t, err)
	}
}

func requireConfig(t *testing.T, expected config.Config, actual config.Config) {
	require.Equal(t, expected.Name, actual.Name)
	require.Equal(t, expected.Value, actual.Value)
	require.Equal(t, expected.Description, actual.Description)
}
