package tests

import (
	"context"
	"testing"

	"github.com/blueprint-uservices/blueprint/examples/dsb_mm/workflow/media"
	"github.com/blueprint-uservices/blueprint/runtime/core/backend"
	"github.com/blueprint-uservices/blueprint/runtime/core/registry"
	"github.com/blueprint-uservices/blueprint/runtime/plugins/simplecache"
	"github.com/blueprint-uservices/blueprint/runtime/plugins/simplenosqldb"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	castInfoServiceRegistry      = registry.NewServiceRegistry[media.CastInfoService]("cast_info_service")
	composeReviewServiceRegistry = registry.NewServiceRegistry[media.ComposeReviewService]("compose_review_service")
	movieIDServiceRegistry       = registry.NewServiceRegistry[media.MovieIdService]("movie_id_service")
	movieInfoServiceRegistry     = registry.NewServiceRegistry[media.MovieInfoService]("movie_info_service")
	movieReviewServiceRegistry   = registry.NewServiceRegistry[media.MovieReviewService]("movie_review_service")
	pageServiceRegistry          = registry.NewServiceRegistry[media.PageService]("page_service")
	plotServiceRegistry          = registry.NewServiceRegistry[media.PlotService]("plot_service")
	ratingServiceRegistry        = registry.NewServiceRegistry[media.RatingService]("rating_service")
	reviewStorageServiceRegistry = registry.NewServiceRegistry[media.ReviewStorageService]("review_storage_service")
	textServiceRegistry          = registry.NewServiceRegistry[media.TextService]("text_service")
	uniqueIDServiceRegistry      = registry.NewServiceRegistry[media.UniqueIdService]("unique_id_service")
	userReviewServiceRegistry    = registry.NewServiceRegistry[media.UserReviewService]("user_review_service")
	userServiceRegistry          = registry.NewServiceRegistry[media.UserService]("user_service")
	wrk2APIServiceRegistry       = registry.NewServiceRegistry[media.Wrk2APIService]("wrk2_api_service")
)

func newCache(ctx context.Context) (backend.Cache, error) {
	return simplecache.NewSimpleCache(ctx)
}

func newDB(ctx context.Context) (backend.NoSQLDatabase, error) {
	return simplenosqldb.NewSimpleNoSQLDB(ctx)
}

