package media

import (
	"context"
	"fmt"
	"sync"

	"github.com/blueprint-uservices/blueprint/runtime/core/backend"
)

type RatingService interface {
	UploadRating(ctx context.Context, reqID int, movieID string, rating int) error
}

type RatingServiceImpl struct {
	composeReviewService ComposeReviewService
	ratingCache          backend.Cache
	mu                   sync.Mutex
}

func NewRatingServiceImpl(ctx context.Context, composeReviewService ComposeReviewService, ratingCache backend.Cache) (RatingService, error) {
	return &RatingServiceImpl{composeReviewService: composeReviewService, ratingCache: ratingCache}, nil
}

func (r *RatingServiceImpl) UploadRating(ctx context.Context, reqID int, movieID string, rating int) error {
	if rating < 0 || rating > 10 {
		return fmt.Errorf("rating must be between 0 and 10")
	}
	if err := r.composeReviewService.UploadRating(ctx, reqID, rating); err != nil {
		return err
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	var sum int
	_, err := r.ratingCache.Get(ctx, movieID+":uncommit_sum", &sum)
	if err != nil {
		return err
	}
	if err := r.ratingCache.Put(ctx, movieID+":uncommit_sum", sum+rating); err != nil {
		return err
	}
	_, err = r.ratingCache.Incr(ctx, movieID+":uncommit_num")
	return err
}
