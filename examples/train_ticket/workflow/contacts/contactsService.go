package contacts

import (
	"context"
	"errors"

	"github.com/blueprint-uservices/blueprint/runtime/core/backend"
	"go.mongodb.org/mongo-driver/bson"
)

// Contacts Service manages contacts for users
type ContactsService interface {
	// Find a contact using its `id`
	FindContactsById(ctx context.Context, id string) (Contact, error)
	// Find all contacts associated with an account with ID `id`
	FindContactsByAccountId(ctx context.Context, id string) ([]Contact, error)
	// Create a new contact
	CreateContacts(ctx context.Context, c Contact) error
	// Delete an existing contact
	Delete(ctx context.Context, c Contact) error
	// Get all existing contacts
	GetAllContacts(ctx context.Context) ([]Contact, error)
	// Modify an existing contact
	Modify(ctx context.Context, contact Contact) (bool, error)
}

type ContactsServiceImpl struct {
	contactsDB backend.NoSQLDatabase
}

func NewContactsServiceImpl(ctx context.Context, db backend.NoSQLDatabase) (*ContactsServiceImpl, error) {
	return &ContactsServiceImpl{contactsDB: db}, nil
}

func (c *ContactsServiceImpl) FindContactsById(ctx context.Context, id string) (Contact, error) {
	coll, err := c.contactsDB.GetCollection(ctx, "contacts", "contacts")
	if err != nil {
		return Contact{}, err
	}
	query := bson.D{{"id", id}}
	res, err := coll.FindOne(ctx, query)
	if err != nil {
		return Contact{}, err
	}
	var contact Contact
	exists, err := res.One(ctx, &contact)
	if err != nil {
		return Contact{}, err
	}
	if !exists {
		return Contact{}, errors.New("Contacts with id " + id + " does not exist!")
	}
	return contact, nil
}

func (c *ContactsServiceImpl) FindContactsByAccountId(ctx context.Context, id string) ([]Contact, error) {
	var account_contacts []Contact
	coll, err := c.contactsDB.GetCollection(ctx, "contacts", "contacts")
	if err != nil {
		return account_contacts, err
	}
	query := bson.D{{"accountid", id}}
	res, err := coll.FindMany(ctx, query)
	if err != nil {
		return account_contacts, err
	}
	err = res.All(ctx, &account_contacts)
	if err != nil {
		return account_contacts, err
	}
	return account_contacts, nil
}

func (c *ContactsServiceImpl) CreateContacts(ctx context.Context, contact Contact) error {
	coll, err := c.contactsDB.GetCollection(ctx, "contacts", "contacts")
	if err != nil {
		return err
	}
	query := bson.D{{"accountid", contact.AccountID}, {"documentnumber", contact.DocumentNumber}, {"documenttype", contact.DocumentType}}
	res, err := coll.FindOne(ctx, query)
	if err != nil {
		return err
	}
	var existing Contact
	exists, err := res.One(ctx, existing)
	if exists {
		return errors.New("Contact already exists")
	}
	if err != nil {
		return err
	}
	return coll.InsertOne(ctx, contact)
}

func (c *ContactsServiceImpl) Delete(ctx context.Context, contact Contact) error {
	coll, err := c.contactsDB.GetCollection(ctx, "contacts", "contacts")
	if err != nil {
		return err
	}
	query := bson.D{{"id", contact.ID}}
	return coll.DeleteOne(ctx, query)
}

func (c *ContactsServiceImpl) GetAllContacts(ctx context.Context) ([]Contact, error) {
	var all_contacts []Contact
	coll, err := c.contactsDB.GetCollection(ctx, "contacts", "contacts")
	if err != nil {
		return all_contacts, err
	}
	res, err := coll.FindMany(ctx, bson.D{})
	if err != nil {
		return all_contacts, err
	}
	err = res.All(ctx, &all_contacts)
	if err != nil {
		return all_contacts, err
	}
	return all_contacts, nil
}

func (c *ContactsServiceImpl) Modify(ctx context.Context, contact Contact) (bool, error) {
	coll, err := c.contactsDB.GetCollection(ctx, "contacts", "contacts")
	if err != nil {
		return false, err
	}
	query := bson.D{{"id", contact.ID}}
	return coll.Upsert(ctx, query, contact)
}
