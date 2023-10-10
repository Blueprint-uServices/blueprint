package users

/*
Blueprint additional comment:

This code is directly translated from the UserService implementation on GitHub,
which includes several layers of database indirection.

All subsequent comments in this file are from the original code
*/

import (
	"context"
	"errors"
	"fmt"
	"net/url"

	"gitlab.mpi-sws.org/cld/blueprint/runtime/core/backend"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var (
	ErrInvalidHexID = errors.New("Invalid Id Hex")
)

type (
	DB struct {
		customers backend.NoSQLCollection
		cards     backend.NoSQLCollection
		addrs     backend.NoSQLCollection
	}

	// DBUser is a wrapper for the users
	DBUser struct {
		User       `bson:",inline"`
		ID         primitive.ObjectID   `bson:"_id"`
		AddressIDs []primitive.ObjectID `bson:"addresses"`
		CardIDs    []primitive.ObjectID `bson:"cards"`
	}

	// DBAddress is a wrapper for Address
	DBAddress struct {
		Address `bson:",inline"`
		ID      primitive.ObjectID `bson:"_id"`
	}

	// DBCard is a wrapper for Card
	DBCard struct {
		Card `bson:",inline"`
		ID   primitive.ObjectID `bson:"_id"`
	}
)

func NewDB(ctx context.Context, db backend.NoSQLDatabase) (*DB, error) {
	wrapper := &DB{}
	var err error

	wrapper.customers, err = db.GetCollection(ctx, "users", "customers")
	if err != nil {
		return nil, err
	}
	wrapper.cards, err = db.GetCollection(ctx, "users", "cards")
	if err != nil {
		return nil, err
	}
	wrapper.addrs, err = db.GetCollection(ctx, "users", "addresses")
	return wrapper, err
}

// New Returns a new DBUser
func NewDBUser() DBUser {
	u := NewUser()
	return DBUser{
		User:       u,
		AddressIDs: make([]primitive.ObjectID, 0),
		CardIDs:    make([]primitive.ObjectID, 0),
	}
}

// AddUserIDs adds userID as string to user
func (mu *DBUser) AddUserIDs() {
	if mu.User.Addresses == nil {
		mu.User.Addresses = make([]Address, 0)
	}
	for _, id := range mu.AddressIDs {
		mu.User.Addresses = append(mu.User.Addresses, Address{
			ID: id.Hex(),
		})
	}
	if mu.User.Cards == nil {
		mu.User.Cards = make([]Card, 0)
	}
	for _, id := range mu.CardIDs {
		mu.User.Cards = append(mu.User.Cards, Card{ID: id.Hex()})
	}
	mu.User.UserID = mu.ID.Hex()
}

// AddID ObjectID as string
func (m *DBAddress) AddID() {
	m.Address.ID = m.ID.Hex()
}

// AddID ObjectID as string
func (m *DBCard) AddID() {
	m.Card.ID = m.ID.Hex()
}

// CreateUser Insert user to DB, including connected addresses and cards, update passed in user with Ids
func (m *DB) CreateUser(ctx context.Context, u *User) error {
	id := primitive.NewObjectID()
	mu := NewDBUser()
	mu.User = *u
	mu.ID = id
	var carderr error
	var addrerr error
	mu.CardIDs, carderr = m.createCards(ctx, u.Cards)
	mu.AddressIDs, addrerr = m.createAddresses(ctx, u.Addresses)
	c := m.customers
	_, err := c.UpsertId(mu.ID, mu)
	if err != nil {
		// Gonna clean up if we can, ignore error
		// because the user save error takes precedence.
		m.cleanAttributes(ctx, mu)
		return err
	}
	mu.User.UserID = mu.ID.Hex()
	// Cheap err for attributes
	if carderr != nil || addrerr != nil {
		return fmt.Errorf("%v %v", carderr, addrerr)
	}
	*u = mu.User
	return nil
}

func (m *DB) createCards(ctx context.Context, cs []Card) ([]primitive.ObjectID, error) {
	ids := make([]primitive.ObjectID, 0)
	for k, ca := range cs {
		id := primitive.NewObjectID()
		mc := DBCard{Card: ca, ID: id}
		c := m.cards
		_, err := c.UpsertId(mc.ID, mc)
		if err != nil {
			return ids, err
		}
		ids = append(ids, id)
		cs[k].ID = id.Hex()
	}
	return ids, nil
}

func (m *DB) createAddresses(ctx context.Context, as []Address) ([]primitive.ObjectID, error) {
	ids := make([]primitive.ObjectID, 0)
	for k, a := range as {
		id := primitive.NewObjectID()
		ma := DBAddress{Address: a, ID: id}
		c := m.addrs
		_, err := c.UpsertId(ma.ID, ma)
		if err != nil {
			return ids, err
		}
		ids = append(ids, id)
		as[k].ID = id.Hex()
	}
	return ids, nil
}

func (m *DB) cleanAttributes(ctx context.Context, mu DBUser) error {
	err := m.addrs.DeleteMany(ctx, bson.D{{"_id", bson.D{{"$in", mu.AddressIDs}}}})
	if err != nil {
		return err
	}
	err = m.cards.DeleteMany(ctx, bson.D{{"_id", bson.D{{"$in", mu.CardIDs}}}})
	return err
}

func (m *DB) appendAttributeId(ctx context.Context, attr string, id primitive.ObjectID, userid string) error {
	userobjectid, err := primitive.ObjectIDFromHex(userid)
	if err != nil {
		return err
	}
	filter := bson.D{{"_id", userobjectid}}
	update := bson.D{{"$addToSet", bson.D{{attr, id}}}}
	return m.customers.UpdateOne(ctx, filter, update)
}

func (m *DB) removeAttributeId(ctx context.Context, attr string, id primitive.ObjectID, userid string) error {
	userobjectid, err := primitive.ObjectIDFromHex(userid)
	if err != nil {
		return err
	}
	filter := bson.D{{"_id", userobjectid}}
	update := bson.D{{"$pull", bson.D{{attr, id}}}}
	return m.customers.UpdateOne(ctx, filter, update)
}

// GetUserByName Get user by their name
func (m *DB) GetUserByName(ctx context.Context, name string) (User, error) {
	query := bson.D{{"username", name}}
	user := NewDBUser()
	cursor, err := m.customers.FindOne(ctx, query)
	if err != nil {
		return user.User, nil
	}
	err = cursor.One(ctx, &user)
	user.AddUserIDs()
	return user.User, err
}

// GetUser Get user by their object id
func (m *DB) GetUser(ctx context.Context, id string) (User, error) {
	userid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return NewUser(), errors.New("Invalid Id Hex")
	}
	query := bson.D{{"_id", userid}}
	user := NewDBUser()
	cursor, err := m.customers.FindOne(ctx, query)
	if err != nil {
		return user.User, nil
	}
	err = cursor.One(ctx, &user)
	user.AddUserIDs()
	return user.User, err
}

