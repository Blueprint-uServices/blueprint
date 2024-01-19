package tests

import (
	"context"
	"fmt"
	"testing"

	"github.com/blueprint-uservices/blueprint/examples/sockshop/workflow/cart"
	"github.com/blueprint-uservices/blueprint/examples/sockshop/workflow/catalogue"
	"github.com/blueprint-uservices/blueprint/examples/sockshop/workflow/frontend"
	"github.com/blueprint-uservices/blueprint/examples/sockshop/workflow/user"
	"github.com/blueprint-uservices/blueprint/runtime/core/registry"
	"github.com/stretchr/testify/require"
	"golang.org/x/exp/slices"
)

// Tests acquire a Frontend instance using a service registry.
// This enables us to run local unit tests, while also enabling
// the Blueprint test plugin to auto-generate tests
// for different deployments when compiling an application.
var frontendRegistry = registry.NewServiceRegistry[frontend.Frontend]("frontend")

func init() {
	// If the tests are run locally, we fall back to this Frontend implementation
	frontendRegistry.Register("local", func(ctx context.Context) (frontend.Frontend, error) {
		user, err := userServiceRegistry.Get(ctx)
		if err != nil {
			return nil, err
		}

		cart, err := cartRegistry.Get(ctx)
		if err != nil {
			return nil, err
		}

		catalogue, err := catalogueRegistry.Get(ctx)
		if err != nil {
			return nil, err
		}

		order, err := ordersRegistry.Get(ctx)
		if err != nil {
			return nil, err
		}

		return frontend.NewFrontend(ctx, user, catalogue, cart, order)
	})
}

