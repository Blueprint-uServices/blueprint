package media

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/blueprint-uservices/blueprint/runtime/core/backend"
)

type ComposeReviewService interface {
	UploadMovieId(ctx context.Context, reqID int, movieID string) error
	UploadUserId(ctx context.Context, reqID int, userID int) error
	UploadUniqueId(ctx context.Context, reqID int, reviewID int) error
	UploadText(ctx context.Context, reqID int, text string) error
	UploadRating(ctx context.Context, reqID int, rating int) error
}

type ComposeReviewServiceImpl struct {
	composeReviewCache   backend.Cache
	reviewStorageService ReviewStorageService
	userReviewService    UserReviewService
	movieReviewService   MovieReviewService
	mu                   sync.Mutex
}

func NewComposeReviewServiceImpl(composeReviewCache backend.Cache, reviewStorageService ReviewStorageService, userReviewService UserReviewService, movieReviewService MovieReviewService) (ComposeReviewService, error) {
	return &ComposeReviewServiceImpl{
		composeReviewCache: composeReviewCache, reviewStorageService: reviewStorageService,
		userReviewService: userReviewService, movieReviewService: movieReviewService,
	}, nil
}

func (c *ComposeReviewServiceImpl) composeAndUpload(ctx context.Context, reqID int) error {
	prefix := strconvReqID(reqID)
	keys := []string{prefix + ":review_id", prefix + ":movie_id", prefix + ":user_id", prefix + ":text", prefix + ":rating"}
	var reviewID, userID, rating int
	var movieID, text string
	values := []interface{}{&reviewID, &movieID, &userID, &text, &rating}
	if err := c.composeReviewCache.Mget(ctx, keys, values); err != nil {
		return err
	}
	if movieID == "" {
		return fmt.Errorf("review %d is missing a movie ID", reqID)
	}
	timestamp := time.Now().UTC().Format(time.RFC3339Nano)
	review := Review{ReviewID: reviewID, UserID: userID, ReqID: reqID, Text: text, MovieID: movieID, Rating: rating, Timestamp: timestamp}
	if err := c.reviewStorageService.StoreReview(ctx, reqID, review); err != nil {
		return err
	}
	if err := c.userReviewService.UploadUserReview(ctx, reqID, userID, reviewID, timestamp); err != nil {
		return err
	}
	if err := c.movieReviewService.UploadMovieReview(ctx, reqID, movieID, reviewID, timestamp); err != nil {
		return err
	}
	for _, key := range append(keys, prefix+":counter") {
		if err := c.composeReviewCache.Delete(ctx, key); err != nil {
			return err
		}
	}
	return nil
}

func strconvReqID(reqID int) string { return fmt.Sprintf("%d", reqID) }

func (c *ComposeReviewServiceImpl) uploadComponent(ctx context.Context, reqID int, suffix string, value interface{}) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	prefix := strconvReqID(reqID)
	if err := c.composeReviewCache.Put(ctx, prefix+suffix, value); err != nil {
		return err
	}
	count, err := c.composeReviewCache.Incr(ctx, prefix+":counter")
	if err != nil {
		return err
	}
	if count == 5 {
		return c.composeAndUpload(ctx, reqID)
	}
	return nil
}

func (c *ComposeReviewServiceImpl) UploadMovieId(ctx context.Context, reqID int, movieID string) error {
	return c.uploadComponent(ctx, reqID, ":movie_id", movieID)
}
func (c *ComposeReviewServiceImpl) UploadUserId(ctx context.Context, reqID int, userID int) error {
	return c.uploadComponent(ctx, reqID, ":user_id", userID)
}
func (c *ComposeReviewServiceImpl) UploadUniqueId(ctx context.Context, reqID int, reviewID int) error {
	return c.uploadComponent(ctx, reqID, ":review_id", reviewID)
}
func (c *ComposeReviewServiceImpl) UploadText(ctx context.Context, reqID int, text string) error {
	return c.uploadComponent(ctx, reqID, ":text", text)
}
func (c *ComposeReviewServiceImpl) UploadRating(ctx context.Context, reqID int, rating int) error {
	return c.uploadComponent(ctx, reqID, ":rating", rating)
}
