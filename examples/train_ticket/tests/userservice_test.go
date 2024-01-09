package tests

import (
	"context"
	"testing"

	"github.com/blueprint-uservices/blueprint/examples/train_ticket/workflow/user"
	"github.com/blueprint-uservices/blueprint/runtime/core/backend"
	"github.com/blueprint-uservices/blueprint/runtime/core/registry"
	"github.com/blueprint-uservices/blueprint/runtime/plugins/simplenosqldb"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
)

var userServiceRegistry = registry.NewServiceRegistry[user.UserService]("user_service")
var userDBRegistry = registry.NewServiceRegistry[backend.NoSQLDatabase]("user_db")

func init() {
	userDBRegistry.Register("local", func(ctx context.Context) (backend.NoSQLDatabase, error) {
		return simplenosqldb.NewSimpleNoSQLDB(ctx)
	})

	userServiceRegistry.Register("local", func(ctx context.Context) (user.UserService, error) {
		db, err := userDBRegistry.Get(ctx)
		if err != nil {
			return nil, err
		}
		return user.NewUserServiceImpl(ctx, db)
	})
}

var vaastav = user.User{
	UserID:       "5",
	Username:     "vaastav",
	Password:     "vaaspass",
	Gender:       0,
	DocumentType: 0,
	DocumentNum:  "0",
	Email:        "vaas@mpi",
}

var jcmace = user.User{
	UserID:       "2",
	Username:     "jcmace",
	Password:     "macepass",
	Gender:       0,
	DocumentType: 1,
	DocumentNum:  "0",
	Email:        "jon@microsoft",
}

var antoinek = user.User{
	UserID:       "1",
	Username:     "antoinek",
	Password:     "kaufpass",
	Gender:       0,
	DocumentType: 0,
	DocumentNum:  "1",
	Email:        "antoinek@mpi",
}

var rdeviti = user.User{
	UserID:       "4",
	Username:     "rdeviti",
	Password:     "robbpass",
	Gender:       1,
	DocumentType: 1,
	DocumentNum:  "1",
	Email:        "rdeviti@mpi",
}

var dg = user.User{
	UserID:       "3",
	Username:     "dg",
	Password:     "deepass",
	Gender:       0,
	DocumentType: 0,
	DocumentNum:  "0",
	Email:        "dg@mpi",
}

func TestUserService(t *testing.T) {
	ctx := context.Background()
	service, err := userServiceRegistry.Get(ctx)
	require.NoError(t, err)

	db, err := userDBRegistry.Get(ctx)
	require.NoError(t, err)
	coll, err := db.GetCollection(ctx, "user", "user")
	require.NoError(t, err)

	defer func() {
		// Empty the database on test exit
		err = coll.DeleteMany(ctx, bson.D{})
		require.NoError(t, err)
	}()

	// Make sure there are no users registered atm.
	users, err := service.GetAllUsers(ctx)
	require.NoError(t, err)
	require.Len(t, users, 0)

	// Register a user
	err = service.SaveUser(ctx, vaastav)
	require.NoError(t, err)

	// Make sure that the registered user exists.
	users, err = service.GetAllUsers(ctx)
	require.NoError(t, err)
	require.Len(t, users, 1)
	requireUser(t, vaastav, users[0])

	// Register all of our users
	err = service.SaveUser(ctx, jcmace)
	require.NoError(t, err)
	err = service.SaveUser(ctx, antoinek)
	require.NoError(t, err)
	err = service.SaveUser(ctx, dg)
	require.NoError(t, err)
	err = service.SaveUser(ctx, rdeviti)
	require.NoError(t, err)

	// Make sure all the registered users exist
	users, err = service.GetAllUsers(ctx)
	require.NoError(t, err)
	require.Len(t, users, 5)

	// Test the find methods
	all_users := []user.User{vaastav, antoinek, jcmace, dg, rdeviti}
	for _, user := range all_users {
		ret_user, err := service.FindByUserID(ctx, user.UserID)
		require.NoError(t, err)
		requireUser(t, user, ret_user)
		ret_user, err = service.FindByUsername(ctx, user.Username)
		require.NoError(t, err)
		requireUser(t, user, ret_user)
	}

	// Find non-existent users!
	_, err = service.FindByUserID(ctx, "9000")
	require.Error(t, err)

	_, err = service.FindByUsername(ctx, "foobar")
	require.Error(t, err)

	// Test the delete method
	err = service.DeleteUser(ctx, rdeviti.UserID)
	require.NoError(t, err)

	// Check that the number of total users has decreased
	users, err = service.GetAllUsers(ctx)
	require.NoError(t, err)
	require.Len(t, users, 4)

	// Try to find the deleted user
	_, err = service.FindByUserID(ctx, rdeviti.UserID)
	require.Error(t, err)

	_, err = service.FindByUsername(ctx, rdeviti.Username)
	require.Error(t, err)

	// Test the update method
	new_user := jcmace
	new_user.Email = "jcmace@mpi"
	success, err := service.UpdateUser(ctx, new_user)
	require.NoError(t, err)
	require.True(t, success)

	// Check that the updated user matches after find
	user, err := service.FindByUserID(ctx, new_user.UserID)
	requireUser(t, new_user, user)
}

func requireUser(t *testing.T, expected user.User, actual user.User) {
	require.Equal(t, expected.UserID, actual.UserID)
	require.Equal(t, expected.Username, actual.Username)
	require.Equal(t, expected.DocumentNum, actual.DocumentNum)
	require.Equal(t, expected.DocumentType, actual.DocumentType)
	require.Equal(t, expected.Email, actual.Email)
	require.Equal(t, expected.Gender, actual.Gender)
	require.Equal(t, expected.Password, actual.Password)
}
