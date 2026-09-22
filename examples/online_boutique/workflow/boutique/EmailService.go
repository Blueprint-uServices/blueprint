package boutique

import (
	"context"
	"fmt"
	"strings"

	"github.com/blueprint-uservices/blueprint/runtime/core/backend"
)

type EmailService interface {
	SendOrderConfirmation(ctx context.Context, email string, order Order) error
}

type EmailServiceImpl struct {
	logger backend.Logger
}

func NewEmailServiceImpl(ctx context.Context) (EmailService, error) {
	return &EmailServiceImpl{logger: backend.GetLogger()}, nil
}

func (s *EmailServiceImpl) SendOrderConfirmation(ctx context.Context, email string, order Order) error {
	var confirmation strings.Builder
	fmt.Fprintf(&confirmation, "Order confirmation #%s\nShipping: %s\n", order.OrderID, order.ShippingTrackingID)
	fmt.Fprintf(&confirmation, "Shipping cost: %d.%02d %s\n", order.ShippingCost.Units, order.ShippingCost.Nanos/10_000_000, order.ShippingCost.CurrencyCode)
	for _, item := range order.Items {
		fmt.Fprintf(&confirmation, "%s x%d: %d.%02d %s\n", item.Item.ProductID, item.Item.Quantity, item.Cost.Units, item.Cost.Nanos/10_000_000, item.Cost.CurrencyCode)
	}
	_, err := s.logger.Info(ctx, "A request to send email confirmation to %s has been received:\n%s", email, confirmation.String())
	return err
}
