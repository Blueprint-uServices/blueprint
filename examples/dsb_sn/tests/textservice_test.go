package tests

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"gitlab.mpi-sws.org/cld/blueprint/examples/dsb_sn/workflow/socialnetwork"
	"gitlab.mpi-sws.org/cld/blueprint/runtime/core/registry"
	"go.mongodb.org/mongo-driver/bson"
)

var textServiceRegistry = registry.NewServiceRegistry[socialnetwork.TextService]("text_service")

func init() {
	textServiceRegistry.Register("local", func(ctx context.Context) (socialnetwork.TextService, error) {
		urlShortenService, err := urlShortenServiceRegistry.Get(ctx)
		if err != nil {
			return nil, err
		}

		userMentionService, err := userMentionServiceRegistry.Get(ctx)
		if err != nil {
			return nil, err
		}

		return socialnetwork.NewTextServiceImpl(ctx, urlShortenService, userMentionService)
	})
}

func TestComposeText(t *testing.T) {
	ctx := context.Background()
	service, err := textServiceRegistry.Get(ctx)
	require.NoError(t, err)

	raw_text := "Hello World!"
	username_text := "@vaastav, @jcmace, @antoinek, @dg"
	links_text := "http://blueprint-uservices.github.io"
	full_text := raw_text + " Check out Blueprint(" + links_text + ") by " + username_text + " !"

	// Test only raw_text
	updated_text, mentions, urls, err := service.ComposeText(ctx, 1000, raw_text)
	require.NoError(t, err)
	require.Equal(t, raw_text, updated_text)
	require.Len(t, mentions, 0)
	require.Len(t, urls, 0)

	// Add users to the data before running the test as usermention requires that mentioned users be valid
	db, err := userDBRegistry.Get(ctx)
	require.NoError(t, err)
	coll, err := db.GetCollection(ctx, "user", "user")
	require.NoError(t, err)
	err = coll.InsertMany(ctx, []interface{}{vaastav, jcmace, antoinek, dg})
	require.NoError(t, err)

	// Test full_text
	updated_text, mentions, urls, err = service.ComposeText(ctx, 1000, full_text)
	require.NoError(t, err)
	require.Len(t, urls, 1)
	require.Len(t, mentions, 4)

	// cleanup database
	err = coll.DeleteMany(ctx, bson.D{{"userid", vaastav.UserID}})
	require.NoError(t, err)
	err = coll.DeleteMany(ctx, bson.D{{"userid", jcmace.UserID}})
	require.NoError(t, err)
	err = coll.DeleteMany(ctx, bson.D{{"userid", antoinek.UserID}})
	require.NoError(t, err)
	err = coll.DeleteMany(ctx, bson.D{{"userid", dg.UserID}})
	require.NoError(t, err)
}