func init() {
	reviewStorageServiceRegistry.Register("local", func(ctx context.Context) (media.ReviewStorageService, error) {
		cache, err := newCache(ctx)
		if err != nil {
			return nil, err
		}
		db, err := newDB(ctx)
		if err != nil {
			return nil, err
		}
		return media.NewReviewStorageServiceImpl(ctx, cache, db)
	})

	userReviewServiceRegistry.Register("local", func(ctx context.Context) (media.UserReviewService, error) {
		reviewStorage, err := reviewStorageServiceRegistry.Get(ctx)
		if err != nil {
			return nil, err
		}
		db, err := newDB(ctx)
		if err != nil {
			return nil, err
		}
		cache, err := newCache(ctx)
		if err != nil {
			return nil, err
		}
		return media.NewUserReviewServiceImpl(ctx, reviewStorage, db, cache)
	})

	movieReviewServiceRegistry.Register("local", func(ctx context.Context) (media.MovieReviewService, error) {
		reviewStorage, err := reviewStorageServiceRegistry.Get(ctx)
		if err != nil {
			return nil, err
		}
		cache, err := newCache(ctx)
		if err != nil {
			return nil, err
		}
		db, err := newDB(ctx)
		if err != nil {
			return nil, err
		}
		return media.NewMovieReviewServiceImpl(ctx, reviewStorage, cache, db)
	})

	composeReviewServiceRegistry.Register("local", func(ctx context.Context) (media.ComposeReviewService, error) {
		reviewStorage, err := reviewStorageServiceRegistry.Get(ctx)
		if err != nil {
			return nil, err
		}
		userReview, err := userReviewServiceRegistry.Get(ctx)
		if err != nil {
			return nil, err
		}
		movieReview, err := movieReviewServiceRegistry.Get(ctx)
		if err != nil {
			return nil, err
		}
		cache, err := newCache(ctx)
		if err != nil {
			return nil, err
		}
		return media.NewComposeReviewServiceImpl(ctx, cache, reviewStorage, userReview, movieReview)
	})

	ratingServiceRegistry.Register("local", func(ctx context.Context) (media.RatingService, error) {
		composeReview, err := composeReviewServiceRegistry.Get(ctx)
		if err != nil {
			return nil, err
		}
		cache, err := newCache(ctx)
		if err != nil {
			return nil, err
		}
		return media.NewRatingServiceImpl(ctx, composeReview, cache)
	})

	movieIDServiceRegistry.Register("local", func(ctx context.Context) (media.MovieIdService, error) {
		rating, err := ratingServiceRegistry.Get(ctx)
		if err != nil {
			return nil, err
		}
		composeReview, err := composeReviewServiceRegistry.Get(ctx)
		if err != nil {
			return nil, err
		}
		cache, err := newCache(ctx)
		if err != nil {
			return nil, err
		}
		db, err := newDB(ctx)
		if err != nil {
			return nil, err
		}
		return media.NewMovieIdServiceImpl(ctx, cache, db, rating, composeReview)
	})

	userServiceRegistry.Register("local", func(ctx context.Context) (media.UserService, error) {
		composeReview, err := composeReviewServiceRegistry.Get(ctx)
		if err != nil {
			return nil, err
		}
		cache, err := newCache(ctx)
		if err != nil {
			return nil, err
		}
		db, err := newDB(ctx)
		if err != nil {
			return nil, err
		}
		return media.NewUserServiceImpl(ctx, cache, db, composeReview, "test-secret")
	})

	castInfoServiceRegistry.Register("local", func(ctx context.Context) (media.CastInfoService, error) {
		cache, err := newCache(ctx)
		if err != nil {
			return nil, err
		}
		db, err := newDB(ctx)
		if err != nil {
			return nil, err
		}
		return media.NewCastInfoService(ctx, cache, db)
	})

	plotServiceRegistry.Register("local", func(ctx context.Context) (media.PlotService, error) {
		cache, err := newCache(ctx)
		if err != nil {
			return nil, err
		}
		db, err := newDB(ctx)
		if err != nil {
			return nil, err
		}
		return media.NewPlotServiceImpl(ctx, cache, db)
	})

	movieInfoServiceRegistry.Register("local", func(ctx context.Context) (media.MovieInfoService, error) {
		cache, err := newCache(ctx)
		if err != nil {
			return nil, err
		}
		db, err := newDB(ctx)
		if err != nil {
			return nil, err
		}
		return media.NewMovieInfoServiceImpl(ctx, cache, db)
	})

	textServiceRegistry.Register("local", func(ctx context.Context) (media.TextService, error) {
		composeReview, err := composeReviewServiceRegistry.Get(ctx)
		if err != nil {
			return nil, err
		}
		return media.NewTextServiceImpl(ctx, composeReview)
	})

	uniqueIDServiceRegistry.Register("local", func(ctx context.Context) (media.UniqueIdService, error) {
		composeReview, err := composeReviewServiceRegistry.Get(ctx)
		if err != nil {
			return nil, err
		}
		return media.NewUniqueIdServiceImpl(ctx, composeReview)
	})

	pageServiceRegistry.Register("local", func(ctx context.Context) (media.PageService, error) {
		movieInfo, err := movieInfoServiceRegistry.Get(ctx)
		if err != nil {
			return nil, err
		}
		movieReview, err := movieReviewServiceRegistry.Get(ctx)
		if err != nil {
			return nil, err
		}
		castInfo, err := castInfoServiceRegistry.Get(ctx)
		if err != nil {
			return nil, err
		}
		plot, err := plotServiceRegistry.Get(ctx)
		if err != nil {
			return nil, err
		}
		return media.NewPageServiceImpl(ctx, movieInfo, movieReview, castInfo, plot)
	})

	wrk2APIServiceRegistry.Register("local", func(ctx context.Context) (media.Wrk2APIService, error) {
		user, err := userServiceRegistry.Get(ctx)
		if err != nil {
			return nil, err
		}
		castInfo, err := castInfoServiceRegistry.Get(ctx)
		if err != nil {
			return nil, err
		}
		text, err := textServiceRegistry.Get(ctx)
		if err != nil {
			return nil, err
		}
		plot, err := plotServiceRegistry.Get(ctx)
		if err != nil {
			return nil, err
		}
		movieID, err := movieIDServiceRegistry.Get(ctx)
		if err != nil {
			return nil, err
		}
		movieInfo, err := movieInfoServiceRegistry.Get(ctx)
		if err != nil {
			return nil, err
		}
		uniqueID, err := uniqueIDServiceRegistry.Get(ctx)
		if err != nil {
			return nil, err
		}
		return media.NewWrk2APIServiceImpl(ctx, user, castInfo, text, plot, movieID, movieInfo, uniqueID)
	})
}

