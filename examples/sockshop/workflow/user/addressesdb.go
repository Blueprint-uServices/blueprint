package user

import (
	"context"
	"errors"

	"github.com/blueprint-uservices/blueprint/runtime/core/backend"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type addressStore struct {
	c backend.NoSQLCollection
}

// The format of an Address stored in the database
type dbAddress struct {
	Address `bson:",inline"`
	ID      primitive.ObjectID `bson:"_id"`
}

func newAddressStore(ctx context.Context, db backend.NoSQLDatabase) (*addressStore, error) {
	c, err := db.GetCollection(ctx, "userservice", "addresses")
	return &addressStore{c: c}, err
}

// Gets an address by object Id
func (s *addressStore) getAddress(ctx context.Context, addressid string) (Address, error) {
	// Convert the address ID
	id, err := primitive.ObjectIDFromHex(addressid)
	if err != nil {
		return Address{}, errors.New("Invalid ID Hex")
	}

	// Run the query
	cursor, err := s.c.FindOne(ctx, bson.D{{"_id", id}})
	if err != nil {
		return Address{}, err
	}
	address := dbAddress{}
	_, err = cursor.One(ctx, &address)

	// Convert from DB address data to Address object
	address.Address.ID = address.ID.Hex()
	return address.Address, err
}

// Gets addresses from the address store
func (s *addressStore) getAddresses(ctx context.Context, addressIds []string) ([]Address, error) {
	if len(addressIds) == 0 {
		return nil, nil
	}

	// Convert the address IDs from hex strings to objects
	ids, err := hexToObjectIds(addressIds)
	if err != nil {
		return nil, err
	}

	// Run the query
	cursor, err := s.c.FindMany(ctx, bson.D{{"_id", bson.D{{"$in", ids}}}})
	if err != nil {
		return nil, err
	}
	dbAddresses := make([]dbAddress, 0, len(addressIds))
	err = cursor.All(ctx, &dbAddresses)

	// Convert from DB address data to Address objects
	addresses := make([]Address, 0, len(dbAddresses))
	for _, address := range dbAddresses {
		address.Address.ID = address.ID.Hex()
		addresses = append(addresses, address.Address)
	}

	return addresses, err
}

func (s *addressStore) getAllAddresses(ctx context.Context) ([]Address, error) {
	// Run the query
	cursor, err := s.c.FindMany(ctx, bson.D{})
	if err != nil {
		return nil, err
	}
	dbAddresses := make([]dbAddress, 0)
	err = cursor.All(ctx, &dbAddresses)

	// Convert from DB address data to Address objects
	addresses := make([]Address, 0, len(dbAddresses))
	for _, address := range dbAddresses {
		address.Address.ID = address.ID.Hex()
		addresses = append(addresses, address.Address)
	}

	return addresses, err
}

// Adds an address to the address DB
func (s *addressStore) createAddress(ctx context.Context, address *Address) (primitive.ObjectID, error) {
	// Create and insert to DB
	dbaddress := dbAddress{Address: *address, ID: primitive.NewObjectID()}
	if _, err := s.c.UpsertID(ctx, dbaddress.ID, dbaddress); err != nil {
		return dbaddress.ID, err
	}

	// Update the provided address
	dbaddress.Address.ID = dbaddress.ID.Hex()
	*address = dbaddress.Address
	return dbaddress.ID, nil
}

// Creates or updates the provided addresses in the addressStore.
func (s *addressStore) createAddresses(ctx context.Context, addresses []Address) ([]primitive.ObjectID, error) {
	if len(addresses) == 0 {
		return []primitive.ObjectID{}, nil
	}
	createdIds := make([]primitive.ObjectID, 0)
	for _, address := range addresses {
		toInsert := dbAddress{
			Address: address,
			ID:      primitive.NewObjectID(),
		}
		_, err := s.c.UpsertID(ctx, toInsert.ID, toInsert)
		if err != nil {
			return createdIds, err
		}
		createdIds = append(createdIds, toInsert.ID)
	}

	return createdIds, nil
}

func (s *addressStore) removeAddress(ctx context.Context, addressid string) error {
	// Convert the address ID
	id, err := primitive.ObjectIDFromHex(addressid)
	if err != nil {
		return errors.New("Invalid ID Hex")
	}
	return s.removeAddresses(ctx, []primitive.ObjectID{id})
}

func (s *addressStore) removeAddresses(ctx context.Context, ids []primitive.ObjectID) error {
	return s.c.DeleteMany(ctx, bson.D{{"_id", bson.D{{"$in", ids}}}})
}

// Set the address's ID to be the hex string of the database ObjectID
func (a *dbAddress) addID() {
	a.Address.ID = a.ID.Hex()
}
