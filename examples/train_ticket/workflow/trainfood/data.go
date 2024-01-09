package trainfood

import "github.com/Blueprint-uServices/blueprint/examples/train_ticket/workflow/food"

type TrainFood struct {
	ID     string
	TripID string
	Foods  []food.Food
}
