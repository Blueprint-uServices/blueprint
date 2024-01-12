package voucher

type Voucher struct {
	VoucherId    string
	OrderId      string
	TravelDate   string
	ContactName  string
	TrainNumber  string
	SeatClass    uint16
	SeatNumber   string
	StartStation string
	DestStation  string
	Price        float64
}
