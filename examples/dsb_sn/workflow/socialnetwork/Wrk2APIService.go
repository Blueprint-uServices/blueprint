package socialnetwork

import (
	"context"
	"errors"
	"math/rand"
	"time"
)

// The Wrk2APIService (Frontend) interface
type Wrk2APIService interface {
	// Reads the home timeline of the user with `userId`.
	// Returns the list of posts[start, stop] from the timeline.
	ReadHomeTimeline(ctx context.Context, userId int64, start int64, stop int64) ([]int64, error)
	// Reads the user timeline of the user with `userId`.
	// Returns the list of posts[start, stop] from the timeline.
	ReadUserTimeline(ctx context.Context, userId int64, start int64, stop int64) ([]int64, error)
	// Creates a Follow-Link between users `userId`-`followeeID`.
	// If the user ids are not provided, then it creates a follow-link between users `username`-`followeeName`.
	// Returns an error if no pairs are provided.
	Follow(ctx context.Context, username string, followeeName string, userId int64, followeeID int64) error
	// Removes a Follow-Link between users `userId`-`followeeID`.
	// If the user ids are not provided, then it creates a follow-link between users `username`-`followeeName`.
	// Returns an error if no pairs are provided.
	Unfollow(ctx context.Context, username string, followeeName string, userId int64, followeeID int64) error
	// Registers a new user with the given `userId`.
	Register(ctx context.Context, firstName string, lastName string, username string, password string, userId int64) error
	// Composes a new post give the provided arguments.
	// Returns the created post's ID and the ids of the mentioned users.
	ComposePost(ctx context.Context, userId int64, username string, post_type int64, text string, media_types []string, media_ids []int64) (int64, []int64, error)
}

// Implementation of [Wrk2APIService]
type Wrk2APIServiceImpl struct {
	userService         UserService
	composePostService  ComposePostService
	userTimelineService UserTimelineService
	homeTimelineService HomeTimelineService
	socialGraphService  SocialGraphService
}

// Creates a [Wrk2APIService] instance to act as the gateway to internal services.
func NewWrk2APIServiceImpl(ctx context.Context, userService UserService, composePostService ComposePostService, userTimelineService UserTimelineService, homeTimelineService HomeTimelineService, socialGraphService SocialGraphService) (Wrk2APIService, error) {
	rand.Seed(time.Now().UnixNano())
	return &Wrk2APIServiceImpl{userService: userService, composePostService: composePostService, userTimelineService: userTimelineService, homeTimelineService: homeTimelineService, socialGraphService: socialGraphService}, nil
}

// Implements Wrk2APIService interface
func (w *Wrk2APIServiceImpl) ReadUserTimeline(ctx context.Context, userId int64, start int64, stop int64) ([]int64, error) {
	reqID := rand.Int63()
	return w.userTimelineService.ReadUserTimeline(ctx, reqID, userId, start, stop)
}

// Implements Wrk2APIService interface
func (w *Wrk2APIServiceImpl) ReadHomeTimeline(ctx context.Context, userId int64, start int64, stop int64) ([]int64, error) {
	reqID := rand.Int63()
	return w.homeTimelineService.ReadHomeTimeline(ctx, reqID, userId, start, stop)
}

// Implements Wrk2APIService interface
func (w *Wrk2APIServiceImpl) Follow(ctx context.Context, username string, followeeName string, userId int64, followeeID int64) error {
	reqID := rand.Int63()
	if userId != 0 && followeeID != 0 {
		return w.socialGraphService.Follow(ctx, reqID, userId, followeeID)
	} else if username != "" && followeeName != "" {
		return w.socialGraphService.FollowWithUsername(ctx, reqID, username, followeeName)
	}
	return errors.New("Invalid Arguments")
}

// Implements Wrk2APIService interface
func (w *Wrk2APIServiceImpl) Unfollow(ctx context.Context, username string, followeeName string, userId int64, followeeID int64) error {
	reqID := rand.Int63()
	if userId != 0 && followeeID != 0 {
		return w.socialGraphService.Unfollow(ctx, reqID, userId, followeeID)
	} else if username != "" && followeeName != "" {
		return w.socialGraphService.UnfollowWithUsername(ctx, reqID, username, followeeName)
	}
	return errors.New("Invalid Arguments")
}

// Implements Wrk2APIService interface
func (w *Wrk2APIServiceImpl) Register(ctx context.Context, firstName string, lastName string, username string, password string, userId int64) error {
	if firstName == "" || lastName == "" || username == "" || password == "" {
		return errors.New("Incomplete Arguments")
	}
	reqID := rand.Int63()
	return w.userService.RegisterUserWithId(ctx, reqID, firstName, lastName, username, password, userId)
}

// Implements Wrk2APIService interface
func (w *Wrk2APIServiceImpl) ComposePost(ctx context.Context, userId int64, username string, post_type int64, text string, media_types []string, media_ids []int64) (int64, []int64, error) {
	if userId == 0 || username == "" || text == "" {
		return -1, []int64{}, errors.New("Incomplete Arguments")
	}
	reqID := rand.Int63()
	return w.composePostService.ComposePost(ctx, reqID, username, userId, text, media_ids, media_types, post_type)
}
