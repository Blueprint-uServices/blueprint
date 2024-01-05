package stationfood

import (
	"gitlab.mpi-sws.org/cld/blueprint/examples/train_ticket/workflow/food"
)

type StationFoodStore struct {
	ID           string
	StationName  string
	StoreName    string
	Telephone    string
	BusinessTime string
	DeliveryFee  float64
	Foods        []food.Food
}
