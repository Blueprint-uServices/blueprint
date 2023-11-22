package tests

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"gitlab.mpi-sws.org/cld/blueprint/examples/sockshop/workflow/user"
	"gitlab.mpi-sws.org/cld/blueprint/runtime/core/registry"
	"gitlab.mpi-sws.org/cld/blueprint/runtime/plugins/simplenosqldb"
)

// Tests acquire a UserService instance using a service registry.
// This enables us to run local unit tests, whiel also enabling
// the Blueprint test plugin to auto-generate tests
// for different deployments when compiling an application.
var userServiceRegistry = registry.NewServiceRegistry[user.UserService]("user_service")

func init() {
	// If the tests are run locally, we fall back to this user service implementation
	userServiceRegistry.Register("local", func(ctx context.Context) (user.UserService, error) {
		db, err := simplenosqldb.NewSimpleNoSQLDB(ctx)
		if err != nil {
			return nil, err
		}

		return user.NewUserServiceImpl(ctx, db)
	})
}

var jon = user.User{
	FirstName: "Jonathan",
	LastName:  "Mace",
	Email:     "jon@mpi",
	Username:  "jon",
	Password:  "secret",
}

var vaastav = user.User{
	FirstName: "Vaastav",
	LastName:  "Anand",
	Email:     "vaastav@mpi",
	Username:  "vaastav",
	Password:  "supersecret",
}

var mpisb = user.Address{
	Street:   "Campus",
	Number:   "E1 5",
	Country:  "Germany",
	City:     "Saarbruecken",
	PostCode: "66123",
}

var amex = user.Card{
	LongNum: "378282246310005",
	Expires: "0530",
	CCV:     "123",
}

var visa = user.Card{
	LongNum: "4012888888881881",
	Expires: "0731",
	CCV:     "456",
}

var mpikl = user.Address{
	Street:   "Paul-Ehrlich-Strasse",
	Number:   "G 26",
	Country:  "Germany",
	City:     "Kaiserslautern",
	PostCode: "67663",
}

var deepak = user.User{
	FirstName: "Deepak",
	LastName:  "Garg",
	Email:     "deepak@mpi",
	Username:  "deepak",
	Password:  "supersupersecret",
	Addresses: []user.Address{mpisb, mpikl},
	Cards:     []user.Card{visa},
}

