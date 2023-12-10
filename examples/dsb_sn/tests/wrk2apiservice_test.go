package tests

import (
	"context"
	"log"
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"
	"gitlab.mpi-sws.org/cld/blueprint/examples/dsb_sn/workflow/socialnetwork"
	"gitlab.mpi-sws.org/cld/blueprint/runtime/core/registry"
	"go.mongodb.org/mongo-driver/bson"
)

var wrk2apiServiceRegistry = registry.NewServiceRegistry[socialnetwork.Wrk2APIService]("wrk2api_service")

func init() {
	wrk2apiServiceRegistry.Register("local", func(ctx context.Context) (socialnetwork.Wrk2APIService, error) {
		userService, err := userServiceRegistry.Get(ctx)
		if err != nil {
			return nil, err
		}

		composePostService, err := composePostServiceRegistry.Get(ctx)
		if err != nil {
			return nil, err
		}

		userTimelineService, err := userTimelineServiceRegistry.Get(ctx)
		if err != nil {
			return nil, err
		}

		homeTimelineService, err := homeTimelineServiceRegistry.Get(ctx)
		if err != nil {
			return nil, err
		}

		socialGraphService, err := socialGraphServiceRegistry.Get(ctx)
		if err != nil {
			return nil, err
		}

		return socialnetwork.NewWrk2APIServiceImpl(ctx, userService, composePostService, userTimelineService, homeTimelineService, socialGraphService)
	})
}

func TestRegister(t *testing.T) {
	ctx := context.Background()
	service, err := wrk2apiServiceRegistry.Get(ctx)
	require.NoError(t, err)

	// Test register with incomplete arguments
	err = service.Register(ctx, "", vaastav.LastName, vaastav.Username, "pwd", vaastav.UserID)
	require.Error(t, err)

	err = service.Register(ctx, vaastav.FirstName, "", vaastav.Username, "pwd", vaastav.UserID)
	require.Error(t, err)

	err = service.Register(ctx, vaastav.FirstName, vaastav.LastName, "", "pwd", vaastav.UserID)
	require.Error(t, err)

	err = service.Register(ctx, vaastav.FirstName, vaastav.LastName, vaastav.Username, "", vaastav.UserID)
	require.Error(t, err)

	// Test register with complete arguments

	err = service.Register(ctx, vaastav.FirstName, vaastav.LastName, vaastav.Username, "vaaspwd", vaastav.UserID)
	require.NoError(t, err)

	err = service.Register(ctx, jcmace.FirstName, jcmace.LastName, jcmace.Username, "jonpwd", jcmace.UserID)
	require.NoError(t, err)

	// Test duplicate register
	err = service.Register(ctx, vaastav.FirstName, vaastav.LastName, vaastav.Username, "vaaspwd", vaastav.UserID)
	require.Error(t, err)

	// Cleanup database
	user_db, err := userDBRegistry.Get(ctx)
	require.NoError(t, err)
	coll, err := user_db.GetCollection(ctx, "user", "user")
	require.NoError(t, err)

	err = coll.DeleteMany(ctx, bson.D{})
	require.NoError(t, err)

	social_db, err := socialGraphDBRegistry.Get(ctx)
	require.NoError(t, err)
	soc_coll, err := social_db.GetCollection(ctx, "social-graph", "social-graph")
	require.NoError(t, err)

	err = soc_coll.DeleteMany(ctx, bson.D{})
	require.NoError(t, err)
}

func TestWrk2Follow(t *testing.T) {
	ctx := context.Background()
	service, err := wrk2apiServiceRegistry.Get(ctx)
	require.NoError(t, err)

	// Register some users
	users := []socialnetwork.User{vaastav, antoinek, jcmace, dg}
	for _, user := range users {
		err = service.Register(ctx, user.FirstName, user.LastName, user.Username, "vaaspwd", user.UserID)
		require.NoError(t, err)
	}

	// Test Follow with UserIDs

	err = service.Follow(ctx, "", "", vaastav.UserID, jcmace.UserID)
	require.NoError(t, err)
	err = service.Follow(ctx, "", "", vaastav.UserID, antoinek.UserID)
	require.NoError(t, err)
	err = service.Follow(ctx, "", "", vaastav.UserID, dg.UserID)

	// Test Follow with Usernames
	err = service.Follow(ctx, jcmace.Username, vaastav.Username, 0, 0)
	require.NoError(t, err)
	err = service.Follow(ctx, antoinek.Username, vaastav.Username, 0, 0)
	require.NoError(t, err)
	err = service.Follow(ctx, dg.Username, vaastav.Username, 0, 0)
	require.NoError(t, err)

	// Cleanup caches
	soc_cache, err := socialGraphCacheRegistry.Get(ctx)
	require.NoError(t, err)
	for _, user := range users {
		idstr := strconv.FormatInt(user.UserID, 10)
		soc_cache.Delete(ctx, idstr+":followers")
		soc_cache.Delete(ctx, idstr+":followees")
	}

	// Cleanup databases
	user_db, err := userDBRegistry.Get(ctx)
	require.NoError(t, err)
	coll, err := user_db.GetCollection(ctx, "user", "user")
	require.NoError(t, err)

	err = coll.DeleteMany(ctx, bson.D{})
	require.NoError(t, err)

	social_db, err := socialGraphDBRegistry.Get(ctx)
	require.NoError(t, err)
	soc_coll, err := social_db.GetCollection(ctx, "social-graph", "social-graph")
	require.NoError(t, err)

	err = soc_coll.DeleteMany(ctx, bson.D{})
	require.NoError(t, err)
}

