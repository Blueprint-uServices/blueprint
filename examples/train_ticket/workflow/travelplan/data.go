package travelplan

import "github.com/blueprint-uservices/blueprint/examples/train_ticket/workflow/travel"

type TransferTravelInfo struct {
	StartStation string
	ViaStation   string
	EndStation   string
	TravelDate   string
	TrainType    string
}

type TransferTravelResult struct {
	FirstSection  []travel.TripResponse
	SecondSection []travel.TripResponse
}

type TravelAdvanceResult struct {
	TripID                  string
	TrainTypeId             string
	StartStation            string
	EndStation              string
	StopStations            []string
	PriceForSecondClassSeat float64
	RemainingSecondClassTix int64
	PriceForFirstClassSeat  float64
	RemainingFirstClassTix  int64
	StartTime               string
	EndTime                 string
}
