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
		// List items in cart for current logged in user, or for the current session if not logged in.
		// SessionID can be the empty string for a non-logged in user / new session
		GetCart(ctx context.Context, sessionID string) ([]cart.Item, error)

		// Deletes the entire cart for a user/session
		DeleteCart(ctx context.Context, sessionID string) error

		// Deletes an item from the user/session's cart
		DeleteItem(ctx context.Context, sessionID string, itemID string) error

		// Adds an item to the user/session's cart.
		// If there is no user or session, then a session is created and the sessionID is returned.
		AddItem(ctx context.Context, sessionID string, itemID string) (string, error)

		// Update item quantity in the user/session's cart
		// If there is no user or session, then a session is created and the sessionID is returned.
		UpdateItem(ctx context.Context, sessionID string, itemID string, quantity int) (string, error)

		// List socks that match any of the tags specified.  Sort the results in the specified order,
		// then return a subset of the results.
		ListItems(ctx context.Context, tags []string, order string, pageNum, pageSize int) ([]catalogue.Sock, error)

		// Gets details about a [Sock]
		GetSock(ctx context.Context, itemID string) (catalogue.Sock, error)

		// Lists all tags
		ListTags(ctx context.Context) ([]string, error)

		// Place an order for the specified items
		NewOrder(ctx context.Context, userID, addressID, cardID, cartID string) (order.Order, error)

		// Get all orders for a customer, sorted by date
		GetOrders(ctx context.Context, userID string) ([]order.Order, error)

		// Get an order by ID
		GetOrder(ctx context.Context, orderID string) (order.Order, error)

		// Log in to an existing user account.  Returns an error if the password
		// doesn't match the registered password
		// Returns the new session ID, which will be the user ID of the logged in user.
		Login(ctx context.Context, sessionID, username, password string) (string, user.User, error)

		// Register a new user account
		// Returns the new session ID, which will be the user ID of the registered user.
		Register(ctx context.Context, username, password, email, first, last string) (string, error)

		// Look up a user by customer ID
		GetUser(ctx context.Context, userID string) (user.User, error)

		// Look up an address by customer ID
		GetAddresses(ctx context.Context, userID string) ([]user.Address, error)

		// Adds a new address for a customer
		PostAddress(ctx context.Context, userID string, address user.Address) (string, error)

		// Look up a card by id.  If id is the empty string, returns all cards.
		GetCards(ctx context.Context, userID string) ([]user.Card, error)

		// Adds a new card for a customer
		PostCard(ctx context.Context, userID string, card user.Card) (string, error)
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