func TestFrontend(t *testing.T) {
	err := initCatalogue()
	require.NoError(t, err)

	ctx := context.Background()
	fe, err := frontendRegistry.Get(ctx)
	require.NoError(t, err)

	{
		// Load the catalogue
		items, err := fe.ListItems(ctx, nil, "", 1, 1000)
		require.NoError(t, err)
		require.True(t, socksequal(items, socks))
	}

	{
		// Get socks by tags
		for _, tag := range alltags {
			items, err := fe.ListItems(ctx, []string{tag}, "", 1, 1000)
			require.NoError(t, err)

			require.True(t, socksequal(items, getsocks(tag)), "ListItems tag=\"%v\"", tag)
		}
	}

	{
		// Get socks with 2 tags
		for _, tag1 := range alltags {
			for _, tag2 := range alltags {
				items, err := fe.ListItems(ctx, []string{tag1, tag2}, "", 1, 1000)
				require.NoError(t, err)

				require.True(t, socksequal(items, getsocks(tag1, tag2)), "ListItems tag1=\"%v\" tag2=\"%v\"", tag1, tag2)
			}
		}
	}

	{
		// Get socks with all tags
		items, err := fe.ListItems(ctx, alltags, "", 1, 1000)
		require.NoError(t, err)

		require.True(t, socksequal(items, socks), "ListItems tags=[%v]", alltags)
	}

	{
		// Get the tags
		dbtags, err := fe.ListTags(ctx)
		require.NoError(t, err)
		require.ElementsMatch(t, dbtags, alltags)
	}

	{
		// Get socks individually
		items, err := fe.ListItems(ctx, nil, "", 1, 1000)
		require.NoError(t, err)
		require.True(t, socksequal(items, socks))
		for _, sock := range items {
			sock2, err := fe.GetSock(ctx, sock.ID)
			require.NoError(t, err)
			require.True(t, sockequal(sock, sock2), "GetSock sock=\"%v\"", sock.Name)
		}
	}

	{
		// Get non-existent sock
		_, err := fe.GetSock(ctx, "hello world")
		require.Error(t, err)
	}

	{
		// Get non-existent tag
		items, err := fe.ListItems(ctx, []string{"nonexistent tag"}, "", 1, 1000)
		require.NoError(t, err)
		require.Empty(t, items)
	}

	username := "jon"
	password := "supersecret"

	{
		// Logging in should fail
		sessionID, _, err := fe.Login(ctx, "", username, password)
		require.Error(t, err)
		require.Equal(t, sessionID, "")
	}

	{
		// Get the catalogue
		items, err := fe.ListItems(ctx, nil, "", 1, 1000)
		require.NoError(t, err)
		require.True(t, socksequal(items, socks))

		// Add a sock to the cart
		sessionID, err := fe.AddItem(ctx, "", items[0].ID)
		require.NoError(t, err)
		require.NotEqual(t, "", sessionID)

		// Get the cart; should have 1 item
		{
			crt, err := fe.GetCart(ctx, sessionID)
			require.NoError(t, err)
			require.Len(t, crt, 1)
			require.Contains(t, crt, cart.Item{ID: items[0].ID, Quantity: 1, UnitPrice: items[0].Price})
		}

		// Add a few more socks
		for i := 0; i < 3; i++ {
			newSessionID, err := fe.AddItem(ctx, sessionID, items[0].ID)
			require.NoError(t, err)
			require.Equal(t, sessionID, newSessionID)
		}

		{
			newSessionID, err := fe.AddItem(ctx, sessionID, items[3].ID)
			require.NoError(t, err)
			require.Equal(t, sessionID, newSessionID)
		}

		{
			// Get and check the cart contents
			crt, err := fe.GetCart(ctx, sessionID)
			require.NoError(t, err)
			require.Len(t, crt, 2)
			require.Contains(t, crt, cart.Item{ID: items[0].ID, Quantity: 4, UnitPrice: items[0].Price})
			require.Contains(t, crt, cart.Item{ID: items[3].ID, Quantity: 1, UnitPrice: items[3].Price})
		}

		// Register a user
		userSessionID, err := fe.Register(ctx, sessionID, username, password, "my@email", "firstn", "lastn")
		require.NoError(t, err)
		require.NotEqual(t, sessionID, userSessionID)

		{
			// Check the cart was migrated to the user
			crt, err := fe.GetCart(ctx, userSessionID)
			require.NoError(t, err)
			require.Len(t, crt, 2)
			require.Contains(t, crt, cart.Item{ID: items[0].ID, Quantity: 4, UnitPrice: items[0].Price})
			require.Contains(t, crt, cart.Item{ID: items[3].ID, Quantity: 1, UnitPrice: items[3].Price})
		}

		// Update item quantity
		{
			newSessionID, err := fe.UpdateItem(ctx, userSessionID, items[0].ID, 2)
			require.NoError(t, err)
			require.Equal(t, userSessionID, newSessionID)
		}

		{
			// Get and check the cart contents
			crt, err := fe.GetCart(ctx, userSessionID)
			require.NoError(t, err)
			require.Len(t, crt, 2)
			require.Contains(t, crt, cart.Item{ID: items[0].ID, Quantity: 2, UnitPrice: items[0].Price})
			require.Contains(t, crt, cart.Item{ID: items[3].ID, Quantity: 1, UnitPrice: items[3].Price})
		}

		{
			// Get customer's orders
			orders, err := fe.GetOrders(ctx, userSessionID)
			require.NoError(t, err)
			require.Empty(t, orders)
		}

		{
			// Add a card and an address
			addressID, err := fe.PostAddress(ctx, userSessionID, user.Address{Street: "Home"})
			require.NoError(t, err)

			cardID, err := fe.PostCard(ctx, userSessionID, user.Card{LongNum: "1234123412341234", CCV: "574"})
			require.NoError(t, err)

			// Check they exist on the user
			u, err := fe.GetUser(ctx, userSessionID)
			require.NoError(t, err)
			require.Len(t, u.Cards, 1)
			require.Equal(t, cardID, u.Cards[0].ID)
			require.Len(t, u.Addresses, 1)
			require.Equal(t, addressID, u.Addresses[0].ID)

			// Check we can get the card
			crd, err := fe.GetCard(ctx, cardID)
			require.NoError(t, err)
			require.Equal(t, cardID, crd.ID)
			require.Equal(t, "1234123412341234", crd.LongNum)

			// Check we can get the address
			addr, err := fe.GetAddress(ctx, addressID)
			require.NoError(t, err)
			require.Equal(t, addressID, addr.ID)
			require.Equal(t, "Home", addr.Street)

			// Place an order
			ordr, err := fe.NewOrder(ctx, userSessionID, addressID, cardID, userSessionID)
			require.NoError(t, err)
			require.Equal(t, "Home", ordr.Address.Street)
			require.Equal(t, "1234123412341234", ordr.Card.LongNum)
			require.Len(t, ordr.Items, 2)
			require.Equal(t, 2*items[0].Price+items[3].Price+4.99, ordr.Total)
			require.Equal(t, userSessionID, ordr.CustomerID)

			// Cart should be empty
			crt, err := fe.GetCart(ctx, userSessionID)
			require.NoError(t, err)
			require.Empty(t, crt)

			// User should have 1 order
			orders, err := fe.GetOrders(ctx, userSessionID)
			require.NoError(t, err)
			require.Len(t, orders, 1)
			require.Equal(t, ordr, orders[0])
		}

		{
			// Delete the user
			usr, err := userServiceRegistry.Get(ctx)
			require.NoError(t, err)
			err = usr.Delete(ctx, "customers", userSessionID)
			require.NoError(t, err)
		}

	}

}

func initCatalogue() error {
	ctx := context.Background()
	s, err := catalogueRegistry.Get(ctx)
	if err != nil {
		return err
	}

	for _, sock := range socks {
		_, err := s.AddSock(ctx, sock)
		if err != nil {
			return fmt.Errorf("unable to add sock %v to catalogue due to %v", sock.Name, err.Error())
		}
	}

	return nil
}

