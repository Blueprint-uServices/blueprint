package media

import (
	"context"
	"errors"

	"github.com/blueprint-uservices/blueprint/runtime/core/backend"
	"go.mongodb.org/mongo-driver/bson"
)

type MovieInfoService interface {
	WriteMovieInfo(ctx context.Context, reqID int, movieID string, title string, casts []Cast, plotID int, thumbnailIDs []string, photoIDs []string, videoIDs []string, avgRating float64, numRating int) error
	ReadMovieInfo(ctx context.Context, reqID int, movieID string) (MovieInfo, error)
	UpdateRating(ctx context.Context, reqID int, movieID string, sumUncommittedRating int, numUncommittedRating int) error
}

type MovieInfoServiceImpl struct {
	movieInfoCache backend.Cache
	movieInfoDB    backend.NoSQLDatabase
}

func NewMovieInfoServiceImpl(movieInfoCache backend.Cache, movieInfoDB backend.NoSQLDatabase) (MovieInfoService, error) {
	return &MovieInfoServiceImpl{movieInfoCache: movieInfoCache, movieInfoDB: movieInfoDB}, nil
}

func (m *MovieInfoServiceImpl) WriteMovieInfo(ctx context.Context, reqID int, movieID string, title string, casts []Cast, plotID int, thumbnailIDs []string, photoIDs []string, videoIDs []string, avgRating float64, numRating int) error {
	info := MovieInfo{movieID, title, casts, plotID, thumbnailIDs, photoIDs, videoIDs, avgRating, numRating}
	collection, err := m.movieInfoDB.GetCollection(ctx, "movie-info", "movie-info")
	if err != nil {
		return err
	}
	if err := collection.InsertOne(ctx, info); err != nil {
		return err
	}
	return m.movieInfoCache.Put(ctx, movieID, info)
}

func (m *MovieInfoServiceImpl) ReadMovieInfo(ctx context.Context, reqID int, movieID string) (MovieInfo, error) {
	var info MovieInfo
	found, err := m.movieInfoCache.Get(ctx, movieID, &info)
	if err != nil {
		return info, err
	}
	if found {
		return info, nil
	}
	collection, err := m.movieInfoDB.GetCollection(ctx, "movie-info", "movie-info")
	if err != nil {
		return info, err
	}
	result, err := collection.FindOne(ctx, bson.D{{"movieid", movieID}})
	if err != nil {
		return info, err
	}
	found, err = result.One(ctx, &info)
	if err != nil {
		return info, err
	}
	if !found {
		return info, errors.New("movie not found")
	}
	return info, m.movieInfoCache.Put(ctx, movieID, info)
}

func (m *MovieInfoServiceImpl) UpdateRating(ctx context.Context, reqID int, movieID string, sumUncommittedRating int, numUncommittedRating int) error {
	if numUncommittedRating <= 0 {
		return nil
	}
	info, err := m.ReadMovieInfo(ctx, reqID, movieID)
	if err != nil {
		return err
	}
	info.NumRating += numUncommittedRating
	info.AvgRating = (info.AvgRating*float64(info.NumRating-numUncommittedRating) + float64(sumUncommittedRating)) / float64(info.NumRating)
	collection, err := m.movieInfoDB.GetCollection(ctx, "movie-info", "movie-info")
	if err != nil {
		return err
	}
	if _, err := collection.ReplaceOne(ctx, bson.D{{"movieid", movieID}}, info); err != nil {
		return err
	}
	return m.movieInfoCache.Put(ctx, movieID, info)
}
