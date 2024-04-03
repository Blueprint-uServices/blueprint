package tests

import (
	"context"
	"testing"

	"github.com/blueprint-uservices/blueprint/examples/dsb_sn/workflow/socialnetwork"
	"github.com/blueprint-uservices/blueprint/runtime/core/registry"
	"github.com/blueprint-uservices/blueprint/runtime/plugins/simplecache"
	"github.com/blueprint-uservices/blueprint/runtime/plugins/simplenosqldb"
	"github.com/stretchr/testify/require"
)

var socialGraphServiceRegistry = registry.NewServiceRegistry[socialnetwork.SocialGraphService]("socialgraph_service")

func init() {
	socialGraphServiceRegistry.Register("local", func(ctx context.Context) (socialnetwork.SocialGraphService, error) {
		cache, err := simplecache.NewSimpleCache(ctx)
		if err != nil {
			return nil, err
		}

		db, err := simplenosqldb.NewSimpleNoSQLDB(ctx)
		if err != nil {
			return nil, err
		}

		userIDService, err := userIDServiceRegistry.Get(ctx)
		if err != nil {
			return nil, err
		}

		return socialnetwork.NewSocialGraphServiceImpl(ctx, cache, db, userIDService)
	})
}

func TestInsertUser(t *testing.T) {
	ctx := context.Background()
	service, err := socialGraphServiceRegistry.Get(ctx)
	require.NoError(t, err)

	err = service.InsertUser(ctx, 1000, vaastav.UserID)
	require.NoError(t, err)

	// cleanup database
	cleanup_social_database(t, ctx)
}

func TestFollow(t *testing.T) {
	ctx := context.Background()
	service, err := socialGraphServiceRegistry.Get(ctx)
	require.NoError(t, err)

	// Add the users to the database first
	err = service.InsertUser(ctx, 1000, vaastav.UserID)
	require.NoError(t, err)
	err = service.InsertUser(ctx, 1001, jcmace.UserID)
	require.NoError(t, err)
	err = service.InsertUser(ctx, 1002, antoinek.UserID)
	require.NoError(t, err)

	// Test Follow
	err = service.Follow(ctx, 1003, vaastav.UserID, jcmace.UserID)
	require.NoError(t, err)

	// Add another follow to test multiple follows
	err = service.Follow(ctx, 1004, vaastav.UserID, antoinek.UserID)
	require.NoError(t, err)

	cleanup_social_database(t, ctx)
}

func TestFollowUsername(t *testing.T) {
	ctx := context.Background()
	service, err := socialGraphServiceRegistry.Get(ctx)
	require.NoError(t, err)

	userdb, err := userDBRegistry.Get(ctx)
	require.NoError(t, err)
	user_coll, err := userdb.GetCollection(ctx, "user", "user")
	require.NoError(t, err)

	err = user_coll.InsertOne(ctx, vaastav)
	require.NoError(t, err)
	err = user_coll.InsertOne(ctx, jcmace)
	require.NoError(t, err)
	err = user_coll.InsertOne(ctx, antoinek)
	require.NoError(t, err)

	// Add the users to the database first
	err = service.InsertUser(ctx, 1000, vaastav.UserID)
	require.NoError(t, err)
	err = service.InsertUser(ctx, 1001, jcmace.UserID)
	require.NoError(t, err)
	err = service.InsertUser(ctx, 1002, antoinek.UserID)
	require.NoError(t, err)

	// Test Follow
	err = service.FollowWithUsername(ctx, 1003, vaastav.Username, jcmace.Username)
	require.NoError(t, err)

	// Add another follow to test multiple follows
	err = service.FollowWithUsername(ctx, 1004, vaastav.Username, antoinek.Username)
	require.NoError(t, err)

	cleanup_social_database(t, ctx)
}

func TestGetFollowers(t *testing.T) {
	ctx := context.Background()
	service, err := socialGraphServiceRegistry.Get(ctx)
	require.NoError(t, err)

	// Add the users to the database first
	err = service.InsertUser(ctx, 1000, vaastav.UserID)
	require.NoError(t, err)
	err = service.InsertUser(ctx, 1001, jcmace.UserID)
	require.NoError(t, err)

	// Test GetFollowers before adding a follow
	followers, err := service.GetFollowers(ctx, 1002, vaastav.UserID)
	require.NoError(t, err)
	require.Len(t, followers, 0)

	followers, err = service.GetFollowers(ctx, 1003, jcmace.UserID)
	require.NoError(t, err)
	require.Len(t, followers, 0)

	// Test Follow
	err = service.Follow(ctx, 1003, vaastav.UserID, jcmace.UserID)
	require.NoError(t, err)

	// Test GetFollowers After adding a follow
	followers, err = service.GetFollowers(ctx, 1002, vaastav.UserID)
	require.NoError(t, err)
	require.Len(t, followers, 0)

	followers, err = service.GetFollowers(ctx, 1003, jcmace.UserID)
	require.NoError(t, err)
	require.Len(t, followers, 1)
	require.Equal(t, vaastav.UserID, followers[0])

	// Cleanup cache
	cleanup_social_database(t, ctx)
}

func TestGetFollowees(t *testing.T) {
	ctx := context.Background()
	service, err := socialGraphServiceRegistry.Get(ctx)
	require.NoError(t, err)

	// Add the users to the database first
	err = service.InsertUser(ctx, 1000, vaastav.UserID)
	require.NoError(t, err)
	err = service.InsertUser(ctx, 1001, jcmace.UserID)
	require.NoError(t, err)

	// Test GetFollowers before adding a follow
	followees, err := service.GetFollowees(ctx, 1002, vaastav.UserID)
	require.NoError(t, err)
	require.Len(t, followees, 0)

	followees, err = service.GetFollowees(ctx, 1003, jcmace.UserID)
	require.NoError(t, err)
	require.Len(t, followees, 0)

	// Test Follow
	err = service.Follow(ctx, 1003, vaastav.UserID, jcmace.UserID)
	require.NoError(t, err)

	// Test GetFollowers After adding a follow
	followees, err = service.GetFollowees(ctx, 1002, vaastav.UserID)
	require.NoError(t, err)
	require.Len(t, followees, 1)
	require.Equal(t, jcmace.UserID, followees[0])

	followees, err = service.GetFollowees(ctx, 1003, jcmace.UserID)
	require.NoError(t, err)
	require.Len(t, followees, 0)

	// Cleanup cache

	cleanup_social_database(t, ctx)
}

func TestUnfollow(t *testing.T) {
	ctx := context.Background()
	service, err := socialGraphServiceRegistry.Get(ctx)
	require.NoError(t, err)

	// Add the users to the database first
	err = service.InsertUser(ctx, 1000, vaastav.UserID)
	require.NoError(t, err)
	err = service.InsertUser(ctx, 1001, jcmace.UserID)
	require.NoError(t, err)

	// Test Follow
	err = service.Follow(ctx, 1003, vaastav.UserID, jcmace.UserID)
	require.NoError(t, err)

	// Test Unfollow
	err = service.Unfollow(ctx, 1004, vaastav.UserID, jcmace.UserID)
	require.NoError(t, err)

	// Clean up cache
	cleanup_social_database(t, ctx)
}
