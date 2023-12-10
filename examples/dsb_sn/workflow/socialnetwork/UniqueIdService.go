package socialnetwork

import (
	"context"
	"log"
	"strconv"
	"strings"
	"time"
)

// The UniqueIdService interface
type UniqueIdService interface {
	// Returns a newly generated unique id to be used as a post's unique identifier.
	ComposeUniqueId(ctx context.Context, reqID int64, postType int64) (int64, error)
}

// Implementation of [UserTimelineService]
type UniqueIdServiceImpl struct {
	counter           int64
	current_timestamp int64
	machine_id        string
}

// Implements UniqueIdService interface
func NewUniqueIdServiceImpl(ctx context.Context) (UniqueIdService, error) {
	return &UniqueIdServiceImpl{counter: 0, current_timestamp: -1, machine_id: GetMachineID()}, nil
}

func (u *UniqueIdServiceImpl) getCounter(timestamp int64) int64 {
	if u.current_timestamp == timestamp {
		retVal := u.counter
		u.counter += 1
		return retVal
	} else {
		u.current_timestamp = timestamp
		u.counter = 1
		return 0
	}
}

// Implements UniqueIdService interface
func (u *UniqueIdServiceImpl) ComposeUniqueId(ctx context.Context, reqID int64, postType int64) (int64, error) {
	timestamp := time.Now().UnixNano() / int64(time.Millisecond)
	idx := u.getCounter(timestamp)
	timestamp_hex := strconv.FormatInt(timestamp, 16)
	if len(timestamp_hex) > 10 {
		timestamp_hex = timestamp_hex[:10]
	} else if len(timestamp_hex) < 10 {
		timestamp_hex = strings.Repeat("0", 10-len(timestamp_hex)) + timestamp_hex
	}
	counter_hex := strconv.FormatInt(idx, 16)
	if len(counter_hex) > 1 {
		counter_hex = counter_hex[:1]
	} else if len(counter_hex) < 1 {
		counter_hex = strings.Repeat("0", 1-len(counter_hex)) + counter_hex
	}
	log.Println(u.machine_id, timestamp_hex, counter_hex)
	post_id_str := u.machine_id + timestamp_hex + counter_hex
	post_id, err := strconv.ParseInt(post_id_str, 16, 64)
	if err != nil {
		return 0, err
	}
	post_id = post_id & 0x7FFFFFFFFFFFFFFF
	return post_id, nil
}
