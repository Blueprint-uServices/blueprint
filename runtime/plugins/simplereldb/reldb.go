package simplereldb

import (
	"context"
	"database/sql"

	"github.com/jmoiron/sqlx"
	"gitlab.mpi-sws.org/cld/blueprint/runtime/core/backend"

	_ "github.com/proullon/ramsql/driver"
)

// A simple in-memory relational DB that uses the ramsql package
type SimpleRelationalDB struct {
	db *sqlx.DB
}

// Instantiates a new [SimpleRelDB] instance that stores query data in-memory
func NewSimpleRelDB(ctx context.Context) (backend.RelationalDB, error) {
	db, err := sqlx.Open("ramsql", "SimpleRelationalDB")
	if err != nil {
		return nil, err
	}
	return &SimpleRelationalDB{db: db}, nil
}

// Exec implements backend.RelationalDB.
func (s *SimpleRelationalDB) Exec(ctx context.Context, query string, args ...any) (sql.Result, error) {
	return s.db.ExecContext(ctx, query, args...)
}

// Query implements backend.RelationalDB.
func (s *SimpleRelationalDB) Query(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	return s.db.QueryContext(ctx, query, args...)
}

// Get implements backend.RelationalDB.
func (s *SimpleRelationalDB) Get(ctx context.Context, dst interface{}, query string, args ...any) error {
	return s.db.GetContext(ctx, dst, query, args...)
}

// Prepare implements backend.RelationalDB.
func (s *SimpleRelationalDB) Prepare(ctx context.Context, query string) (*sql.Stmt, error) {
	return s.db.Prepare(query)
}

// Select implements backend.RelationalDB.
func (s *SimpleRelationalDB) Select(ctx context.Context, dst interface{}, query string, args ...any) error {
	return s.db.SelectContext(ctx, dst, query, args...)
}
