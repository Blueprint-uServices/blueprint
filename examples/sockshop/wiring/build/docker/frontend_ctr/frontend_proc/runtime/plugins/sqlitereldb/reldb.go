// Package sqlitereldb implements a [backend.RelationalDB] using the in-memory Golang
// SQLite package [github.com/mattn/go-sqlite3].
//
// If you are directly running go code (e.g. not from a docker container), the go-sqlite3
// package requires CGO_ENABLED=1 and you must have gcc installed.  See [https://github.com/mattn/go-sqlite3]
// for more details about installation instructions.
package sqlitereldb

import (
	"context"
	"database/sql"

	"github.com/jmoiron/sqlx"

	_ "github.com/mattn/go-sqlite3"
)

// An in-memory relational DB that uses the go-sqlite3 package
type SqliteRelDB struct {
	db *sqlx.DB
}

// Instantiates a new [SqliteRelDB] instance that stores query data in-memory
func NewSqliteRelDB(ctx context.Context) (*SqliteRelDB, error) {
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
