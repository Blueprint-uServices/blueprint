package mysql

import (
	"context"
	"database/sql"

	"github.com/jmoiron/sqlx"

	_ "github.com/go-sql-driver/mysql"
)

// A relational DB that uses the mysql package
type MySqlDB struct {
	name string
	db   *sqlx.DB
}

// Instantiates a new [MySqlDB] instance that stores query data in a MySqlDB instance
func NewMySqlDB(ctx context.Context, addr string, name string, username string, password string) (*MySqlDB, error) {
	db, err := sqlx.Open("mysql", username+":"+password+"@tcp("+addr+")/")
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

	db, err = sqlx.Open("mysql", username+":"+password+"@tcp("+addr+")/"+name)
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

func (s *MySqlDB) Prepare(ctx context.Context, query string) (*sql.Stmt, error) {
	return s.db.PrepareContext(ctx, query)
}

func (s *MySqlDB) Select(ctx context.Context, dst interface{}, query string, args ...any) error {
	return s.db.SelectContext(ctx, dst, query, args...)
}

func (s *MySqlDB) Get(ctx context.Context, dst interface{}, query string, args ...any) error {
	return s.db.GetContext(ctx, dst, query, args...)
}
