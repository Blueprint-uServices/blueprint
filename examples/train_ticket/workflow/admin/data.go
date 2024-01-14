package admin

import (
	"github.com/blueprint-uservices/blueprint/examples/train_ticket/workflow/common"
	"github.com/blueprint-uservices/blueprint/examples/train_ticket/workflow/route"
	"github.com/blueprint-uservices/blueprint/examples/train_ticket/workflow/train"
)

type AdminTrip struct {
	T  common.Trip
	R  route.Route
	TT train.TrainType
}