func TestWrk2Unfollow(t *testing.T) {
	ctx := context.Background()
	service, err := wrk2apiServiceRegistry.Get(ctx)
	require.NoError(t, err)

	// Register some users
	users := []socialnetwork.User{vaastav, antoinek, jcmace, dg}
	for _, user := range users {
		err = service.Register(ctx, user.FirstName, user.LastName, user.Username, "vaaspwd", user.UserID)
		require.NoError(t, err)
	}

	// Follow some users

	err = service.Follow(ctx, "", "", vaastav.UserID, jcmace.UserID)
	require.NoError(t, err)
	err = service.Follow(ctx, "", "", vaastav.UserID, antoinek.UserID)
	require.NoError(t, err)
	err = service.Follow(ctx, "", "", vaastav.UserID, dg.UserID)

	err = service.Follow(ctx, jcmace.Username, vaastav.Username, 0, 0)
	require.NoError(t, err)
	err = service.Follow(ctx, antoinek.Username, vaastav.Username, 0, 0)
	require.NoError(t, err)
	err = service.Follow(ctx, dg.Username, vaastav.Username, 0, 0)
	require.NoError(t, err)

	// Test Unfollow with just userID
	err = service.Unfollow(ctx, "", "", dg.UserID, vaastav.UserID)
	require.NoError(t, err)
	err = service.Unfollow(ctx, "", "", antoinek.UserID, vaastav.UserID)
	require.NoError(t, err)
	err = service.Unfollow(ctx, "", "", jcmace.UserID, vaastav.UserID)

	// Test Unfollow with usernames
	err = service.Unfollow(ctx, vaastav.Username, jcmace.Username, 0, 0)
	require.NoError(t, err)
	err = service.Unfollow(ctx, vaastav.Username, antoinek.Username, 0, 0)
	require.NoError(t, err)
	err = service.Unfollow(ctx, vaastav.Username, dg.Username, 0, 0)
	require.NoError(t, err)

	// Cleanup caches
	soc_cache, err := socialGraphCacheRegistry.Get(ctx)
	require.NoError(t, err)
	for _, user := range users {
		idstr := strconv.FormatInt(user.UserID, 10)
		soc_cache.Delete(ctx, idstr+":followers")
		soc_cache.Delete(ctx, idstr+":followees")
	}

	// Cleanup databases
	user_db, err := userDBRegistry.Get(ctx)
	require.NoError(t, err)
	coll, err := user_db.GetCollection(ctx, "user", "user")
	require.NoError(t, err)

	err = coll.DeleteMany(ctx, bson.D{})
	require.NoError(t, err)

	social_db, err := socialGraphDBRegistry.Get(ctx)
	require.NoError(t, err)
	soc_coll, err := social_db.GetCollection(ctx, "social-graph", "social-graph")
	require.NoError(t, err)

	err = soc_coll.DeleteMany(ctx, bson.D{})
	require.NoError(t, err)
}

