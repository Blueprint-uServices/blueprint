package travel

type TripResponse struct {
	ComfortClass         int64
	EconomyClass         int64
	StartingStation      string
	EndStation           string
	StartingTime         string
	EndTime              string
	Duration             string
	TripId               string
	TrainTypeId          string
	PriceForComfortClass float64
	PriceForEconomyClass float64

	StopStations                  []string
	NumberOfRestTicketFirstClass  uint16
	NumberOfRestTicketSecondClass uint16
}
