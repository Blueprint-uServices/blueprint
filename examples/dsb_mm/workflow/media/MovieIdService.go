package media

import (
	"context"
	"errors"

	"github.com/blueprint-uservices/blueprint/runtime/core/backend"
	"go.mongodb.org/mongo-driver/bson"
)

type MovieIdService interface {
	UploadMovieId(ctx context.Context, reqID int, title string, rating int) error
	RegisterMovieId(ctx context.Context, reqID int, title string, movieID string) error
}

type MovieIdServiceImpl struct {
	movieIdCache         backend.Cache
	movieIdDB            backend.NoSQLDatabase
	ratingService        RatingService
	composeReviewService ComposeReviewService
}

func NewMovieIdServiceImpl(ctx context.Context, movieIdCache backend.Cache, movieIdDB backend.NoSQLDatabase, ratingService RatingService, composeReviewService ComposeReviewService) (MovieIdService, error) {
	return &MovieIdServiceImpl{movieIdCache: movieIdCache, movieIdDB: movieIdDB, ratingService: ratingService, composeReviewService: composeReviewService}, nil
}

func (m *MovieIdServiceImpl) RegisterMovieId(ctx context.Context, reqID int, title string, movieID string) error {
	collection, err := m.movieIdDB.GetCollection(ctx, "movie-id", "movie-id")
	if err != nil {
		return err
	}
	result, err := collection.FindOne(ctx, bson.D{{"title", title}})
	if err != nil {
		return err
	}
	var existing MovieID
	found, err := result.One(ctx, &existing)
	if err != nil {
		return err
	}
	if found {
		return errors.New("movie already exists in the database")
	}
	movie := MovieID{MovID: movieID, Title: title}
	if err := collection.InsertOne(ctx, movie); err != nil {
		return err
	}
	return m.movieIdCache.Put(ctx, title, movieID)
}

func (m *MovieIdServiceImpl) UploadMovieId(ctx context.Context, reqID int, title string, rating int) error {
	var movieID string
	found, err := m.movieIdCache.Get(ctx, title, &movieID)
	if err != nil {
		return err
	}
	if !found {
		collection, err := m.movieIdDB.GetCollection(ctx, "movie-id", "movie-id")
		if err != nil {
			return err
		}
		result, err := collection.FindOne(ctx, bson.D{{"title", title}})
		if err != nil {
			return err
		}
		var movie MovieID
		found, err = result.One(ctx, &movie)
		if err != nil {
			return err
		}
		if !found {
			return errors.New("movie not found")
		}
		movieID = movie.MovID
		if err := m.movieIdCache.Put(ctx, title, movieID); err != nil {
			return err
		}
	}
	if err := m.composeReviewService.UploadMovieId(ctx, reqID, movieID); err != nil {
		return err
	}
	return m.ratingService.UploadRating(ctx, reqID, movieID, rating)
}
