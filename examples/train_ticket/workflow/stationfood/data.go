package stationfood

import (
	"github.com/Blueprint-uServices/blueprint/examples/train_ticket/workflow/food"
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
