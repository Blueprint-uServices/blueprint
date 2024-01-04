// Package hotelreservation implements the workflow specification of the Hotel Reservation application
package hotelreservation

type Point struct {
	Pid  string
	Plat float64
	Plon float64
}

func (p Point) remote() {}

func (p Point) Id() string { return p.Pid }

func (p Point) Lat() float64 { return p.Plat }

func (p Point) Lon() float64 { return p.Plon }

type User struct {
	Username string
	Password string
}

func (u User) remote() {}

type RoomType struct {
	BookableRate       float64
	Code               string
	RoomDescription    string
	TotalRate          float64
	TotalRateInclusive float64
}

func (rt RoomType) remote() {}

type RatePlan struct {
	HotelID string
	Code    string
	InDate  string
	OutDate string
	RType   RoomType
}

func (rp RatePlan) remote() {}

type Reservation struct {
	HotelId      string
	CustomerName string
	InDate       string
	OutDate      string
	Number       int64
}

func (r Reservation) remote() {}

type HotelNumber struct {
	HotelId string
	Number  int64
}

func (h HotelNumber) remote() {}

type Hotel struct {
	HId    string
	HLat   float64
	HLon   float64
	HRate  float64
	HPrice float64
}

func (h Hotel) remote() {}

type Address struct {
	StreetNumber string
	StreetName   string
	City         string
	State        string
	Country      string
	PostalCode   string
	Lat          float64
	Lon          float64
}

func (a Address) remote() {}

type HotelProfile struct {
	ID          string
	Name        string
	PhoneNumber string
	Description string
	Address     Address
}

func (hp HotelProfile) remote() {}
