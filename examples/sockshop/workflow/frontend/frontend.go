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

func NewFrontend(ctx context.Context, user user.UserService, catalogue catalogue.CatalogueService, cart cart.CartService, order order.OrderService) (Frontend, error) {
	f := &frontend{
		user:      user,
		catalogue: catalogue,
		cart:      cart,
		order:     order,
	}
	return f, nil
}

// AddItem implements Frontend.
func (*frontend) AddItem(ctx context.Context, sessionID string, itemID string) (string, error) {
	panic("unimplemented")
}

// DeleteCart implements Frontend.
func (*frontend) DeleteCart(ctx context.Context, sessionID string) error {
	panic("unimplemented")
}

// DeleteItem implements Frontend.
func (*frontend) DeleteItem(ctx context.Context, sessionID string, itemID string) error {
	panic("unimplemented")
}

// GetAddresses implements Frontend.
func (*frontend) GetAddresses(ctx context.Context, userID string) ([]user.Address, error) {
	panic("unimplemented")
}

// GetCards implements Frontend.
func (*frontend) GetCards(ctx context.Context, userID string) ([]user.Card, error) {
	panic("unimplemented")
}

// GetCart implements Frontend.
func (*frontend) GetCart(ctx context.Context, sessionID string) ([]cart.Item, error) {
	panic("unimplemented")
}

// GetOrder implements Frontend.
func (*frontend) GetOrder(ctx context.Context, orderID string) (order.Order, error) {
	panic("unimplemented")
}

// GetOrders implements Frontend.
func (*frontend) GetOrders(ctx context.Context, userID string) ([]order.Order, error) {
	panic("unimplemented")
}

// GetSock implements Frontend.
func (*frontend) GetSock(ctx context.Context, itemID string) (catalogue.Sock, error) {
	panic("unimplemented")
}

// GetUser implements Frontend.
func (*frontend) GetUser(ctx context.Context, userID string) (user.User, error) {
	panic("unimplemented")
}

// ListItems implements Frontend.
func (*frontend) ListItems(ctx context.Context, tags []string, order string, pageNum int, pageSize int) ([]catalogue.Sock, error) {
	panic("unimplemented")
}

// ListTags implements Frontend.
func (*frontend) ListTags(ctx context.Context) ([]string, error) {
	panic("unimplemented")
}

// Login implements Frontend.
func (*frontend) Login(ctx context.Context, sessionID string, username string, password string) (string, user.User, error) {
	panic("unimplemented")
}

// NewOrder implements Frontend.
func (*frontend) NewOrder(ctx context.Context, userID string, addressID string, cardID string, cartID string) (order.Order, error) {
	panic("unimplemented")
}

// PostAddress implements Frontend.
func (*frontend) PostAddress(ctx context.Context, userID string, address user.Address) (string, error) {
	panic("unimplemented")
}

// PostCard implements Frontend.
func (*frontend) PostCard(ctx context.Context, userID string, card user.Card) (string, error) {
	panic("unimplemented")
}

// Register implements Frontend.
func (*frontend) Register(ctx context.Context, username string, password string, email string, first string, last string) (string, error) {
	panic("unimplemented")
}

// UpdateItem implements Frontend.
func (*frontend) UpdateItem(ctx context.Context, sessionID string, itemID string, quantity int) (string, error) {
	panic("unimplemented")
}
