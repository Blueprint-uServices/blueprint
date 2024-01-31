package trainfood

import "github.com/blueprint-uservices/blueprint/examples/train_ticket/workflow/common"

type TrainFood struct {
	ID     string
	TripID string
	Foods  []common.Food
}
