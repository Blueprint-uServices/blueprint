// package train implements ts-train-service from the original TrainTicket application
package train

type TrainType struct {
	ID           string
	Name         string
	EconomyClass int64
	ComfortClass int64
	AvgSpeed     int64
}
