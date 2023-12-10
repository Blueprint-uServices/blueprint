package tests

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"gitlab.mpi-sws.org/cld/blueprint/examples/dsb_sn/workflow/socialnetwork"
	"gitlab.mpi-sws.org/cld/blueprint/runtime/core/registry"
	"go.mongodb.org/mongo-driver/bson"
)

var composePostServiceRegistry = registry.NewServiceRegistry[socialnetwork.ComposePostService]("composepost_service")

func init() {

	composePostServiceRegistry.Register("local", func(ctx context.Context) (socialnetwork.ComposePostService, error) {
		postStorageService, err := postStorageServiceRegistry.Get(ctx)
		if err != nil {
			return nil, err
		}

		userTimelineService, err := userTimelineServiceRegistry.Get(ctx)
		if err != nil {
			return nil, err
		}

		userService, err := userServiceRegistry.Get(ctx)
		if err != nil {
			return nil, err
		}

		uniqueIdService, err := uniqueIdServiceRegistry.Get(ctx)
		if err != nil {
			return nil, err
		}

		mediaService, err := mediaServiceRegistry.Get(ctx)
		if err != nil {
			return nil, err
		}

		textService, err := textServiceRegistry.Get(ctx)
		if err != nil {
			return nil, err
		}

		hometimelineService, err := homeTimelineServiceRegistry.Get(ctx)
		if err != nil {
			return nil, err
		}

		return socialnetwork.NewComposePostServiceImpl(ctx, postStorageService, userTimelineService, userService, uniqueIdService, mediaService, textService, hometimelineService)
	})
}

func TestComposePost(t *testing.T) {
	ctx := context.Background()
	service, err := composePostServiceRegistry.Get(ctx)
	require.NoError(t, err)

	load_users(t, ctx)

	var mediaids []int64
	var mediatypes []string

	for _, media := range post1.Medias {
		mediaids = append(mediaids, media.MediaID)
		mediatypes = append(mediatypes, media.MediaType)
	}

	id, mentions, err := service.ComposePost(ctx, 1, post1.Creator.Username, post1.Creator.UserID, post1.Text, mediaids, mediatypes, socialnetwork.POST)
	require.NoError(t, err)
	require.True(t, id > 0)
	require.Len(t, mentions, len(post1.UserMentions))

	cleanup_dbs(t, ctx)
	cleanup_post_backends(t, ctx, []int64{id})
}

func load_users(t *testing.T, ctx context.Context) {
	service, err := userServiceRegistry.Get(ctx)
	require.NoError(t, err)

	// Add users
	req_id := 1000
	users := []socialnetwork.User{vaastav, antoinek, jcmace, dg}
	for _, user := range users {
		req_id += 1
		err = service.RegisterUserWithId(ctx, int64(req_id), user.FirstName, user.LastName, user.Username, "vaaspwd", user.UserID)
		require.NoError(t, err)
	}
}

func cleanup_dbs(t *testing.T, ctx context.Context) {
	db, err := userDBRegistry.Get(ctx)
	require.NoError(t, err)
	coll, err := db.GetCollection(ctx, "user", "user")
	require.NoError(t, err)

	err = coll.DeleteMany(ctx, bson.D{})
	require.NoError(t, err)
}
