package backend

import (
	"context"
	"database/sql"
)

// A Relational database backend is used for storing and querying structured data using SQL queries.
//
// Most golang relational databases implement the database/sql interfaces.  This interface exposes
// just a subset of that functionality.
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
}
