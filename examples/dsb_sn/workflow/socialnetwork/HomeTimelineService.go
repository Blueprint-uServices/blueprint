package socialnetwork

import (
	"context"
	"strconv"

	"gitlab.mpi-sws.org/cld/blueprint/runtime/core/backend"
)

// The HomeTimelineService Interface
// The full Timeline of a user is represented as an array of post ids: post_ids[id_0 ,..., id_n].
type HomeTimelineService interface {
	// Reads the timeline of the user that has the id `userID`.
	// The return value is represented by the slice: post_ids[start:stop].
	ReadHomeTimeline(ctx context.Context, reqID int64, userID int64, start int64, stop int64) ([]int64, error)
	// Adds a new post to the home timeline of the following users:
	// (i)   user with id `userID`,
	// (ii)  all the followers of the user with `userID`
	// (iii) all the mentioned users in the post listed in the `userMentionIDs`.
	// The new post ID is placed at the nth position in the post ids array.
	//    post_ids = append(post_ids, `postID`)
	WriteHomeTimeline(ctx context.Context, reqID int64, postID int64, userID int64, timestamp int64, userMentionIDs []int64) error
}

// Implementation of [HomeTimelineService]
type HomeTimelineServiceImpl struct {
	homeTimelineCache  backend.Cache
	postStorageService PostStorageService
	socialGraphService SocialGraphService
}

// Creates a [HomeTimelineService] instance that maintains the home timelines for the various users.
func NewHomeTimelineServiceImpl(ctx context.Context, homeTimelineCache backend.Cache, postStorageService PostStorageService, socialGraphService SocialGraphService) (HomeTimelineService, error) {
	return &HomeTimelineServiceImpl{homeTimelineCache: homeTimelineCache, postStorageService: postStorageService, socialGraphService: socialGraphService}, nil
}

// Implements HomeTimelineService interface
func (h *HomeTimelineServiceImpl) WriteHomeTimeline(ctx context.Context, reqID int64, postID int64, userID int64, timestamp int64, userMentionIDs []int64) error {
	followers, err := h.socialGraphService.GetFollowers(ctx, reqID, userID)
	if err != nil {
		return err
	}
	followers_set := make(map[int64]bool)
	for _, follower := range followers {
		followers_set[follower] = true
	}
	for _, um := range userMentionIDs {
		followers_set[um] = true
	}
	for id, _ := range followers_set {
		id_str := strconv.FormatInt(id, 10)
		var posts []PostInfo
		_, err = h.homeTimelineCache.Get(ctx, id_str, &posts)
		if err != nil {
			return err
		}
		posts = append(posts, PostInfo{PostID: postID, Timestamp: timestamp})
		err = h.homeTimelineCache.Put(ctx, id_str, posts)
		if err != nil {
			return err
		}
	}
	return nil
}

// Implements HomeTimelineService interface
func (h *HomeTimelineServiceImpl) ReadHomeTimeline(ctx context.Context, reqID int64, userID int64, start int64, stop int64) ([]int64, error) {
	if stop <= start || start < 0 {
		return []int64{}, nil
	}
	userIDStr := strconv.FormatInt(userID, 10)
	var postIDs []int64
	var postInfos []PostInfo
	_, err := h.homeTimelineCache.Get(ctx, userIDStr, &postInfos)
	if err != nil {
		return []int64{}, err
	}
	for _, pinfo := range postInfos {
		postIDs = append(postIDs, pinfo.PostID)
	}
	if start < int64(len(postIDs)) {
		minstop := stop
		if stop > int64(len(postIDs)) {
			minstop = int64(len(postIDs))
		}
		postIDs = postIDs[start:minstop]
	}
	_, err = h.postStorageService.ReadPosts(ctx, reqID, postIDs)
	if err != nil {
		return postIDs, err
	}
	return postIDs, nil
}
