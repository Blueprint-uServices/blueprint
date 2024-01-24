// Package simple provides basic in-memory implementations of the Cache, Queue, NoSQLDB, and RelationalDB [backends]
// that are used by workflow services.
//
// The simple backend implementations are alternatives to the heavyweight "full system" implementations such as
// [memcached], [rabbitmq], [mongodb], [mysql], etc.
//
// The simple backend implementations are in-memory data structures; they must reside within the same process as the
// services that use them.
//
// # Wiring Spec Usage
//
// To instantiate a simple backend in your wiring spec, use the corresponding method for the backend type, giving
// the backend instance a name:
//
//	simple.NoSQLDB(spec, "my_nosql_db")
//	simple.RelationalDB(spec, "my_relational_db")
//	simple.Queue(spec, "my_queue")
//	simple.Cache(spec, "my_cache")
//
// After instantiating a backend, it can be provided as argument to a workflow service.
//
// # Wiring Spec Example
//
// Consider the [SockShop User Service] which makes use of a `backend.NoSQLDatabase`.  The service has the
// following constructor:
//
//	func NewUserServiceImpl(ctx context.Context, user_db backend.NoSQLDatabase) (UserService, error)
//
// In the wiring spec, we can instantiate the service and provide it with a simple NoSQLDB as follows:
//
//	user_db := simple.NoSQLDB(spec, "user_db")
//	user_service := workflow.Service[user.UserService](spec, "user_service", user_db)
//
// # Description
//
// The simple implementations are just in-memory data structures, so they can't be shared by services running in
// different processes.  You will encounter a compilation error if you attempt to do so.
//
// The simple implementations are primarily handy when developing and testing workflows, as they avoiding having
// to deploy full-fledged applications.  However, they do not necessarily implement all features (e.g. all operators
// of a query language), so in some cases they may be insufficient and you might need to resort to testing using
// proper backends.
//
// Implementations of the backends can be found in the following locations:
//   - NoSQLDB: [runtime/plugins/simplenosqldb]
//   - RelationalDB: [runtime/plugins/sqlitereldb]
//   - Queue: [runtime/plugins/simplequeue]
//   - Cache: [runtime/plugins/simplecache]
//
// [mongodb]: https://github.com/Blueprint-uServices/blueprint/tree/main/plugins/mongodb
// [backends]: https://github.com/Blueprint-uServices/blueprint/tree/main/runtime/core/backend
// [SockShop User Service]: https://github.com/Blueprint-uServices/blueprint/tree/main/examples/sockshop/workflow/user
// [memcached]: https://github.com/Blueprint-uServices/blueprint/tree/main/plugins/memcached
// [rabbitmq]: https://github.com/Blueprint-uServices/blueprint/tree/main/plugins/rabbitmq
// [mysql]: https://github.com/Blueprint-uServices/blueprint/tree/main/plugins/mysql
// [runtime/plugins/simplenosqldb]: https://github.com/Blueprint-uServices/blueprint/tree/main/runtime/plugins/simplenosqldb
// [runtime/plugins/sqlitereldb]: https://github.com/Blueprint-uServices/blueprint/tree/main/runtime/plugins/sqlitereldb
// [runtime/plugins/simplequeue]: https://github.com/Blueprint-uServices/blueprint/tree/main/runtime/plugins/simplequeue
// [runtime/plugins/simplecache]: https://github.com/Blueprint-uServices/blueprint/tree/main/runtime/plugins/simplecache
package simple

import (
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/coreplugins/pointer"
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/ir"
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/wiring"
	"github.com/blueprint-uservices/blueprint/runtime/core/backend"
	"github.com/blueprint-uservices/blueprint/runtime/plugins/simplecache"
	"github.com/blueprint-uservices/blueprint/runtime/plugins/simplenosqldb"
	"github.com/blueprint-uservices/blueprint/runtime/plugins/simplequeue"
	"github.com/blueprint-uservices/blueprint/runtime/plugins/sqlitereldb"
)

// [NoSQLDB] can be used by wiring specs to create an in-memory [backend.NoSQLDatabase] instance with the specified name.
// In the compiled application, uses the [simplenosqldb.SimpleNoSQLDB] implementation from the Blueprint runtime package
// The SimpleNoSQLDB has limited support for query and update operations.
func NoSQLDB(spec wiring.WiringSpec, name string) string {
	return define[backend.NoSQLDatabase, simplenosqldb.SimpleNoSQLDB](spec, name)
}

// [RelationalDB] can be used by wiring specs to create an in-memory [backend.RelationalDB] instance with the specified name.
// In the compiled application, uses the [sqlitereldb.SqliteRelDB] implementation from the Blueprint runtime package
// The compiled application might fail to run if gcc is not installed and CGO_ENABLED is not set.
func RelationalDB(spec wiring.WiringSpec, name string) string {
	return define[backend.RelationalDB, sqlitereldb.SqliteRelDB](spec, name)
}

// [Queue] can be used by wiring specs to create an in-memory [backend.Queue] instance with the specified name.
// In the compiled application, uses the [simplequeue.SimpleQueue] implementation from the Blueprint runtime package
func Queue(spec wiring.WiringSpec, name string) string {
	return define[backend.Queue, simplequeue.SimpleQueue](spec, name)
}

// [Cache] can be used by wiring specs to create an in-memory [backend.Cache] instance with the specified name.
// In the compiled application, uses the [simplecache.SimpleCache] implementation from the Blueprint runtime package
func Cache(spec wiring.WiringSpec, name string) string {
	return define[backend.Cache, simplecache.SimpleCache](spec, name)
}

func define[BackendInterface any, BackendImpl any](spec wiring.WiringSpec, name string) string {
	// The nodes that we are defining
	backendName := name + ".backend"

	// Define the backend instance
	spec.Define(backendName, &SimpleBackend{}, func(namespace wiring.Namespace) (ir.IRNode, error) {
		return newSimpleBackend[BackendImpl](name)
	})

	// Create a pointer to the backend instance
	pointer.CreatePointer[*SimpleBackend](spec, name, backendName)

	// Return the pointer; anybody who wants to access the backend instance should do so through the pointer
	return name
}
