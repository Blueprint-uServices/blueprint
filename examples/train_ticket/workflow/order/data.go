package order

type Order struct {
	Id                     string
	BoughtDate             string
	TravelDate             string
	AccountId              string
	ContactsName           string
	DocumentType           uint16
	ContactsDocumentNumber string
	TrainNumber            string
	CoachNumber            string
	SeatClass              uint16
	SeatNumber             string
	From                   string
	To                     string
	Status                 uint16
	Price                  float64
}

type OrderInfo struct {
	LoginId               string
	TravelDateStart       string
	TravelDateEnd         string
	BoughtDateStart       string
	BoughtDateEnd         string
	State                 uint16
	EnableTravelDateQuery bool
	EnableBoughtDateQuery bool
	EnableStateQuery      bool
}
type SoldTicket struct {
	TravelDate      string
	TrainNumber     string
	NoSeat          uint16
	BusinessSeat    uint16
	FirstClassSeat  uint16
	SecondClassSeat uint16
	HardSeat        uint16
	SoftSeat        uint16
	HardBed         uint16
	SoftBed         uint16
	HighSoftBed     uint16
}

type Ticket struct {
	SeatNo       string
	StartStation string
	DestStation  string
}

// SeatClass
const (
	None uint16 = iota
	Business
	FirstClass
	SecondClass
	HardSeat
	SoftSeat
	HardBed
	SoftBed
	HighSoftBed
)

// OrderStatus
const (
	NotPaid uint16 = iota
	Paid
	Collected
	Change
	Cancel
	Refund
	Used
)
