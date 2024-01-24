package user

import (
	"context"
	"errors"

	"github.com/blueprint-uservices/blueprint/runtime/core/backend"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Implements user service database interactions
type userStore struct {
	customers backend.NoSQLCollection
	cards     *cardStore
	addresses *addressStore
}

// The format of a User stored in the database
type dbUser struct {
	User       `bson:",inline"`
	ID         primitive.ObjectID   `bson:"_id"`
	AddressIDs []primitive.ObjectID `bson:"addresses"`
	CardIDs    []primitive.ObjectID `bson:"cards"`
}

func newUserStore(ctx context.Context, db backend.NoSQLDatabase) (*userStore, error) {
	users, err := db.GetCollection(ctx, "userservice", "users")
	if err != nil {
		return nil, err
	}

	cards, err := newCardStore(ctx, db)
	if err != nil {
		return nil, err
	}

	addresses, err := newAddressStore(ctx, db)
	if err != nil {
		return nil, err
	}

	store := &userStore{
		customers: users,
		cards:     cards,
		addresses: addresses,
	}

	return store, nil
}

// Generates database IDs for the user then adds to the database
func (s *userStore) createUser(ctx context.Context, user *User) error {
	u := dbUser{
		User:       *user,
		ID:         primitive.NewObjectID(),
		AddressIDs: []primitive.ObjectID{},
		CardIDs:    []primitive.ObjectID{},
	}
	var err error
	if u.CardIDs, err = s.cards.createCards(ctx, user.Cards); err != nil {
		return err
	}
	if u.AddressIDs, err = s.addresses.createAddresses(ctx, user.Addresses); err != nil {
		return err
	}
	_, err = s.customers.UpsertID(ctx, u.ID, u)
	if err != nil {
		// Gonna clean up if we can, ignore error
		// because the user save error takes precedence.
		s.addresses.removeAddresses(ctx, u.AddressIDs)
		s.cards.removeCards(ctx, u.CardIDs)
		return err
	}
	u.UserID = u.ID.Hex()
	*user = u.User
	return nil
}

// Get user by their name
func (s *userStore) getUserByName(ctx context.Context, username string) (User, error) {
	// Execute query
	cursor, err := s.customers.FindOne(ctx, bson.D{{"username", username}})
	if err != nil {
		return newUser(), err
	}

	// Extract query result
	u := dbUser{}
	if _, err := cursor.One(ctx, &u); err != nil {
		return newUser(), err
	}

	// Set the hex string IDs for the user, cards, and addresses before returning
	u.AddUserIDs()
	return u.User, nil
}

// Get user by their object id
func (s *userStore) getUser(ctx context.Context, userid string) (User, error) {
	// Convert user ID to bson object ID
	id, err := primitive.ObjectIDFromHex(userid)
	if err != nil {
		return newUser(), errors.New("Invalid Id Hex")
	}

	// Execute query
	cursor, err := s.customers.FindOne(ctx, bson.D{{"_id", id}})
	if err != nil {
		return newUser(), err
	}

	// Extract query result
	u := dbUser{}
	if _, err := cursor.One(ctx, &u); err != nil {
		return newUser(), err
	}

	// Set the hex string IDs for the user, cards, and addresses before returning
	u.AddUserIDs()
	return u.User, nil
}

// Get all users
func (s *userStore) getUsers(ctx context.Context) ([]User, error) {
	// Execute query
	cursor, err := s.customers.FindMany(ctx, bson.D{})
	if err != nil {
		return nil, err
	}

	// Extract query results
	dbUsers := []dbUser{}
	if err := cursor.All(ctx, &dbUsers); err != nil {
		return nil, err
	}

	// Convert from database users to user objects
	users := []User{}
	for _, dbUser := range dbUsers {
		dbUser.AddUserIDs()
		users = append(users, dbUser.User)
	}
	return users, nil
}

// Given a user, load all cards and addresses connected to that user
func (s *userStore) getUserAttributes(ctx context.Context, u *User) error {
	// Query the address store
	addresses, err := s.addresses.getAddresses(ctx, u.addressIDs())
	if err != nil {
		return err
	}

	// Query the card store
	cards, err := s.cards.getCards(ctx, u.cardIDs())
	if err != nil {
		return err
	}

	// Set the complete address and card data on the user
	u.Addresses = addresses
	u.Cards = cards
	return nil
}

// Adds a card to the cards DB and saves it for a user if there is a user
func (s *userStore) createCard(ctx context.Context, userid string, card *Card) error {
	if userid == "" {
		// An anonymous user; simply insert the card to the DB
		_, err := s.cards.createCard(ctx, card)
		return err
	}

	// A userid is provided; first check it's valid
	id, err := primitive.ObjectIDFromHex(userid)
	if err != nil {
		return errors.New("Invalid ID Hex")
	}

	// Insert the card to the DB
	cardID, err := s.cards.createCard(ctx, card)
	if err != nil {
		return err
	}

	// Update the user
	filter := bson.D{{"_id", id}}
	update := bson.D{{"$addToSet", bson.D{{"cards", cardID}}}}
	_, err = s.customers.UpdateOne(ctx, filter, update)
	return err
}

// Adds an address to the address DB and saves it for a user if there is a user
func (s *userStore) createAddress(ctx context.Context, userid string, address *Address) error {
	if userid == "" {
		// An anonymous user; simply insert the address to the DB
		_, err := s.addresses.createAddress(ctx, address)
		return err
	}

	// A userid is provided; first check it's valid
	id, err := primitive.ObjectIDFromHex(userid)
	if err != nil {
		return errors.New("Invalid ID Hex")
	}

	// Insert the address to the DB
	addressID, err := s.addresses.createAddress(ctx, address)
	if err != nil {
		return err
	}

	// Update the user
	filter := bson.D{{"_id", id}}
	update := bson.D{{"$addToSet", bson.D{{"addresses", addressID}}}}
	_, err = s.customers.UpdateOne(ctx, filter, update)
	return err
}

func (s *userStore) delete(ctx context.Context, entity string, id string) error {
	switch entity {
	case "customers":
		return s.deleteUser(ctx, id)
	case "addresses":
		return s.deleteAddress(ctx, id)
	case "cards":
		return s.deleteCard(ctx, id)
	default:
		return errors.New("Invalid entity " + entity)
	}
}

func (s *userStore) deleteUser(ctx context.Context, userid string) error {
	// Check valid user ID
	id, err := primitive.ObjectIDFromHex(userid)
	if err != nil {
		return errors.New("Invalid Id Hex")
	}

	// Get user details
	u, err := s.getUser(ctx, userid)
	if err != nil {
		return err
	}

	// Delete addresses
	addressIds, err := hexToObjectIds(u.addressIDs())
	if err != nil {
		return err
	}
	if err := s.addresses.removeAddresses(ctx, addressIds); err != nil {
		return err
	}

	// Delete cards
	cardIds, err := hexToObjectIds(u.cardIDs())
	if err != nil {
		return err
	}
	if err := s.cards.removeCards(ctx, cardIds); err != nil {
		return err
	}

	// Delete user
	return s.customers.DeleteMany(ctx, bson.D{{"_id", id}})
}

func (s *userStore) deleteAddress(ctx context.Context, addressid string) error {
	// Remove from customers db from any customers that have this address
	if err := s.deleteAttr(ctx, "addresses", addressid); err != nil {
		return err
	}

	// Remove from addresses db
	return s.addresses.removeAddress(ctx, addressid)
}

func (s *userStore) deleteCard(ctx context.Context, cardid string) error {
	// Remove from customers db from any customers that have this card
	if err := s.deleteAttr(ctx, "cards", cardid); err != nil {
		return err
	}

	// Remove from addresses db
	return s.cards.removeCard(ctx, cardid)
}

func (s *userStore) deleteAttr(ctx context.Context, attr, idhex string) error {
	// Check valid ID
	id, err := primitive.ObjectIDFromHex(idhex)
	if err != nil {
		return errors.New("Invalid Id Hex")
	}

	// Remove customer attr
	filter := bson.D{{attr, id}}
	update := bson.D{{"$pull", bson.D{{attr, id}}}}
	_, err = s.customers.UpdateMany(ctx, filter, update)
	return err
}

// Sets the user's ID to be the hex string of the database ObjectID.
// Also constructs (empty) Address and Card objects containing the IDs
// of the user's addresses and cards.
func (u *dbUser) AddUserIDs() {
	u.User.UserID = u.ID.Hex()
	u.User.Addresses = nil
	u.User.Cards = nil
	for _, id := range u.AddressIDs {
		u.Addresses = append(u.Addresses, Address{ID: id.Hex()})
	}
	for _, id := range u.CardIDs {
		u.Cards = append(u.Cards, Card{ID: id.Hex()})
	}
}

func (u *dbUser) convertAddressAndCardIDs() {
	for _, cardId := range u.CardIDs {
		u.Cards = append(u.Cards, Card{ID: cardId.Hex()})
	}
	for _, addressId := range u.AddressIDs {
		u.Addresses = append(u.Addresses, Address{ID: addressId.Hex()})
	}
}

// Converts bson object ids from hex strings to object representations
func hexToObjectIds(hexes []string) ([]primitive.ObjectID, error) {
	ids := make([]primitive.ObjectID, 0, len(hexes))
	for _, hex := range hexes {
		objectId, err := primitive.ObjectIDFromHex(hex)
		if err != nil {
			return nil, errors.New("Invalid Id Hex")
		}
		ids = append(ids, objectId)
	}
	return ids, nil
}
