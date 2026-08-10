package media

import (
	"context"
	"hash/fnv"
	"net"
	"sync"
	"time"
)

type UniqueIdService interface {
	UploadUniqueId(ctx context.Context, reqID int64) error
}

type UniqueIdServiceImpl struct {
	composeReviewService ComposeReviewService
	counter              int64
	currentTimestamp     int64
	machineID            string
	mu                   sync.Mutex
}

func GetMachineID() string {
	interfaces, err := net.Interfaces()
	if err != nil {
		return "0"
	}
	for _, iface := range interfaces {
		if iface.Flags&net.FlagUp != 0 && len(iface.HardwareAddr) != 0 {
			return iface.HardwareAddr.String()
		}
	}
	return "0"
}

func NewUniqueIdServiceImpl(composeReviewService ComposeReviewService) (UniqueIdService, error) {
	return &UniqueIdServiceImpl{composeReviewService: composeReviewService, currentTimestamp: -1, machineID: GetMachineID()}, nil
}

func generatedID(machineID string, timestamp int64, counter int64) int64 {
	h := fnv.New32a()
	_, _ = h.Write([]byte(machineID))
	machine := int64(h.Sum32() & 0x3ff)
	return ((timestamp & ((1 << 41) - 1)) << 22) | (machine << 12) | (counter & 0xfff)
}

func (u *UniqueIdServiceImpl) nextID() int64 {
	u.mu.Lock()
	defer u.mu.Unlock()
	timestamp := time.Now().UnixMilli()
	if u.currentTimestamp == timestamp {
		u.counter++
	} else {
		u.currentTimestamp = timestamp
		u.counter = 0
	}
	return generatedID(u.machineID, timestamp, u.counter)
}

func (u *UniqueIdServiceImpl) UploadUniqueId(ctx context.Context, reqID int64) error {
	return u.composeReviewService.UploadUniqueId(ctx, int(reqID), int(u.nextID()))
}