// We write the service test as a single test because we don't want to tear down and
// spin up the Mongo backends between tests, so state will persist in the database
// between tests.
func TestAll(t *testing.T) {

	ctx := context.Background()
	service, err := userServiceRegistry.Get(ctx)
	assert.NoError(t, err)

	{
		// Should be no users in the DB initially
		users, err := service.GetUsers(ctx, "")
		assert.NoError(t, err)
		assert.Len(t, users, 0)

		// Check we cannot login
		_, err = service.Login(ctx, jon.Username, jon.Password)
		assert.Error(t, err)
	}

	{
		// Register a user and check we can get it back
		u := jon
		uid, err := service.Register(ctx, u.Username, u.Password, u.Email, u.FirstName, u.LastName)
		assert.NoError(t, err)

		// Check the user we just registered
		expectUser(t, service, uid, jon)

		// Check we can login
		expectLogin(t, service, jon)

		// Check overall service state
		expectUsers(t, service, 1)
		expectAddresses(t, service, 0)
		expectCards(t, service, 0)
	}

	{
		// Register a second user and check we can get it back
		u := vaastav
		uid, err := service.Register(ctx, u.Username, u.Password, u.Email, u.FirstName, u.LastName)
		assert.NoError(t, err)

		// Check the user we just registered
		expectUser(t, service, uid, vaastav)

		// Check we can login
		expectLogin(t, service, jon)
		expectLogin(t, service, vaastav)

		// Check overall service state
		expectUsers(t, service, 2)
		expectAddresses(t, service, 0)
		expectCards(t, service, 0)
	}

	{
		// Register an address
		aid, err := service.PostAddress(ctx, deepak.Addresses[1])
		assert.NoError(t, err)

		// Check the address we just registered
		expectAddress(t, service, aid, deepak.Addresses[1])

		// Check overall service state
		expectUsers(t, service, 2)
		expectAddresses(t, service, 1)
		expectCards(t, service, 0)
	}

	{
		// Register a card
		cid, err := service.PostCard(ctx, deepak.Cards[0])
		assert.NoError(t, err)

		// Check the card we just registered
		expectCard(t, service, cid, deepak.Cards[0])

		// Check overall service state
		expectUsers(t, service, 2)
		expectAddresses(t, service, 1)
		expectCards(t, service, 1)
	}

	{
		// Post the third user and check we can get it back.
		uid, err := service.PostUser(ctx, deepak)
		assert.NoError(t, err)

		// Check the user we just registered
		u := expectUser(t, service, uid, deepak)

		// Check we can login
		expectLogin(t, service, deepak)
		expectLogin(t, service, jon)
		expectLogin(t, service, vaastav)

		// Explicitly check the cards and addresses
		expectCard(t, service, u.Cards[0].ID, deepak.Cards[0])
		expectAddress(t, service, u.Addresses[0].ID, deepak.Addresses[0])
		expectAddress(t, service, u.Addresses[1].ID, deepak.Addresses[1])

		// Check overall service state.  Addresses and cards get duplicated
		expectUsers(t, service, 3)
		expectAddresses(t, service, 3)
		expectCards(t, service, 2)
	}

	{
		// Delete an address
		u := expectLogin(t, service, deepak)

		toDelete := u.Addresses[0]
		err := service.Delete(ctx, "addresses", toDelete.ID)
		assert.NoError(t, err)

		// Check address was removed from system
		expectUsers(t, service, 3)
		expectAddresses(t, service, 2)
		expectCards(t, service, 2)

		// Log in again and check address was removed from user
		u, err = service.Login(ctx, deepak.Username, deepak.Password)
		assert.NoError(t, err)

		assert.Len(t, u.Addresses, 1)
		assert.Len(t, u.Cards, 1)

		expectCard(t, service, u.Cards[0].ID, deepak.Cards[0])
		expectAddress(t, service, u.Addresses[0].ID, deepak.Addresses[1])
	}

	{
		// Delete a user
		u := expectLogin(t, service, jon)
		err := service.Delete(ctx, "customers", u.UserID)
		assert.NoError(t, err)

		// Check user was removed from system
		expectUsers(t, service, 2)
		expectAddresses(t, service, 2)
		expectCards(t, service, 2)
	}

	{
		// Delete a card
		u, err := service.Login(ctx, deepak.Username, deepak.Password)
		assert.NoError(t, err)

		toDelete := u.Cards[0]
		err = service.Delete(ctx, "cards", toDelete.ID)
		assert.NoError(t, err)

		// Check card was removed from system
		expectUsers(t, service, 2)
		expectAddresses(t, service, 2)
		expectCards(t, service, 1)

		// Log in again and check card was removed from user
		u, err = service.Login(ctx, deepak.Username, deepak.Password)
		assert.NoError(t, err)

		assert.Len(t, u.Addresses, 1)
		assert.Len(t, u.Cards, 0)
		expectAddress(t, service, u.Addresses[0].ID, deepak.Addresses[1])

		// Delete the user
		err = service.Delete(ctx, "customers", u.UserID)
		assert.NoError(t, err)

		// Check user was removed from system
		expectUsers(t, service, 1)
		expectAddresses(t, service, 1)
		expectCards(t, service, 1)

		// Try to log in again, expect error
		_, err = service.Login(ctx, deepak.Username, deepak.Password)
		assert.Error(t, err)
	}
}

func expectUsers(t *testing.T, service user.UserService, expectedCount int) []user.User {
	// Get all users
	users, err := service.GetUsers(context.Background(), "")
	assert.NoError(t, err)
	assert.Len(t, users, expectedCount)
	return users
}

func expectCards(t *testing.T, service user.UserService, expectedCount int) []user.Card {
	// Get all cards
	cards, err := service.GetCards(context.Background(), "")
	assert.NoError(t, err)
	assert.Len(t, cards, expectedCount)
	return cards
}

