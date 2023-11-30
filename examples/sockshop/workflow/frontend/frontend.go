// Package frontend implements the SockShop frontend service, typically deployed via HTTP
package frontend

import (
	"context"

	"gitlab.mpi-sws.org/cld/blueprint/examples/sockshop/workflow/cart"
	"gitlab.mpi-sws.org/cld/blueprint/examples/sockshop/workflow/catalogue"
	"gitlab.mpi-sws.org/cld/blueprint/examples/sockshop/workflow/order"
	"gitlab.mpi-sws.org/cld/blueprint/examples/sockshop/workflow/user"
)

type (
	// The SockShop Frontend receives requests from users and proxies them to the application's other services
	Frontend interface {
	}
)

type frontend struct {
	user      user.UserService
	catalogue catalogue.CatalogueService
	cart      cart.CartService
	order     order.OrderService
}

func NewFrontend(ctx context.Context, user user.UserService, catalogue catalogue.CatalogueService, cart cart.CartService, order order.OrderService) (*frontend, error) {
	f := &frontend{
		user:      user,
		catalogue: catalogue,
		cart:      cart,
		order:     order,
	}
	return f, nil
}
