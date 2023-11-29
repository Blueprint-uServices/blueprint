// Package sqlitereldb implements a [backend.RelationalDB] using in-memory SQLite.
//
// There are some pre-requisites for this to work.  CGO_ENABLED must be set to 1 and gcc
// must be installed.
package sqlitereldb

import (
	"context"
	"database/sql"

	"github.com/jmoiron/sqlx"
	"gitlab.mpi-sws.org/cld/blueprint/runtime/core/backend"

	_ "github.com/mattn/go-sqlite3"
)

// A simple in-memory relational DB that uses the ramsql package
type SqliteRelDB struct {
	db *sqlx.DB
}

// Instantiates a new [SimpleRelDB] instance that stores query data in-memory
func NewSqliteRelDB(ctx context.Context) (backend.RelationalDB, error) {
	db, err := sqlx.Open("sqlite3", "file:foobar?mode=memory&cache=shared")
	if err != nil {
		return nil, err
	}
	return &SqliteRelDB{db: db}, nil
}

// Exec implements backend.RelationalDB.
func (s *SqliteRelDB) Exec(ctx context.Context, query string, args ...any) (sql.Result, error) {
	return s.db.ExecContext(ctx, query, args...)
}

// Query implements backend.RelationalDB.
func (s *SqliteRelDB) Query(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	return s.db.QueryContext(ctx, query, args...)
}

// Get implements backend.RelationalDB.
func (s *SqliteRelDB) Get(ctx context.Context, dst interface{}, query string, args ...any) error {
	return s.db.GetContext(ctx, dst, query, args...)
}

// Prepare implements backend.RelationalDB.
func (s *SqliteRelDB) Prepare(ctx context.Context, query string) (*sql.Stmt, error) {
	return s.db.Prepare(query)
}

// Select implements backend.RelationalDB.
func (s *SqliteRelDB) Select(ctx context.Context, dst interface{}, query string, args ...any) error {
	return s.db.SelectContext(ctx, dst, query, args...)
}