// GetUsers Get all users
func (m *DB) GetUsers(ctx context.Context) ([]User, error) {
	cursor, err := m.customers.FindMany(ctx, bson.D{})
	if err != nil {
		return nil, err
	}
	var mus []DBUser
	err = cursor.All(ctx, &mus)
	if err != nil {
		return nil, err
	}
	us := make([]User, 0, len(mus))
	for _, mu := range mus {
		mu.AddUserIDs()
		us = append(us, mu.User)
	}
	return us, err
}

// GetUserAttributes given a user, load all cards and addresses connected to that user
func (m *DB) GetUserAttributes(ctx context.Context, u *User) error {
	{
		ids := make([]primitive.ObjectID, 0, len(u.Addresses))
		for _, a := range u.Addresses {
			id, err := primitive.ObjectIDFromHex(a.ID)
			if err != nil {
				return ErrInvalidHexID
			}
			ids = append(ids, id)
		}

		cursor, err := m.addrs.FindMany(ctx, bson.D{{"_id", bson.D{{"$in", ids}}}})
		if err != nil {
			return err
		}
		var ma []DBAddress
		if err := cursor.All(ctx, &ma); err != nil {
			return err
		}

		na := make([]Address, 0, len(ma))
		for _, a := range ma {
			a.Address.ID = a.ID.Hex()
			na = append(na, a.Address)
		}
		u.Addresses = na
	}

	{
		ids := make([]primitive.ObjectID, 0, len(u.Cards))
		for _, c := range u.Cards {
			id, err := primitive.ObjectIDFromHex(c.ID)
			if err != nil {
				return ErrInvalidHexID
			}
			ids = append(ids, id)
		}

		cursor, err := m.cards.FindMany(ctx, bson.D{{"_id", bson.D{{"$in", ids}}}})
		if err != nil {
			return err
		}
		var mc []DBCard
		if err := cursor.All(ctx, &mc); err != nil {
			return err
		}

		nc := make([]Card, 0, len(mc))
		for _, ca := range mc {
			ca.Card.ID = ca.ID.Hex()
			nc = append(nc, ca.Card)
		}
		u.Cards = nc
	}
	return nil
}

// GetCard Gets card by objects Id
func (m *DB) GetCard(ctx context.Context, id string) (Card, error) {
	cardid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return Card{}, errors.New("Invalid Id Hex")
	}

	query := bson.D{{"_id", cardid}}
	cursor, err := m.cards.FindMany(ctx, query)
	if err != nil {
		return Card{}, err
	}

	var mc DBCard
	err = cursor.One(ctx, &mc)
	if err != nil {
		return Card{}, err
	}
	mc.AddID()
	return mc.Card, err
}

