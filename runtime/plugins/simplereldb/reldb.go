package simplereldb

import (
	"context"
	"database/sql"

	"gitlab.mpi-sws.org/cld/blueprint/runtime/core/backend"

	_ "github.com/proullon/ramsql/driver"
)

// A simple in-memory relational DB that uses the ramsql package
type SimpleRelationalDB struct {
	name string
	db   *sql.DB
}

// Instantiates a new [SimpleRelDB] instance that stores query data in-memory
func NewSimpleRelDB(ctx context.Context, name string) (backend.RelationalDB, error) {
	db, err := sql.Open("ramsql", "TestLoadUserAddresses")
	if err != nil {
		return nil, err
	}
	return &SimpleRelationalDB{name: name, db: db}, nil
}

// Exec implements backend.RelationalDB.
func (s *SimpleRelationalDB) Exec(ctx context.Context, query string, args ...any) (sql.Result, error) {
	return s.db.ExecContext(ctx, query, args...)
}

// Query implements backend.RelationalDB.
func (s *SimpleRelationalDB) Query(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	return s.db.QueryContext(ctx, query, args...)
}