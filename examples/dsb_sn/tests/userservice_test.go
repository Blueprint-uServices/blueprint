package tests

import (
	"context"
	"testing"

	"github.com/blueprint-uservices/blueprint/examples/dsb_sn/workflow/socialnetwork"
	"github.com/blueprint-uservices/blueprint/runtime/core/backend"
	"github.com/blueprint-uservices/blueprint/runtime/core/registry"
	"github.com/blueprint-uservices/blueprint/runtime/plugins/simplecache"
	"github.com/blueprint-uservices/blueprint/runtime/plugins/simplenosqldb"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
)

var userServiceRegistry = registry.NewServiceRegistry[socialnetwork.UserService]("user_service")

var userCacheRegistry = registry.NewServiceRegistry[backend.Cache]("user_cache")

var userDBRegistry = registry.NewServiceRegistry[backend.NoSQLDatabase]("user_db")

var vaastav = socialnetwork.User{
	FirstName: "Vaastav",
	LastName:  "Anand",
	UserID:    5,
	Username:  "vaastav",
}

var jcmace = socialnetwork.User{
	FirstName: "Jonathan",
	LastName:  "Mace",
	UserID:    2,
	Username:  "jcmace",
}

var antoinek = socialnetwork.User{
	FirstName: "Antoine",
	LastName:  "Kaufmann",
	UserID:    1,
	Username:  "antoinek",
}

var dg = socialnetwork.User{
	FirstName: "Deepak",
	LastName:  "Garg",
	UserID:    3,
	Username:  "dg",
}

func init() {

	userCacheRegistry.Register("local", func(ctx context.Context) (backend.Cache, error) {
		return simplecache.NewSimpleCache(ctx)
	})

	userDBRegistry.Register("local", func(ctx context.Context) (backend.NoSQLDatabase, error) {
		return simplenosqldb.NewSimpleNoSQLDB(ctx)
	})

	userServiceRegistry.Register("local", func(ctx context.Context) (socialnetwork.UserService, error) {
		cache, err := userCacheRegistry.Get(ctx)
		if err != nil {
			return nil, err
		}

		db, err := userDBRegistry.Get(ctx)
		if err != nil {
			return nil, err
		}

		socialgraphservice, err := socialGraphServiceRegistry.Get(ctx)
		if err != nil {
			return nil, err
		}

		return socialnetwork.NewUserServiceImpl(ctx, cache, db, socialgraphservice, "secret")
	})
}

func TestRegisterUserWithId(t *testing.T) {
	ctx := context.Background()
	service, err := userServiceRegistry.Get(ctx)
	require.NoError(t, err)

	db, err := userDBRegistry.Get(ctx)
	require.NoError(t, err)
	coll, err := db.GetCollection(ctx, "user", "user")
	require.NoError(t, err)

	soc_db, err := socialGraphDBRegistry.Get(ctx)
	require.NoError(t, err)
	soc_coll, err := soc_db.GetCollection(ctx, "social-graph", "social-graph")
	require.NoError(t, err)

	// Check database is empty before we add any users
	var downloaded_users []socialnetwork.User
	vals, err := coll.FindMany(ctx, bson.D{})
	require.NoError(t, err)
	err = vals.All(ctx, &downloaded_users)
	require.NoError(t, err)
	require.Len(t, downloaded_users, 0)

	// Add users
	req_id := 1000
	users := []socialnetwork.User{vaastav, antoinek, jcmace, dg}
	for _, user := range users {
		req_id += 1
		err = service.RegisterUserWithId(ctx, int64(req_id), user.FirstName, user.LastName, user.Username, "vaaspwd", user.UserID)
		require.NoError(t, err)
	}

	// Try duplicate registration
	for _, user := range users {
		req_id += 1
		err = service.RegisterUserWithId(ctx, int64(req_id), user.FirstName, user.LastName, user.Username, "vaaspwd", user.UserID)
		require.Error(t, err)
	}

	// Check values in database

	vals, err = coll.FindMany(ctx, bson.D{})
	require.NoError(t, err)
	err = vals.All(ctx, &downloaded_users)
	require.NoError(t, err)
	require.Len(t, downloaded_users, 4)

	// Individually check each user with an individual query
	for _, test_user := range users {
		var user socialnetwork.User
		val, err := coll.FindOne(ctx, bson.D{{"username", test_user.Username}})
		require.NoError(t, err)
		ok, err := val.One(ctx, &user)
		require.NoError(t, err)
		require.True(t, ok)
		requireUser(t, test_user, user, true)
	}

	// Cleanup database
	err = coll.DeleteMany(ctx, bson.D{})
	require.NoError(t, err)
	err = soc_coll.DeleteMany(ctx, bson.D{})
	require.NoError(t, err)
}

func TestLogin(t *testing.T) {
	ctx := context.Background()
	service, err := userServiceRegistry.Get(ctx)
	require.NoError(t, err)

	db, err := userDBRegistry.Get(ctx)
	require.NoError(t, err)
	coll, err := db.GetCollection(ctx, "user", "user")
	require.NoError(t, err)

	cache, err := userCacheRegistry.Get(ctx)
	require.NoError(t, err)

	soc_db, err := socialGraphDBRegistry.Get(ctx)
	require.NoError(t, err)
	soc_coll, err := soc_db.GetCollection(ctx, "social-graph", "social-graph")
	require.NoError(t, err)

	// Check database is empty before we add any users
	var downloaded_users []socialnetwork.User
	vals, err := coll.FindMany(ctx, bson.D{})
	require.NoError(t, err)
	err = vals.All(ctx, &downloaded_users)
	require.NoError(t, err)
	require.Len(t, downloaded_users, 0)

	// Add users
	req_id := 1000
	users := []socialnetwork.User{vaastav, antoinek, jcmace, dg}
	for _, user := range users {
		err = service.RegisterUserWithId(ctx, int64(req_id), user.FirstName, user.LastName, user.Username, "vaaspwd", user.UserID)
		require.NoError(t, err)
		req_id += 1

		res, err := service.Login(ctx, int64(req_id), user.Username, "vaaspwd")
		require.NoError(t, err)
		require.NotEqual(t, "", res)
		req_id += 1
	}

	// Cleanup cache
	for _, user := range users {
		req_id += 1
		err = cache.Delete(ctx, user.Username+":Login")
		require.NoError(t, err)
	}

	// Cleanup database
	err = coll.DeleteMany(ctx, bson.D{})
	require.NoError(t, err)
	err = soc_coll.DeleteMany(ctx, bson.D{})
	require.NoError(t, err)
}

