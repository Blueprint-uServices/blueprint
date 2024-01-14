package routeplan

type RoutePlanInfo struct {
	Num          int64
	StartStation string
	EndStation   string
	TravelDate   string
}

type RoutePlanResultUnit struct {
	ID                      string
	TrainTypeName           string
	StartStation            string
	EndStation              string
	StopStations            []string
	PriceForSecondClassSeat float64
	PriceForFirstClassSeat  float64
	StartTime               string
	EndTime                 string
}