// GetCards Gets all cards
func (m *DB) GetCards(ctx context.Context) ([]Card, error) {
	cursor, err := m.cards.FindMany(ctx, bson.D{})
	if err != nil {
		return nil, err
	}

	var mcs []DBCard
	err = cursor.All(ctx, &mcs)
	if err != nil {
		return nil, err
	}

	cs := make([]Card, 0, len(mcs))
	for _, mc := range mcs {
		mc.AddID()
		cs = append(cs, mc.Card)
	}
	return cs, err
}

// CreateCard adds card to MongoDB
func (m *DB) CreateCard(ctx context.Context, ca *Card, userid string) error {
	if _, err := primitive.ObjectIDFromHex(userid); userid != "" && err != nil {
		return errors.New("Invalid Id Hex")
	}

	c := m.cards
	id := primitive.NewObjectID()
	mc := DBCard{Card: *ca, ID: id}
	_, err := c.UpsertId(mc.ID, mc)
	if err != nil {
		return err
	}
	// Address for anonymous user
	if userid != "" {
		err = m.appendAttributeId("cards", mc.ID, userid)
		if err != nil {
			return err
		}
	}
	mc.AddID()
	*ca = mc.Card
	return err
}

// GetAddress Gets an address by object Id
func (m *DB) GetAddress(ctx context.Context, id string) (Address, error) {
	s := m.Session.Copy()
	defer s.Close()
	if !bson.IsObjectIdHex(id) {
		return Address{}, errors.New("Invalid Id Hex")
	}
	c := m.addrs
	ma := DBAddress{}
	err := c.FindId(primitive.ObjectIDHex(id)).One(&ma)
	ma.AddID()
	return ma.Address, err
}

// GetAddresses gets all addresses
func (m *DB) GetAddresses(ctx context.Context) ([]Address, error) {
	// TODO: add pagination
	s := m.Session.Copy()
	defer s.Close()
	c := m.addrs
	var mas []DBAddress
	err := c.Find(nil).All(&mas)
	as := make([]Address, 0)
	for _, ma := range mas {
		ma.AddID()
		as = append(as, ma.Address)
	}
	return as, err
}

// CreateAddress Inserts Address into MongoDB
func (m *DB) CreateAddress(ctx context.Context, a *Address, userid string) error {
	if _, err := primitive.ObjectIDFromHex(userid); err != nil {
		return errors.New("Invalid Id Hex")
	}
	c := m.addrs
	id := primitive.NewObjectID()
	ma := DBAddress{Address: *a, ID: id}
	_, err := c.UpsertId(ma.ID, ma)
	if err != nil {
		return err
	}
	// Address for anonymous user
	if userid != "" {
		err = m.appendAttributeId("addresses", ma.ID, userid)
		if err != nil {
			return err
		}
	}
	ma.AddID()
	*a = ma.Address
	return err
}

// CreateAddress Inserts Address into MongoDB
func (m *DB) Delete(ctx context.Context, entity, id string) error {
	if _, err := primitive.ObjectIDFromHex(id); err != nil {
		return errors.New("Invalid Id Hex")
	}
	var c backend.NoSQLCollection
	switch entity {
	case "customers":
		c = m.customers
	case "cards":
		c = m.cards
	case "addresses":
		c = m.addrs
	}
	if entity == "customers" {
		u, err := m.GetUser(id)
		if err != nil {
			return err
		}
		aids := make([]primitive.ObjectID, 0)
		for _, a := range u.Addresses {
			aids = append(aids, primitive.ObjectIDHex(a.ID))
		}
		cids := make([]primitive.ObjectID, 0)
		for _, c := range u.Cards {
			cids = append(cids, primitive.ObjectIDHex(c.ID))
		}
		ac := m.addrs
		ac.RemoveAll(bson.M{"_id": bson.M{"$in": aids}})
		cc := m.cards
		cc.RemoveAll(bson.M{"_id": bson.M{"$in": cids}})
	} else {
		c := m.customers
		c.UpdateAll(bson.M{},
			bson.M{"$pull": bson.M{entity: primitive.ObjectIDHex(id)}})
	}
	return c.Remove(bson.M{"_id": primitive.ObjectIDHex(id)})
}

func getURL() url.URL {
	ur := url.URL{
		Scheme: "mongodb",
		Host:   host,
		Path:   db,
	}
	if name != "" {
		u := url.UserPassword(name, password)
		ur.User = u
	}
	return ur
}

// EnsureIndexes ensures username is unique
func (m *DB) EnsureIndexes() error {
	s := m.Session.Copy()
	defer s.Close()
	i := mgo.Index{
		Key:        []string{"username"},
		Unique:     true,
		DropDups:   true,
		Background: true,
		Sparse:     false,
	}
	c := m.customers
	return c.EnsureIndex(i)
}

func (m *DB) Ping() error {
	s := m.Session.Copy()
	defer s.Close()
	return s.Ping()
}