func TestRegisterUser(t *testing.T) {
	ctx := context.Background()
	service, err := userServiceRegistry.Get(ctx)
	require.NoError(t, err)

	db, err := userDBRegistry.Get(ctx)
	require.NoError(t, err)
	coll, err := db.GetCollection(ctx, "user", "user")
	require.NoError(t, err)

	soc_db, err := socialGraphDBRegistry.Get(ctx)
	require.NoError(t, err)
	soc_coll, err := soc_db.GetCollection(ctx, "social-graph", "social-graph")
	require.NoError(t, err)

	// Check database is empty before we add any users
	var downloaded_users []socialnetwork.User
	vals, err := coll.FindMany(ctx, bson.D{})
	require.NoError(t, err)
	err = vals.All(ctx, &downloaded_users)
	require.NoError(t, err)
	require.Len(t, downloaded_users, 0)

	// Add users
	base_req_id := 1000
	users := []socialnetwork.User{vaastav, antoinek, jcmace, dg}
	for idx, user := range users {
		err = service.RegisterUser(ctx, int64(base_req_id)+int64(idx), user.FirstName, user.LastName, user.Username, "vaaspwd")
		require.NoError(t, err)
	}

	// Check values in database

	vals, err = coll.FindMany(ctx, bson.D{})
	require.NoError(t, err)
	err = vals.All(ctx, &downloaded_users)
	require.NoError(t, err)
	require.Len(t, downloaded_users, 4)

	// Individually check each user with an individual query
	for _, test_user := range users {
		var user socialnetwork.User
		val, err := coll.FindOne(ctx, bson.D{{"username", test_user.Username}})
		require.NoError(t, err)
		ok, err := val.One(ctx, &user)
		require.NoError(t, err)
		require.True(t, ok)
		requireUser(t, test_user, user, false)
	}

	// Cleanup database
	err = coll.DeleteMany(ctx, bson.D{})
	require.NoError(t, err)
	err = soc_coll.DeleteMany(ctx, bson.D{})
	require.NoError(t, err)
}

func TestComposeCreatorWithUserId(t *testing.T) {
	ctx := context.Background()
	service, err := userServiceRegistry.Get(ctx)
	require.NoError(t, err)

	creator, err := service.ComposeCreatorWithUserId(ctx, 1000, vaastav.UserID, vaastav.Username)
	require.NoError(t, err)
	require.Equal(t, vaastav.UserID, creator.UserID)
	require.Equal(t, vaastav.Username, creator.Username)
}

func TestComposeCreatorWithUsername(t *testing.T) {
	ctx := context.Background()
	service, err := userServiceRegistry.Get(ctx)
	require.NoError(t, err)

	db, err := userDBRegistry.Get(ctx)
	require.NoError(t, err)
	coll, err := db.GetCollection(ctx, "user", "user")
	require.NoError(t, err)

	cache, err := userCacheRegistry.Get(ctx)
	require.NoError(t, err)

	soc_db, err := socialGraphDBRegistry.Get(ctx)
	require.NoError(t, err)
	soc_coll, err := soc_db.GetCollection(ctx, "social-graph", "social-graph")
	require.NoError(t, err)

	// Add users
	req_id := 1000
	users := []socialnetwork.User{vaastav, antoinek, jcmace, dg}
	for _, user := range users {
		req_id += 1
		err = service.RegisterUserWithId(ctx, int64(req_id), user.FirstName, user.LastName, user.Username, "vaaspwd", user.UserID)
		require.NoError(t, err)
	}

	// Test Compose with existing user
	for _, user := range users {
		req_id += 1
		creator, err := service.ComposeCreatorWithUsername(ctx, int64(req_id), user.Username)
		require.NoError(t, err)
		require.Equal(t, user.UserID, creator.UserID)
		require.Equal(t, user.Username, creator.Username)
	}

	// Test Compose with non-existing user
	_, err = service.ComposeCreatorWithUsername(ctx, 1500, "foobar")
	require.Error(t, err)

	// Cleanup cache
	for _, user := range users {
		req_id += 1
		err = cache.Delete(ctx, user.Username+":UserID")
		require.NoError(t, err)
	}

	// Cleanup database
	err = coll.DeleteMany(ctx, bson.D{})
	require.NoError(t, err)
	err = soc_coll.DeleteMany(ctx, bson.D{})
	require.NoError(t, err)

}

func requireUser(t *testing.T, u1 socialnetwork.User, u2 socialnetwork.User, checkId bool) {
	if checkId {
		require.Equal(t, u1.UserID, u2.UserID)
	}
	require.Equal(t, u1.FirstName, u2.FirstName)
	require.Equal(t, u1.LastName, u2.LastName)
	require.Equal(t, u1.Username, u2.Username)
}
