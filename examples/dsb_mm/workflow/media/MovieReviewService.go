package media

import (
	"context"
	"errors"
	"sync"

	"github.com/blueprint-uservices/blueprint/runtime/core/backend"
	"go.mongodb.org/mongo-driver/bson"
)

type ReviewInfo struct {
	ReviewID  int
	Timestamp string
}

type MovieReview struct {
	MovieID     string
	ReviewInfos []ReviewInfo
}

type MovieReviewService interface {
	UploadMovieReview(ctx context.Context, reqID int, movieID string, reviewID int, timestamp string) error
	ReadMovieReviews(ctx context.Context, reqID int, movieID string, start int, stop int) ([]Review, error)
}

type MovieReviewServiceImpl struct {
	reviewStorageService ReviewStorageService
	movieReviewCache     backend.Cache
	movieReviewDB        backend.NoSQLDatabase
	mu                   sync.Mutex
}

func NewMovieReviewServiceImpl(reviewStorageService ReviewStorageService, movieReviewCache backend.Cache, movieReviewDB backend.NoSQLDatabase) (MovieReviewService, error) {
	return &MovieReviewServiceImpl{reviewStorageService: reviewStorageService, movieReviewCache: movieReviewCache, movieReviewDB: movieReviewDB}, nil
}

func (m *MovieReviewServiceImpl) UploadMovieReview(ctx context.Context, reqID int, movieID string, reviewID int, timestamp string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	collection, err := m.movieReviewDB.GetCollection(ctx, "movie-review", "movie-review")
	if err != nil {
		return err
	}
	query := bson.D{{"movieid", movieID}}
	result, err := collection.FindOne(ctx, query)
	if err != nil {
		return err
	}
	var movieReview MovieReview
	found, err := result.One(ctx, &movieReview)
	if err != nil {
		return err
	}
	movieReview.MovieID = movieID
	movieReview.ReviewInfos = append([]ReviewInfo{{ReviewID: reviewID, Timestamp: timestamp}}, movieReview.ReviewInfos...)
	if found {
		_, err = collection.ReplaceOne(ctx, query, movieReview)
	} else {
		err = collection.InsertOne(ctx, movieReview)
	}
	if err != nil {
		return err
	}
	return m.movieReviewCache.Put(ctx, movieID, movieReview.ReviewInfos)
}

func (m *MovieReviewServiceImpl) ReadMovieReviews(ctx context.Context, reqID int, movieID string, start int, stop int) ([]Review, error) {
	if start < 0 || stop <= start {
		return []Review{}, nil
	}
	infos, err := m.readReviewInfos(ctx, movieID)
	if err != nil {
		return nil, err
	}
	if start >= len(infos) {
		return []Review{}, nil
	}
	if stop > len(infos) {
		stop = len(infos)
	}
	ids := make([]int, 0, stop-start)
	for _, info := range infos[start:stop] {
		ids = append(ids, info.ReviewID)
	}
	return m.reviewStorageService.ReadReviews(ctx, reqID, ids)
}

func (m *MovieReviewServiceImpl) readReviewInfos(ctx context.Context, movieID string) ([]ReviewInfo, error) {
	var infos []ReviewInfo
	found, err := m.movieReviewCache.Get(ctx, movieID, &infos)
	if err != nil || found {
		return infos, err
	}
	collection, err := m.movieReviewDB.GetCollection(ctx, "movie-review", "movie-review")
	if err != nil {
		return nil, err
	}
	result, err := collection.FindOne(ctx, bson.D{{"movieid", movieID}})
	if err != nil {
		return nil, err
	}
	var movieReview MovieReview
	found, err = result.One(ctx, &movieReview)
	if err != nil {
		return nil, err
	}
	if !found {
		return []ReviewInfo{}, nil
	}
	if movieReview.MovieID == "" {
		return nil, errors.New("invalid movie review record")
	}
	return movieReview.ReviewInfos, m.movieReviewCache.Put(ctx, movieID, movieReview.ReviewInfos)
}
