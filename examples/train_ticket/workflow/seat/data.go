package seat

const (
	NONE = iota
	BUSINESS
	FIRSTCLASS
	SECONDCLASS
	HARDSEAT
	SOFTSEAT
	HARDBED
	SOFTBED
	HIGHSOFTBED
)

type Seat struct {
	TravelDate   string
	TrainNumber  string
	StartStation string
	DstStation   string
	SeatType     int
	TotalNum     int64
	Stations     []string
}
