package tests

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"gitlab.mpi-sws.org/cld/blueprint/examples/DSB_sn/workflow/socialnetwork"
	"gitlab.mpi-sws.org/cld/blueprint/runtime/core/registry"
	"go.mongodb.org/mongo-driver/bson"
)

var userIDServiceRegistry = registry.NewServiceRegistry[socialnetwork.UserIDService]("userId_service")

var hello_user = socialnetwork.User{
	FirstName: "Hello",
	LastName:  "World!",
	UserID:    10,
	Username:  "hello",
}

func init() {
	userIDServiceRegistry.Register("local", func(ctx context.Context) (socialnetwork.UserIDService, error) {
		cache, err := userCacheRegistry.Get(ctx)
		if err != nil {
			return nil, err
		}

		db, err := userDBRegistry.Get(ctx)
		if err != nil {
			return nil, err
		}

		return socialnetwork.NewUserIDServiceImpl(ctx, cache, db)
	})
}

func TestGetUserID(t *testing.T) {
	ctx := context.Background()
	service, err := userIDServiceRegistry.Get(ctx)
	require.NoError(t, err)

	// Check username that is not in the database
	id, err := service.GetUserId(ctx, 1000, hello_user.Username)
	require.Error(t, err)
	require.Equal(t, int64(-1), id)

	// Add the previously non-existent user into the database.
	db, err := userDBRegistry.Get(ctx)
	require.NoError(t, err)
	coll, err := db.GetCollection(ctx, "user", "user")
	require.NoError(t, err)

	err = coll.InsertOne(ctx, hello_user)
	require.NoError(t, err)

	// GetUserId should now succeed
	id, err = service.GetUserId(ctx, 1001, hello_user.Username)
	require.NoError(t, err)
	require.Equal(t, hello_user.UserID, id)

	// Cleanup database

	err = coll.DeleteOne(ctx, bson.D{{"username", hello_user.Username}})
	require.NoError(t, err)
}