func expectAddresses(t *testing.T, service user.UserService, expectedCount int) []user.Address {
	// Get all addresses
	addresses, err := service.GetAddresses(context.Background(), "")
	assert.NoError(t, err)
	assert.Len(t, addresses, expectedCount)
	return addresses
}

func expectUser(t *testing.T, service user.UserService, uid string, expected user.User) user.User {
	// Make sure the uid isn't the empty string
	assert.NotEmpty(t, uid)

	// Get the user
	users, err := service.GetUsers(context.Background(), uid)
	assert.NoError(t, err)
	assert.Len(t, users, 1)
	actual := users[0]

	// Check the user content
	matchUsers(t, expected, actual)
	assert.Equal(t, uid, actual.UserID)

	// Load and check addresses
	for i := range expected.Addresses {
		addressId := actual.Addresses[i].ID
		expectedAddress := expected.Addresses[i]
		expectAddress(t, service, addressId, expectedAddress)
	}

	// Load and check cards
	for i := range expected.Cards {
		cardId := actual.Cards[i].ID
		expectedCard := expected.Cards[i]
		expectCard(t, service, cardId, expectedCard)
	}

	return actual
}

func expectLogin(t *testing.T, service user.UserService, expected user.User) user.User {
	// Log in the user
	actual, err := service.Login(context.Background(), expected.Username, expected.Password)
	assert.NoError(t, err)

	// Check the user content
	matchUsers(t, expected, actual)

	// Check address data is already there
	for i := range expected.Addresses {
		matchAddresses(t, expected.Addresses[i], actual.Addresses[i])
	}

	// Check card data is already there (masked)
	for i := range expected.Cards {
		matchCards(t, expected.Cards[i], actual.Cards[i], true)
	}

	return actual
}

func expectAddress(t *testing.T, service user.UserService, addressId string, expected user.Address) user.Address {
	// Make sure the addressid isn't the empty string
	assert.NotEmpty(t, addressId)

	// Get the address
	addresses, err := service.GetAddresses(context.Background(), addressId)
	assert.NoError(t, err)
	assert.Len(t, addresses, 1)

	// Check the address content
	actual := addresses[0]
	matchAddresses(t, expected, actual)
	assert.Equal(t, addressId, actual.ID)
	return actual
}

func expectCard(t *testing.T, service user.UserService, cardid string, expected user.Card) user.Card {
	// Make sure the cardid isn't the empty string
	assert.NotEmpty(t, cardid)

	// Get the card
	cards, err := service.GetCards(context.Background(), cardid)
	assert.NoError(t, err)
	assert.Len(t, cards, 1)

	// Check the cards content
	actual := cards[0]
	matchCards(t, expected, actual, false)
	assert.Equal(t, cardid, actual.ID)
	return actual
}

func matchUsers(t *testing.T, expected user.User, actual user.User) {
	assert.Equal(t, expected.Username, actual.Username)
	assert.Equal(t, expected.FirstName, actual.FirstName)
	assert.Equal(t, expected.LastName, actual.LastName)
	assert.Equal(t, expected.Email, actual.Email)
	assert.Len(t, actual.Addresses, len(expected.Addresses))
	assert.Len(t, actual.Cards, len(expected.Cards))
}

func matchAddresses(t *testing.T, expected user.Address, actual user.Address) {
	assert.Equal(t, expected.Street, actual.Street)
	assert.Equal(t, expected.Number, actual.Number)
	assert.Equal(t, expected.Country, actual.Country)
	assert.Equal(t, expected.City, actual.City)
	assert.Equal(t, expected.PostCode, actual.PostCode)
}

func matchCards(t *testing.T, expected user.Card, actual user.Card, isMasked ...bool) {
	if len(isMasked) > 0 && isMasked[0] == true {
		l := len(actual.LongNum) - 4
		expectMasked := fmt.Sprintf("%v%v", strings.Repeat("*", l), actual.LongNum[l:])
		assert.Equal(t, expectMasked, actual.LongNum)
	} else {
		assert.Equal(t, expected.LongNum, actual.LongNum)
	}
	assert.Equal(t, expected.Expires, actual.Expires)
	assert.Equal(t, expected.CCV, actual.CCV)
}