func TestMediaStorageServices(t *testing.T) {
	ctx := context.Background()
	castInfoService, err := castInfoServiceRegistry.Get(ctx)
	require.NoError(t, err)
	plotService, err := plotServiceRegistry.Get(ctx)
	require.NoError(t, err)
	movieInfoService, err := movieInfoServiceRegistry.Get(ctx)
	require.NoError(t, err)
	reviewStorageService, err := reviewStorageServiceRegistry.Get(ctx)
	require.NoError(t, err)

	require.NoError(t, castInfoService.WriteCastInfo(ctx, 1, 101, "Ada Actor", true, "Lead actor"))
	cast, err := castInfoService.ReadCastInfo(ctx, 2, []int{101})
	require.NoError(t, err)
	assert.Equal(t, []media.CastInfo{{CastInfoId: 101, Name: "Ada Actor", Gender: true, Intro: "Lead actor"}}, cast)

	require.NoError(t, plotService.WritePlot(ctx, 3, 7, "A test movie plot"))
	plot, err := plotService.ReadPlot(ctx, 4, 7)
	require.NoError(t, err)
	assert.Equal(t, "A test movie plot", plot)

	info := media.MovieInfo{
		MovieID: "movie-1", Title: "Test Movie", Casts: []media.Cast{{CastID: 1, Character: "Lead", CastInfoID: 101}},
		PlotID: 7, ThumbnailIDs: []string{"thumb-1"}, PhotoIDs: []string{"photo-1"}, VideoIDs: []string{"video-1"},
		AvgRating: 4, NumRating: 2,
	}
	require.NoError(t, movieInfoService.WriteMovieInfo(ctx, 5, info.MovieID, info.Title, info.Casts, info.PlotID, info.ThumbnailIDs, info.PhotoIDs, info.VideoIDs, info.AvgRating, info.NumRating))
	storedInfo, err := movieInfoService.ReadMovieInfo(ctx, 6, info.MovieID)
	require.NoError(t, err)
	assert.Equal(t, info, storedInfo)
	require.NoError(t, movieInfoService.UpdateRating(ctx, 7, info.MovieID, 10, 2))
	storedInfo, err = movieInfoService.ReadMovieInfo(ctx, 8, info.MovieID)
	require.NoError(t, err)
	assert.Equal(t, 4, storedInfo.NumRating)
	assert.Equal(t, 4.5, storedInfo.AvgRating)

	review := media.Review{ReviewID: 88, UserID: 12, ReqID: 9, Text: "Great", MovieID: info.MovieID, Rating: 5, Timestamp: "now"}
	require.NoError(t, reviewStorageService.StoreReview(ctx, 9, review))
	reviews, err := reviewStorageService.ReadReviews(ctx, 10, []int{88})
	require.NoError(t, err)
	assert.Equal(t, []media.Review{review}, reviews)
}

func TestMediaReviewCompositionAndPage(t *testing.T) {
	ctx := context.Background()
	wrk2APIService, err := wrk2APIServiceRegistry.Get(ctx)
	require.NoError(t, err)
	userService, err := userServiceRegistry.Get(ctx)
	require.NoError(t, err)
	pageService, err := pageServiceRegistry.Get(ctx)
	require.NoError(t, err)
	userReviewService, err := userReviewServiceRegistry.Get(ctx)
	require.NoError(t, err)

	require.NoError(t, wrk2APIService.RegisterUser(ctx, "Grace", "Viewer", "grace", "secret"))
	userID, err := userService.GetUserId(ctx, 11, "grace")
	require.NoError(t, err)
	assert.NotZero(t, userID)
	token, err := userService.Login(ctx, 12, "grace", "secret")
	require.NoError(t, err)
	assert.NotEmpty(t, token)
	_, err = userService.Login(ctx, 13, "grace", "wrong")
	assert.Error(t, err)

	require.NoError(t, wrk2APIService.WriteCastInfo(ctx, 201, "Grace Actor", false, "Supporting actor"))
	require.NoError(t, wrk2APIService.WritePlot(ctx, 17, "An end-to-end plot"))
	require.NoError(t, wrk2APIService.RegisterMovie(ctx, "End-to-End Movie", "movie-e2e"))
	require.NoError(t, wrk2APIService.WriteMovieInfo(ctx, "movie-e2e", "End-to-End Movie", []media.Cast{{CastID: 2, Character: "Hero", CastInfoID: 201}}, 17, nil, nil, nil, 0, 0))
	require.NoError(t, wrk2APIService.ComposeReview(ctx, "End-to-End Movie", "Excellent movie", "grace", "secret", 8))

	page, err := pageService.ReadPage(ctx, 14, "movie-e2e", 0, 10)
	require.NoError(t, err)
	assert.Equal(t, "End-to-End Movie", page.MovieInfo.Title)
	assert.Equal(t, "An end-to-end plot", page.Plot)
	assert.Equal(t, []media.CastInfo{{CastInfoId: 201, Name: "Grace Actor", Gender: false, Intro: "Supporting actor"}}, page.CastInfo)
	require.Len(t, page.Reviews, 1)
	assert.Equal(t, "Excellent movie", page.Reviews[0].Text)
	assert.Equal(t, "movie-e2e", page.Reviews[0].MovieID)
	assert.Equal(t, 8, page.Reviews[0].Rating)
	assert.Equal(t, userID, page.Reviews[0].UserID)

	userReviews, err := userReviewService.ReadUserReviews(ctx, 15, userID, 0, 10)
	require.NoError(t, err)
	assert.Equal(t, page.Reviews, userReviews)

	assert.Error(t, wrk2APIService.RegisterMovie(ctx, "End-to-End Movie", "duplicate"))
	assert.Error(t, wrk2APIService.ComposeReview(ctx, "End-to-End Movie", "Rejected", "grace", "wrong", 3))
}