func TestWrk2Compose(t *testing.T) {
	ctx := context.Background()
	service, err := wrk2apiServiceRegistry.Get(ctx)
	require.NoError(t, err)

	var all_ids []int64

	defer func() {
		cleanup_utimeline_db(t, ctx)
		cleanup_post_backends(t, ctx, all_ids)
	}()

	// Register some users
	users := []socialnetwork.User{vaastav, antoinek, jcmace, dg}
	for _, user := range users {
		err = service.Register(ctx, user.FirstName, user.LastName, user.Username, "vaaspwd", user.UserID)
		require.NoError(t, err)
	}

	// Test Compose with some incomplete arguments!
	var mediaids []int64
	var mediatypes []string

	for _, media := range post1.Medias {
		mediaids = append(mediaids, media.MediaID)
		mediatypes = append(mediatypes, media.MediaType)
	}

	id, mentions, err := service.ComposePost(ctx, 0, vaastav.Username, post1.PostType, post1.Text, mediatypes, mediaids)
	require.Error(t, err)
	require.Equal(t, int64(-1), id)
	require.Len(t, mentions, 0)

	id, mentions, err = service.ComposePost(ctx, vaastav.UserID, "", post1.PostType, post1.Text, mediatypes, mediaids)
	require.Error(t, err)
	require.Equal(t, int64(-1), id)
	require.Len(t, mentions, 0)

	id, mentions, err = service.ComposePost(ctx, vaastav.UserID, vaastav.Username, post1.PostType, "", mediatypes, mediaids)
	require.Error(t, err)
	require.Equal(t, int64(-1), id)
	require.Len(t, mentions, 0)

	// Test compose with complete arguments
	id, mentions, err = service.ComposePost(ctx, vaastav.UserID, vaastav.Username, post1.PostType, post1.Text, mediatypes, mediaids)
	require.NoError(t, err)
	require.True(t, id > 0)
	require.Len(t, mentions, len(post1.UserMentions))
	all_ids = append(all_ids, id)

	// Cleanup databases
	user_db, err := userDBRegistry.Get(ctx)
	require.NoError(t, err)
	coll, err := user_db.GetCollection(ctx, "user", "user")
	require.NoError(t, err)

	err = coll.DeleteMany(ctx, bson.D{})
	require.NoError(t, err)

	social_db, err := socialGraphDBRegistry.Get(ctx)
	require.NoError(t, err)
	soc_coll, err := social_db.GetCollection(ctx, "social-graph", "social-graph")
	require.NoError(t, err)

	err = soc_coll.DeleteMany(ctx, bson.D{})
	require.NoError(t, err)
}

func cleanup_utimeline_db(t *testing.T, ctx context.Context) {
	db, err := userTimelineDBRegistry.Get(ctx)
	require.NoError(t, err)
	coll, err := db.GetCollection(ctx, "usertimeline", "usertimeline")
	require.NoError(t, err)
	// Cleanup database
	log.Println("Cleaning up database")
	err = coll.DeleteMany(ctx, bson.D{})
	require.NoError(t, err)
}

func TestWrk2ReadTimelines(t *testing.T) {
	ctx := context.Background()
	service, err := wrk2apiServiceRegistry.Get(ctx)
	require.NoError(t, err)

	// Register some users
	users := []socialnetwork.User{vaastav, antoinek, jcmace, dg}
	for _, user := range users {
		err = service.Register(ctx, user.FirstName, user.LastName, user.Username, "vaaspwd", user.UserID)
		require.NoError(t, err)
	}

	var all_ids []int64

	defer func() {
		cleanup_utimeline_db(t, ctx)
		cleanup_post_backends(t, ctx, all_ids)
	}()

	// Test Compose with some incomplete arguments!
	var mediaids []int64
	var mediatypes []string

	for _, media := range post1.Medias {
		mediaids = append(mediaids, media.MediaID)
		mediatypes = append(mediatypes, media.MediaType)
	}

	id, mentions, err := service.ComposePost(ctx, vaastav.UserID, vaastav.Username, post1.PostType, post1.Text, mediatypes, mediaids)
	require.NoError(t, err)
	require.True(t, id > 0)
	require.Len(t, mentions, len(post1.UserMentions))
	all_ids = append(all_ids, id)

	for _, user := range users {
		if user.UserID == vaastav.UserID {
			continue
		}
		posts, err := service.ReadHomeTimeline(ctx, user.UserID, 0, 5)
		require.NoError(t, err)
		require.Len(t, posts, 1)
	}

	// Antoine has no posts so there should be an error here
	posts, err := service.ReadUserTimeline(ctx, antoinek.UserID, 0, 5)
	require.Error(t, err)

	posts, err = service.ReadUserTimeline(ctx, vaastav.UserID, 0, 5)
	require.NoError(t, err)
	require.Len(t, posts, 1)

	// Cleanup databases
	user_db, err := userDBRegistry.Get(ctx)
	require.NoError(t, err)
	coll, err := user_db.GetCollection(ctx, "user", "user")
	require.NoError(t, err)

	err = coll.DeleteMany(ctx, bson.D{})
	require.NoError(t, err)

	social_db, err := socialGraphDBRegistry.Get(ctx)
	require.NoError(t, err)
	soc_coll, err := social_db.GetCollection(ctx, "social-graph", "social-graph")
	require.NoError(t, err)

	err = soc_coll.DeleteMany(ctx, bson.D{})
	require.NoError(t, err)

}
