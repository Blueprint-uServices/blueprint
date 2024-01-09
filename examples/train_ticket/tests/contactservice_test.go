package tests

import (
	"context"
	"fmt"
	"testing"

	"github.com/blueprint-uservices/blueprint/examples/train_ticket/workflow/contacts"
	"github.com/blueprint-uservices/blueprint/runtime/core/registry"
	"github.com/blueprint-uservices/blueprint/runtime/plugins/simplenosqldb"
	"github.com/stretchr/testify/require"
)

var contactsServiceRegistry = registry.NewServiceRegistry[contacts.ContactsService]("contacts_service")

func init() {
	contactsServiceRegistry.Register("local", func(ctx context.Context) (contacts.ContactsService, error) {

		db, err := simplenosqldb.NewSimpleNoSQLDB(ctx)
		if err != nil {
			return nil, err
		}

		return contacts.NewContactsServiceImpl(ctx, db)
	})
}

func genTestContactsData() []contacts.Contact {
	res := []contacts.Contact{}
	for i := 0; i < 10; i++ {
		c := contacts.Contact{
			ID:             fmt.Sprintf("c%d", i),
			AccountID:      fmt.Sprintf("a%d", i),
			Name:           fmt.Sprintf("contact%d", i),
			DocumentType:   int(contacts.ID_CARD),
			DocumentNumber: fmt.Sprintf("doc%d", i%2),
			PhoneNumber:    "555-0000",
		}
		res = append(res, c)
	}
	return res
}

func genTestContactsAccountData(accountID string) []contacts.Contact {
	res := []contacts.Contact{}
	doc_types := []int{int(contacts.ID_CARD), int(contacts.PASSPORT), int(contacts.OTHER)}
	for i, d := range doc_types {
		c := contacts.Contact{
			ID:             fmt.Sprintf("%s_c%d", accountID, i),
			AccountID:      accountID,
			Name:           fmt.Sprintf("contact%d", i),
			DocumentType:   d,
			DocumentNumber: fmt.Sprintf("doc%d", i),
			PhoneNumber:    "555-0001",
		}
		res = append(res, c)
	}
	return res
}

func TestContactsService(t *testing.T) {
	ctx := context.Background()
	service, err := contactsServiceRegistry.Get(ctx)
	require.NoError(t, err)

	testData := genTestContactsData()

	// Test Creation
	for _, c := range testData {
		err = service.CreateContacts(ctx, c)
		require.NoError(t, err)
	}

	// Check if things were correctly loaded
	cs, err := service.GetAllContacts(ctx)
	require.NoError(t, err)
	require.Len(t, cs, 10)

	// Individually find each contact using contact id
	for _, c := range testData {
		stored_c, err := service.FindContactsById(ctx, c.ID)
		require.NoError(t, err)
		requireContact(t, c, stored_c)
	}

	// Add all documents for a given account
	aid := "account_all"
	accData := genTestContactsAccountData(aid)
	for _, c := range accData {
		err = service.CreateContacts(ctx, c)
		require.NoError(t, err)
	}

	// Check if things were correctly loaded
	cs, err = service.GetAllContacts(ctx)
	require.NoError(t, err)
	require.Len(t, cs, 13)

	// Find documents by their id
	for _, c := range accData {
		stored_c, err := service.FindContactsById(ctx, c.ID)
		require.NoError(t, err)
		requireContact(t, c, stored_c)
	}

	// Find documents by account id
	cs, err = service.FindContactsByAccountId(ctx, aid)
	require.NoError(t, err)
	require.Len(t, cs, 3)

	// Test Modifications!
	for _, c := range accData {
		c.Name = "Modified" + c.Name
		res, err := service.Modify(ctx, c)
		require.NoError(t, err)
		require.True(t, res)

		// Check if the update actually took
		stored_c, err := service.FindContactsById(ctx, c.ID)
		require.NoError(t, err)
		requireContact(t, c, stored_c)
	}

	// Test adding duplicate contacts
	err = service.CreateContacts(ctx, testData[0])
	require.Error(t, err)

	// Test Deletion
	for _, c := range testData {
		err = service.Delete(ctx, c)
		require.NoError(t, err)
	}
	for _, c := range accData {
		err = service.Delete(ctx, c)
		require.NoError(t, err)
	}
}

func requireContact(t *testing.T, expected contacts.Contact, actual contacts.Contact) {
	require.Equal(t, expected.ID, actual.ID)
	require.Equal(t, expected.AccountID, actual.AccountID)
	require.Equal(t, expected.DocumentNumber, actual.DocumentNumber)
	require.Equal(t, expected.DocumentType, actual.DocumentType)
	require.Equal(t, expected.Name, actual.Name)
	require.Equal(t, expected.PhoneNumber, actual.PhoneNumber)
}
