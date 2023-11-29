package mysql

import (
	"context"
	"database/sql"

	"gitlab.mpi-sws.org/cld/blueprint/runtime/core/backend"

	_ "github.com/go-sql-driver/mysql"
)

// A relational DB that uses the mysql package
type MySqlDB struct {
	name string
	db   *sql.DB
}

// Instantiates a new [MySqlDB] instance that stores query data in a MySqlDB instance
func NewMySqlDB(ctx context.Context, addr string, name string, username string, password string) (backend.RelationalDB, error) {
	db, err := sql.Open("mysql", username+":"+password+"@tcp("+addr+")/")
	if err != nil {
		return nil, err
	}

	if err != nil {
		return nil, err
	}

	_, err = db.Exec("CREATE DATABASE IF NOT EXISTS " + name)
	if err != nil {
		return nil, err
	}

	db.Close()

	db, err = sql.Open("mysql", username+":"+password+"@tcp("+addr+")/"+name)
	if err != nil {
		return nil, err
	}

	return &MySqlDB{name: name, db: db}, nil
}

// Exec implements backend.RelationalDB
func (s *MySqlDB) Exec(ctx context.Context, query string, args ...any) (sql.Result, error) {
	return s.db.ExecContext(ctx, query, args...)
}

// Query implements backend.RelationalDB
func (s *MySqlDB) Query(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	return s.db.QueryContext(ctx, query, args...)
}
