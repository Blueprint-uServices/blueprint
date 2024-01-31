// Package preserve implements ts-preserve-service from the original Train Ticket application
package preserve

import (
	"context"

	"github.com/blueprint-uservices/blueprint/examples/train_ticket/workflow/common"
)

type PreserveService interface {
	Preserve(ctx context.Context, oti common.OrderTicketsInfo) error
}
