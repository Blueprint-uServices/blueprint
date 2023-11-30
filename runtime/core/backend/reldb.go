package backend

import (
	"context"
	"database/sql"
)

// A Relational database backend is used for storing and querying structured data using SQL queries.
//
// SQL is relatively standardized in golang under the database/sql interfaces.  Blueprint's [RelationalDB]
// interface exposes the github.com/jmoiron/sqlx interfaces, which are more convenient for casual usage
// and help in marshalling structs into rows and back.
type RelationalDB interface {
	// Exec executes a query without returning any rows. The args are for any placeholder parameters in the query.
	//
	// Returns a [sql.Result] object from the [database/sql] package.
	Exec(ctx context.Context, query string, args ...any) (sql.Result, error)

	// Query executes a query that returns rows, typically a SELECT. The args are for any placeholder parameters in the query.
	//
	// Returns a [sql.Rows] object from the [database/sql] package, that can be used to access query results.
	// Rows' cursor starts before the first row of the result set. Use Next to advance from row to row.
	Query(ctx context.Context, query string, args ...any) (*sql.Rows, error)

	// Prepare creates a prepared statement for later queries or executions. Multiple queries or executions may
	// be run concurrently from the returned statement. The caller must call the statement's Close method
	// when the statement is no longer needed.
	//
	// Returns a [sql.Stmt] object from the [database/sql] package, that can be used to execute the prepared
	// statement.
	Prepare(ctx context.Context, query string) (*sql.Stmt, error)

	// Select using this DB. Any placeholder parameters are replaced with supplied args.
	//
	// Uses [github.com/jmoiron/sqlx] to marshal query results into dst.
	Select(ctx context.Context, dst interface{}, query string, args ...any) error

	// Get using this DB. Any placeholder parameters are replaced with supplied args. An error is returned if the result set is empty.
	//
	// Uses [github.com/jmoiron/sqlx] to marshal query results into dst.
	Get(ctx context.Context, dst interface{}, query string, args ...any) error
}
