package stationfood

import "github.com/blueprint-uservices/blueprint/examples/train_ticket/workflow/common"

type StationFoodStore struct {
	ID           string
	StationName  string
	StoreName    string
	Telephone    string
	BusinessTime string
	DeliveryFee  float64
	Foods        []common.Food
}
