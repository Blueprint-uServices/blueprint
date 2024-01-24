// Package backend provides the interfaces for common backends like caches, queues, databases, etc. that are often
// used by application workflow specs.
//
// Workflow services can, and should, make use of the interfaces defined in this package.
//
// To use a backend, an application's workflow should require this module and import the interfaces from this package.
// Service constructors should receive the backend interface as an argument, e.g.
//
//	func NewMyService(ctx context.Context, db backend.NoSQLDB) (MyService, error) {...}
package backend
