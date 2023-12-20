package trainfood

import "gitlab.mpi-sws.org/cld/blueprint/examples/train_ticket/workflow/food"

type TrainFood struct {
	ID     string
	TripID string
	Foods  []food.Food
}
