package trainfood

import "github.com/blueprint-uservices/blueprint/examples/train_ticket/workflow/food"

type TrainFood struct {
	ID     string
	TripID string
	Foods  []food.Food
}
