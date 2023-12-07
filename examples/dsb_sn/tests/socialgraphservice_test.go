package tests

import (
	"context"
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"
	"gitlab.mpi-sws.org/cld/blueprint/examples/dsb_sn/workflow/socialnetwork"
	"gitlab.mpi-sws.org/cld/blueprint/runtime/core/backend"
	"gitlab.mpi-sws.org/cld/blueprint/runtime/core/registry"
	"gitlab.mpi-sws.org/cld/blueprint/runtime/plugins/simplecache"
	"gitlab.mpi-sws.org/cld/blueprint/runtime/plugins/simplenosqldb"
	"go.mongodb.org/mongo-driver/bson"
)

var socialGraphServiceRegistry = registry.NewServiceRegistry[socialnetwork.SocialGraphService]("socialgraph_service")
var socialGraphDBRegistry = registry.NewServiceRegistry[backend.NoSQLDatabase]("socialgraph_db")
var socialGraphCacheRegistry = registry.NewServiceRegistry[backend.Cache]("socialgraph_cache")

func init() {

	socialGraphDBRegistry.Register("local", func(ctx context.Context) (backend.NoSQLDatabase, error) {
		return simplenosqldb.NewSimpleNoSQLDB(ctx)
	})

	socialGraphCacheRegistry.Register("local", func(ctx context.Context) (backend.Cache, error) {
		return simplecache.NewSimpleCache(ctx)
	})

	socialGraphServiceRegistry.Register("local", func(ctx context.Context) (socialnetwork.SocialGraphService, error) {
		cache, err := socialGraphCacheRegistry.Get(ctx)
		if err != nil {
			return nil, err
		}

		db, err := socialGraphDBRegistry.Get(ctx)
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
	db, err := socialGraphDBRegistry.Get(ctx)
	require.NoError(t, err)
	coll, err := db.GetCollection(ctx, "social-graph", "social-graph")
	require.NoError(t, err)

	err = coll.DeleteOne(ctx, bson.D{{"userid", vaastav.UserID}})
	require.NoError(t, err)
}

func TestFollow(t *testing.T) {
	ctx := context.Background()
	service, err := socialGraphServiceRegistry.Get(ctx)
	require.NoError(t, err)

	db, err := socialGraphDBRegistry.Get(ctx)
	require.NoError(t, err)
	coll, err := db.GetCollection(ctx, "social-graph", "social-graph")
	require.NoError(t, err)

	cache, err := socialGraphCacheRegistry.Get(ctx)
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

	// Check cache contents
	vaastavID := strconv.FormatInt(vaastav.UserID, 10)
	jcmaceID := strconv.FormatInt(jcmace.UserID, 10)
	antoinekID := strconv.FormatInt(antoinek.UserID, 10)
	var followeeinfo []socialnetwork.FolloweeInfo
	var followerinfo []socialnetwork.FollowerInfo
	exists, err := cache.Get(ctx, vaastavID+":followees", &followeeinfo)
	require.NoError(t, err)
	require.True(t, exists)
	require.Len(t, followeeinfo, 1)
	exists, err = cache.Get(ctx, jcmaceID+":followers", &followerinfo)
	require.NoError(t, err)
	require.True(t, exists)
	require.Len(t, followerinfo, 1)

	// Add another follow to test multiple follows
	err = service.Follow(ctx, 1004, vaastav.UserID, antoinek.UserID)
	require.NoError(t, err)

	// Check cache contents
	exists, err = cache.Get(ctx, vaastavID+":followees", &followeeinfo)
	require.NoError(t, err)
	require.True(t, exists)
	require.Len(t, followeeinfo, 2)
	exists, err = cache.Get(ctx, antoinekID+":followers", &followerinfo)
	require.NoError(t, err)
	require.True(t, exists)
	require.Len(t, followerinfo, 1)

	// Cleanup cache
	err = cache.Delete(ctx, vaastavID+":followees")
	require.NoError(t, err)
	err = cache.Delete(ctx, jcmaceID+":followers")
	require.NoError(t, err)
	err = cache.Delete(ctx, antoinekID+":followers")

	// Check database contents
	var vaasInfo socialnetwork.UserInfo
	var jcInfo socialnetwork.UserInfo
	var akInfo socialnetwork.UserInfo

	val, err := coll.FindOne(ctx, bson.D{{"userid", vaastav.UserID}})
	require.NoError(t, err)
	exists, err = val.One(ctx, &vaasInfo)
	require.NoError(t, err)
	require.True(t, exists)
	require.Equal(t, vaastav.UserID, vaasInfo.UserID)
	require.Len(t, vaasInfo.Followees, 2)
	require.Len(t, vaasInfo.Followers, 0)

	val, err = coll.FindOne(ctx, bson.D{{"userid", jcmace.UserID}})
	require.NoError(t, err)
	exists, err = val.One(ctx, &jcInfo)
	require.NoError(t, err)
	require.True(t, exists)
	require.Equal(t, jcmace.UserID, jcInfo.UserID)
	require.Len(t, jcInfo.Followees, 0)
	require.Len(t, jcInfo.Followers, 1)

	val, err = coll.FindOne(ctx, bson.D{{"userid", antoinek.UserID}})
	require.NoError(t, err)
	exists, err = val.One(ctx, &akInfo)
	require.NoError(t, err)
	require.True(t, exists)
	require.Equal(t, antoinek.UserID, akInfo.UserID)
	require.Len(t, akInfo.Followees, 0)
	require.Len(t, akInfo.Followers, 1)

	// Cleanup database

	err = coll.DeleteOne(ctx, bson.D{{"userid", vaastav.UserID}})
	require.NoError(t, err)
	err = coll.DeleteOne(ctx, bson.D{{"userid", jcmace.UserID}})
	require.NoError(t, err)
	err = coll.DeleteOne(ctx, bson.D{{"userid", antoinek.UserID}})
	require.NoError(t, err)
}

func TestFollowUsername(t *testing.T) {
	ctx := context.Background()
	service, err := socialGraphServiceRegistry.Get(ctx)
	require.NoError(t, err)

	db, err := socialGraphDBRegistry.Get(ctx)
	require.NoError(t, err)
	coll, err := db.GetCollection(ctx, "social-graph", "social-graph")
	require.NoError(t, err)

	cache, err := socialGraphCacheRegistry.Get(ctx)
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

	// Check cache contents
	vaastavID := strconv.FormatInt(vaastav.UserID, 10)
	jcmaceID := strconv.FormatInt(jcmace.UserID, 10)
	antoinekID := strconv.FormatInt(antoinek.UserID, 10)
	var followeeinfo []socialnetwork.FolloweeInfo
	var followerinfo []socialnetwork.FollowerInfo
	exists, err := cache.Get(ctx, vaastavID+":followees", &followeeinfo)
	require.NoError(t, err)
	require.True(t, exists)
	require.Len(t, followeeinfo, 1)
	exists, err = cache.Get(ctx, jcmaceID+":followers", &followerinfo)
	require.NoError(t, err)
	require.True(t, exists)
	require.Len(t, followerinfo, 1)

	// Add another follow to test multiple follows
	err = service.FollowWithUsername(ctx, 1004, vaastav.Username, antoinek.Username)
	require.NoError(t, err)

	// Check cache contents
	exists, err = cache.Get(ctx, vaastavID+":followees", &followeeinfo)
	require.NoError(t, err)
	require.True(t, exists)
	require.Len(t, followeeinfo, 2)
	exists, err = cache.Get(ctx, antoinekID+":followers", &followerinfo)
	require.NoError(t, err)
	require.True(t, exists)
	require.Len(t, followerinfo, 1)

	// Cleanup cache
	err = cache.Delete(ctx, vaastavID+":followees")
	require.NoError(t, err)
	err = cache.Delete(ctx, jcmaceID+":followers")
	require.NoError(t, err)
	err = cache.Delete(ctx, antoinekID+":followers")

	// Check database contents
	var vaasInfo socialnetwork.UserInfo
	var jcInfo socialnetwork.UserInfo
	var akInfo socialnetwork.UserInfo

	val, err := coll.FindOne(ctx, bson.D{{"userid", vaastav.UserID}})
	require.NoError(t, err)
	exists, err = val.One(ctx, &vaasInfo)
	require.NoError(t, err)
	require.True(t, exists)
	require.Equal(t, vaastav.UserID, vaasInfo.UserID)
	require.Len(t, vaasInfo.Followees, 2)
	require.Len(t, vaasInfo.Followers, 0)

	val, err = coll.FindOne(ctx, bson.D{{"userid", jcmace.UserID}})
	require.NoError(t, err)
	exists, err = val.One(ctx, &jcInfo)
	require.NoError(t, err)
	require.True(t, exists)
	require.Equal(t, jcmace.UserID, jcInfo.UserID)
	require.Len(t, jcInfo.Followees, 0)
	require.Len(t, jcInfo.Followers, 1)

	val, err = coll.FindOne(ctx, bson.D{{"userid", antoinek.UserID}})
	require.NoError(t, err)
	exists, err = val.One(ctx, &akInfo)
	require.NoError(t, err)
	require.True(t, exists)
	require.Equal(t, antoinek.UserID, akInfo.UserID)
	require.Len(t, akInfo.Followees, 0)
	require.Len(t, akInfo.Followers, 1)

	// Cleanup databases

	err = coll.DeleteOne(ctx, bson.D{{"userid", vaastav.UserID}})
	require.NoError(t, err)
	err = coll.DeleteOne(ctx, bson.D{{"userid", jcmace.UserID}})
	require.NoError(t, err)
	err = coll.DeleteOne(ctx, bson.D{{"userid", antoinek.UserID}})
	require.NoError(t, err)

	err = user_coll.DeleteMany(ctx, bson.D{})
	require.NoError(t, err)
}

func TestGetFollowers(t *testing.T) {
	ctx := context.Background()
	service, err := socialGraphServiceRegistry.Get(ctx)
	require.NoError(t, err)

	db, err := socialGraphDBRegistry.Get(ctx)
	require.NoError(t, err)
	coll, err := db.GetCollection(ctx, "social-graph", "social-graph")
	require.NoError(t, err)

	cache, err := socialGraphCacheRegistry.Get(ctx)
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

	vaastavID := strconv.FormatInt(vaastav.UserID, 10)
	jcmaceID := strconv.FormatInt(jcmace.UserID, 10)
	err = cache.Delete(ctx, vaastavID+":followees")
	require.NoError(t, err)
	err = cache.Delete(ctx, jcmaceID+":followers")
	require.NoError(t, err)

	// Cleanup database
	err = coll.DeleteMany(ctx, bson.D{})
	require.NoError(t, err)
}

func TestGetFollowees(t *testing.T) {
	ctx := context.Background()
	service, err := socialGraphServiceRegistry.Get(ctx)
	require.NoError(t, err)

	db, err := socialGraphDBRegistry.Get(ctx)
	require.NoError(t, err)
	coll, err := db.GetCollection(ctx, "social-graph", "social-graph")
	require.NoError(t, err)

	cache, err := socialGraphCacheRegistry.Get(ctx)
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

	vaastavID := strconv.FormatInt(vaastav.UserID, 10)
	jcmaceID := strconv.FormatInt(jcmace.UserID, 10)
	err = cache.Delete(ctx, vaastavID+":followees")
	require.NoError(t, err)
	err = cache.Delete(ctx, jcmaceID+":followers")
	require.NoError(t, err)

	// Cleanup database
	err = coll.DeleteMany(ctx, bson.D{})
	require.NoError(t, err)
}

func TestUnfollow(t *testing.T) {
	ctx := context.Background()
	service, err := socialGraphServiceRegistry.Get(ctx)
	require.NoError(t, err)

	db, err := socialGraphDBRegistry.Get(ctx)
	require.NoError(t, err)
	coll, err := db.GetCollection(ctx, "social-graph", "social-graph")
	require.NoError(t, err)

	cache, err := socialGraphCacheRegistry.Get(ctx)
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

	// Check cache contents
	vaastavID := strconv.FormatInt(vaastav.UserID, 10)
	jcmaceID := strconv.FormatInt(jcmace.UserID, 10)
	var followeeinfo []socialnetwork.FolloweeInfo
	var followerinfo []socialnetwork.FollowerInfo
	exists, err := cache.Get(ctx, vaastavID+":followees", &followeeinfo)
	require.NoError(t, err)
	require.True(t, exists)
	require.Len(t, followeeinfo, 0)
	exists, err = cache.Get(ctx, jcmaceID+":followers", &followerinfo)
	require.NoError(t, err)
	require.True(t, exists)
	require.Len(t, followerinfo, 0)

	// Test Unfollow
	var vaasInfo socialnetwork.UserInfo
	var jcInfo socialnetwork.UserInfo

	val, err := coll.FindOne(ctx, bson.D{{"userid", vaastav.UserID}})
	require.NoError(t, err)
	exists, err = val.One(ctx, &vaasInfo)
	require.NoError(t, err)
	require.True(t, exists)
	require.Equal(t, vaastav.UserID, vaasInfo.UserID)
	// The following check fails because even tho the element is unset, its not removed. Bug in simplenosqldb.
	require.Len(t, vaasInfo.Followees, 0)
	require.Len(t, vaasInfo.Followers, 0)

	val, err = coll.FindOne(ctx, bson.D{{"userid", jcmace.UserID}})
	require.NoError(t, err)
	exists, err = val.One(ctx, &jcInfo)
	require.NoError(t, err)
	require.True(t, exists)
	require.Equal(t, jcmace.UserID, jcInfo.UserID)
	require.Len(t, jcInfo.Followees, 0)
	require.Len(t, jcInfo.Followers, 0)

	// Clean up cache
	err = cache.Delete(ctx, vaastavID+":followees")
	require.NoError(t, err)
	err = cache.Delete(ctx, jcmaceID+":followers")
	require.NoError(t, err)

	// Clean up database
	err = coll.DeleteOne(ctx, bson.D{{"userid", vaastav.UserID}})
	require.NoError(t, err)
	err = coll.DeleteOne(ctx, bson.D{{"userid", jcmace.UserID}})
	require.NoError(t, err)
}