func sock(name, description string, price float32, qty int, url1, url2 string, tags ...string) catalogue.Sock {
	return catalogue.Sock{Name: name, Description: description,
		Price: price, Quantity: qty, ImageURL_1: url1, ImageURL_2: url2, Tags: tags}
}

func sockmap(socks []catalogue.Sock) map[string]catalogue.Sock {
	sockmap := make(map[string]catalogue.Sock)
	for i := range socks {
		sockmap[socks[i].Name] = socks[i]
	}
	return sockmap
}

func socklist(names ...string) []catalogue.Sock {
	sockmap := sockmap(socks)
	socks := []catalogue.Sock{}
	for _, name := range names {
		if sock, exists := sockmap[name]; exists {
			socks = append(socks, sock)
		}
	}
	return socks
}

func getsocks(tags ...string) []catalogue.Sock {
	filtered := []catalogue.Sock{}
outer:
	for _, sock := range socks {
		for _, tag := range tags {
			if slices.Contains(sock.Tags, tag) {
				filtered = append(filtered, sock)
				continue outer
			}
		}
	}
	return filtered
}

var alltags = []string{"brown", "geek", "formal", "blue", "skin", "red", "action", "sport", "black", "magic", "green"}

var socks = []catalogue.Sock{
	sock("Weave special", "Limited issue Weave socks.", 17.15, 33, "/catalogue/images/weave1.jpg", "/catalogue/images/weave2.jpg", "geek", "black"),
	sock("Nerd leg", "For all those leg lovers out there. A perfect example of a swivel chair trained calf. Meticulously trained on a diet of sitting and Pina Coladas. Phwarr...", 7.99, 115, "/catalogue/images/bit_of_leg_1.jpeg", "/catalogue/images/bit_of_leg_2.jpeg", "blue", "skin"),
	sock("Crossed", "A mature sock, crossed, with an air of nonchalance.", 17.32, 738, "/catalogue/images/cross_1.jpeg", "/catalogue/images/cross_2.jpeg", "formal", "blue", "red", "action"),
	sock("SuperSport XL", "Ready for action. Engineers: be ready to smash that next bug! Be ready, with these super-action-sport-masterpieces. This particular engineer was chased away from the office with a stick.", 15.00, 820, "/catalogue/images/puma_1.jpeg", "/catalogue/images/puma_2.jpeg", "formal", "sport", "black"),
	sock("Holy", "Socks fit for a Messiah. You too can experience walking in water with these special edition beauties. Each hole is lovingly proggled to leave smooth edges. The only sock approved by a higher power.", 99.99, 1, "/catalogue/images/holy_1.jpeg", "/catalogue/images/holy_2.jpeg", "action", "magic"),
	sock("YouTube.sock", "We were not paid to sell this sock. It's just a bit geeky.", 10.99, 801, "/catalogue/images/youtube_1.jpeg", "/catalogue/images/youtube_2.jpeg", "geek", "formal"),
	sock("Figueroa", "enim officia aliqua excepteur esse deserunt quis aliquip nostrud anim", 14, 808, "/catalogue/images/WAT.jpg", "/catalogue/images/WAT2.jpg", "formal", "blue", "green"),
	sock("Classic", "Keep it simple.", 12, 127, "/catalogue/images/classic.jpg", "/catalogue/images/classic2.jpg", "brown", "green"),
	sock("Colourful", "proident occaecat irure et excepteur labore minim nisi amet irure", 18, 438, "/catalogue/images/colourful_socks.jpg", "/catalogue/images/colourful_socks.jpg", "brown", "blue"),
	sock("Cat socks", "consequat amet cupidatat minim laborum tempor elit ex consequat in", 15, 175, "/catalogue/images/catsocks.jpg", "/catalogue/images/catsocks2.jpg", "brown", "formal", "green"),
}

func socksequal(as, bs []catalogue.Sock) bool {
	if len(as) != len(bs) {
		return false
	}
	mapa := sockmap(as)
	mapb := sockmap(bs)
	if len(mapa) != len(mapb) {
		return false
	}
	for name, a := range mapa {
		if b, exists := mapb[name]; !exists || !sockequal(a, b) {
			return false
		}
	}
	return true
}

func sockequal(a, b catalogue.Sock) bool {
	if !(a.Name == b.Name && a.Price == b.Price && a.Quantity == b.Quantity && len(a.Tags) == len(b.Tags)) {
		return false
	}
	for i := range a.Tags {
		if !slices.Contains(b.Tags, a.Tags[i]) {
			return false
		}
		if !slices.Contains(a.Tags, b.Tags[i]) {
			return false
		}
	}
	return true
}
