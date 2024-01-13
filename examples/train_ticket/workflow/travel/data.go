package travel

type TripResponse struct {
	ComfortClass         uint16
	EconomyClass         uint16
	StartingStation      string
	EndStation           string
	StartingTime         string
	EndTime              string
	TripId               string
	TrainTypeId          string
	PriceForComfortClass float32
	PriceForEconomyClass float32

	StopStations                  []string
	NumberOfRestTicketFirstClass  uint16
	NumberOfRestTicketSecondClass uint16
}
