package waitlistorder

type WaitlistOrder struct {
	ID             string
	AccountID      string
	ContactID      string
	ContactName    string
	ContactDocType int
	ContactDocNum  int
	TrainNumber    string
	SeatType       int
	From           string
	To             string
	Price          string
	WaitUtilTime   string
	CreatedTime    string
	Status         int
}

type WaitlistOrderVO struct {
	AccountID string
	TripID    string
	SeatType  int64
	Date      string
	From      string
	To        string
	Price     string
}

const (
	NOTPAID int = iota
	PAID
	COLLECTED
	CANCEL
	REFUNDS
	EXPIRED
)
