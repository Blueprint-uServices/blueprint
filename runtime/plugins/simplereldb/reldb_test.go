package simplereldb

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRelDB(t *testing.T) {
	ctx := context.Background()

	db, err := NewSimpleRelDB(ctx, "TestRelDB")
	require.NoError(t, err)

	batch := []string{
		`CREATE TABLE IF NOT EXISTS address (id BIGSERIAL PRIMARY KEY, street TEXT, street_number INT);`,
		`CREATE TABLE IF NOT EXISTS  user_addresses (address_id INT, user_id INT);`,
		`INSERT INTO address (street, street_number) VALUES ('rue Victor Hugo', 32);`,
		`INSERT INTO address (street, street_number) VALUES ('boulevard de la République', 23);`,
		`INSERT INTO address (street, street_number) VALUES ('rue Charles Martel', 5);`,
		`INSERT INTO address (street, street_number) VALUES ('chemin du bout du monde', 323);`,
		`INSERT INTO address (street, street_number) VALUES ('boulevard de la liberté', 2);`,
		`INSERT INTO address (street, street_number) VALUES ('avenue des champs', 12);`,
		`INSERT INTO user_addresses (address_id, user_id) VALUES (2, 1);`,
		`INSERT INTO user_addresses (address_id, user_id) VALUES (4, 1);`,
		`INSERT INTO user_addresses (address_id, user_id) VALUES (2, 2);`,
		`INSERT INTO user_addresses (address_id, user_id) VALUES (2, 3);`,
		`INSERT INTO user_addresses (address_id, user_id) VALUES (4, 4);`,
		`INSERT INTO user_addresses (address_id, user_id) VALUES (4, 5);`,
	}

	for _, b := range batch {
		_, err = db.Exec(ctx, b)
		require.NoError(t, err)
	}

	query := `SELECT address.street_number, address.street FROM address 
							JOIN user_addresses ON address.id=user_addresses.address_id 
							WHERE user_addresses.user_id = $1;`
	userID := 1
	rows, err := db.Query(ctx, query, userID)
	require.NoError(t, err)

	var number int
	var street string

	require.True(t, rows.Next())
	require.NoError(t, rows.Scan(&number, &street))
	require.Equal(t, 23, number)
	require.Equal(t, "boulevard de la République", street)

	require.True(t, rows.Next())
	require.NoError(t, rows.Scan(&number, &street))
	require.Equal(t, 323, number)
	require.Equal(t, "chemin du bout du monde", street)

	require.False(t, rows.Next())
}
