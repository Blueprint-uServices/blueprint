package mysql

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

// Test requires a functional mysql instance to be already running
func TestRelDB(t *testing.T) {
	ctx := context.Background()

	db, err := NewMySqlDB(ctx, "127.0.0.1:3306", "TestRelDB", "root", "pass")
	require.NoError(t, err)

	batch := []string{
		`DELETE FROM address;`,
		`DELETE FROM user_addresses;`,
		`CREATE TABLE IF NOT EXISTS address (id INT PRIMARY KEY, street TEXT, street_number INT);`,
		`CREATE TABLE IF NOT EXISTS  user_addresses (address_id INT, user_id INT);`,
		`INSERT INTO address (id, street, street_number) VALUES (1, 'rue Victor Hugo', 32);`,
		`INSERT INTO address (id, street, street_number) VALUES (2, 'boulevard de la République', 23);`,
		`INSERT INTO address (id, street, street_number) VALUES (3, 'rue Charles Martel', 5);`,
		`INSERT INTO address (id, street, street_number) VALUES (4, 'chemin du bout du monde', 323);`,
		`INSERT INTO address (id, street, street_number) VALUES (5, 'boulevard de la liberté', 2);`,
		`INSERT INTO address (id, street, street_number) VALUES (6, 'avenue des champs', 12);`,
		`INSERT INTO user_addresses (address_id, user_id) VALUES (2, 1);`,
		`INSERT INTO user_addresses (address_id, user_id) VALUES (4, 1);`,
		`INSERT INTO user_addresses (address_id, user_id) VALUES (2, 2);`,
		`INSERT INTO user_addresses (address_id, user_id) VALUES (2, 3);`,
		`INSERT INTO user_addresses (address_id, user_id) VALUES (4, 4);`,
		`INSERT INTO user_addresses (address_id, user_id) VALUES (4, 5);`,
	}

	for _, b := range batch {
		_, err = db.Exec(ctx, b)
		t.Log(b)
		require.NoError(t, err)
	}

	query := `SELECT address.street_number, address.street FROM address 
							JOIN user_addresses ON address.id=user_addresses.address_id 
							WHERE user_addresses.user_id = ?;`
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
