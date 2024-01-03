// Package common implements ts-common from the original train ticket application
package common

import (
	"gitlab.mpi-sws.org/cld/blueprint/examples/train_ticket/workflow/route"
	"gitlab.mpi-sws.org/cld/blueprint/examples/train_ticket/workflow/train"
)

type TripType struct {
	Name  string
	Index int64
}

func getTripTypeFromIndex(index int64) TripType {
	switch index {
	case 1:
		return TripType{Name: "G", Index: index}
	case 2:
		return TripType{Name: "D", Index: index}
	case 3:
		return TripType{Name: "Z", Index: index}
	case 4:
		return TripType{Name: "T", Index: index}
	case 5:
		return TripType{Name: "K", Index: index}
	}
	return TripType{}
}

func getTripTypeFromChar(char string) TripType {
	switch {
	case char == "G":
		return TripType{Name: char, Index: 1}
	case char == "D":
		return TripType{Name: char, Index: 2}
	case char == "Z":
		return TripType{Name: char, Index: 3}
	case char == "T":
		return TripType{Name: char, Index: 4}
	case char == "K":
		return TripType{Name: char, Index: 5}
	}
	return TripType{}
}

type TripID struct {
	TT     TripType
	Number string
}

func parseTripID(trainNum string) TripID {
	ttype := getTripTypeFromChar(trainNum[:1])
	number := trainNum[1:]
	return TripID{TT: ttype, Number: number}
}

type Trip struct {
	ID                  string
	TID                 TripID
	TrainTypeName       string
	RouteID             string
	StartStationName    string
	StationsName        string
	TerminalStationName string
	StartTime           string
	EndTime             string
}

type Travel struct {
	T             Trip
	StartPlace    string
	EndPlace      string
	DepartureTime string
}

type TravelResult struct {
	Status  bool
	Percent float64
	TType   train.TrainType
	Route   route.Route
	Prices  map[string]string
}
