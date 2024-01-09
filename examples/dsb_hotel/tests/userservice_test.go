package tests

import (
	"context"
	"testing"

	"github.com/blueprint-uservices/blueprint/examples/dsb_hotel/workflow/hotelreservation"
	"github.com/blueprint-uservices/blueprint/runtime/core/registry"
	"github.com/blueprint-uservices/blueprint/runtime/plugins/simplenosqldb"
	"github.com/stretchr/testify/assert"
)

// Tests acquire a UserService instance using a service registry.
// This enables us to run local unit tests, while also enabling
// the Blueprint test plugin to auto-generate tests
// for different deployments when compiling an application.
var userServiceRegistry = registry.NewServiceRegistry[hotelreservation.UserService]("user_service")

func init() {

	userServiceRegistry.Register("local", func(ctx context.Context) (hotelreservation.UserService, error) {
		db, err := simplenosqldb.NewSimpleNoSQLDB(ctx)
		if err != nil {
			return nil, err
		}

		return hotelreservation.NewUserServiceImpl(ctx, db)
	})
}

func TestCheckUser(t *testing.T) {
	ctx := context.Background()
	service, err := userServiceRegistry.Get(ctx)
	assert.NoError(t, err)

	valid, err := service.CheckUser(ctx, "Cornell_1", "1111111111")
	assert.NoError(t, err)
	assert.True(t, valid)
}
