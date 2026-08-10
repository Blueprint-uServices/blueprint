package media

import (
	"context"
	"strconv"

	"github.com/blueprint-uservices/blueprint/runtime/core/backend"
	"go.mongodb.org/mongo-driver/bson"
)

type ReviewStorageService interface {
	StoreReview(ctx context.Context, reqID int, review Review) error
	ReadReviews(ctx context.Context, reqID int, reviewIDs []int) ([]Review, error)
}

type ReviewStorageServiceImpl struct {
	reviewStorageCache backend.Cache
	reviewStorageDB    backend.NoSQLDatabase
}

func NewReviewStorageServiceImpl(reviewStorageCache backend.Cache, reviewStorageDB backend.NoSQLDatabase) (ReviewStorageService, error) {
	return &ReviewStorageServiceImpl{reviewStorageCache: reviewStorageCache, reviewStorageDB: reviewStorageDB}, nil
}

func (r *ReviewStorageServiceImpl) StoreReview(ctx context.Context, reqID int, review Review) error {
	collection, err := r.reviewStorageDB.GetCollection(ctx, "review", "review")
	if err != nil {
		return err
	}
	if err := collection.InsertOne(ctx, review); err != nil {
		return err
	}
	return r.reviewStorageCache.Put(ctx, strconv.Itoa(review.ReviewID), review)
}

func (r *ReviewStorageServiceImpl) ReadReviews(ctx context.Context, reqID int, reviewIDs []int) ([]Review, error) {
	reviews := make([]Review, 0, len(reviewIDs))
	for _, id := range reviewIDs {
		var review Review
		found, err := r.reviewStorageCache.Get(ctx, strconv.Itoa(id), &review)
		if err != nil {
			return nil, err
		}
		if !found {
			collection, err := r.reviewStorageDB.GetCollection(ctx, "review", "review")
			if err != nil {
				return nil, err
			}
			result, err := collection.FindOne(ctx, bson.D{{"reviewid", id}})
			if err != nil {
				return nil, err
			}
			found, err = result.One(ctx, &review)
			if err != nil {
				return nil, err
			}
			if !found {
				continue
			}
			if err := r.reviewStorageCache.Put(ctx, strconv.Itoa(id), review); err != nil {
				return nil, err
			}
		}
		reviews = append(reviews, review)
	}
	return reviews, nil
}
