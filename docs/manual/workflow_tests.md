# Workflow Tests

After writing some workflow services, you will want to write tests for those services.

## White-box unit tests

White-box unit tests can be written alongside the workflow spec using typical [Golang conventions](https://go.dev/doc/tutorial/add-a-test).  White-box unit tests can be run in the usual way using `go test`.

Services should be manually instantiated in white-box unit tests.

```
import (
    "context"
    "testing"
)

// A test for the EchoService
func TestEcho(t *testing.T) {
    ctx := context.Background()
    echo, err := NewEchoService(ctx)
    
    // ... test code here ...
}
```

For services that use backends such as `backend.Cache` there are corresponding simple implementations that can be used, e.g. [simplecache](../../runtime/plugins/simplecache/cache.go) can be instantiated with `simplecache.NewSimpleCache`.

```
import (
    "context"
    "testing"
    "gitlab.mpi-sws.org/cld/blueprint/runtime/plugins/simplecache"
)

// A test for the MultiEchoerService
func TestMultiEchoer(t *testing.T) {
    ctx := context.Background()
    echo, err := NewEchoService(ctx)
    cache, err := simplecache.NewSimpleCache(ctx)
    multi, err := NewMultiEchoer(ctx, echo, cache)

    // ... test code here ...
}
```


## Black-box unit tests

Blueprint's recommended approach to black-box unit tests is ***not*** the same as for white-box unit tests, because blueprint offers a [gotests plugin](../../plugins/gotests/) that automatically converts black-box unit tests into tests that can be run against a compiled Blueprint application.

The following sections describe how to structure black-box tests assuming you wish to make use of Blueprint's [gotests plugin](../../plugins/gotests/) plugin.  If you do not wish to make use of Blueprint's gotests plugin, then you can write Black-box unit tests in the same way as white-box unit tests.

### Module structure

Black-box tests should be written in a separate module (typically called `tests` in a sibling directory to `workflow`) and that tests module should have a dependency to the workflow module.

Test files can be written following typical [Golang conventions](https://go.dev/doc/tutorial/add-a-test).  However, tests should **not** manually instantiate services.  Instead, a [ServiceRegistry](../../runtime/core/registry/registry.go) should be used.

```
package tests

import (
    "testing"
    "gitlab.mpi-sws.org/cld/blueprint/runtime/core/registry"
)

var echoRegistry = registry.NewServiceRegistry[echo.EchoService]("echo")

func init() {
    echoRegistry.Register("local", func(ctx context.Context) (echo.EchoService, error) {
        return echo.NewEchoService(ctx)
    })
}

// A test for the EchoService
func TestEcho(t *testing.T) {
    echo, err := echoRegistry.Get(context.Background())
    
    // ... test code here ...
}
```

A static initialization block can be used to provide a "local" service; this code is equivalent to how the services were instantiated in the white-box tests.


```
package tests

import (
    "testing"
    "gitlab.mpi-sws.org/cld/blueprint/runtime/core/registry"
    "gitlab.mpi-sws.org/cld/blueprint/runtime/plugins/simplecache"
)

var multiEchoerRegistry = registry.NewServiceRegistry[echo.MultiEchoerService]("multi_echoer")

func init() {
    multiEchoerRegistry.Register("local", func(ctx context.Context) (echo.MultiEchoerService, error) {
        ctx := context.Background()
        echo, err := NewEchoService(ctx)
        cache, err := simplecache.NewSimpleCache(ctx)
        return NewMultiEchoer(ctx, echo, cache)
    })
}

// A test for the MultiEchoerService
func TestMultiEchoer(t *testing.T) {
    multi, err := multiEchoerRegistry.Get(context.Background())
}
```

### Rules for defining gotests tests

* A `ServiceRegistry` is used for definining and subsequently instantiating services.
* The ServiceRegistry is declared as a ***package variable** (e.g. `echoVegistry`)
* The ServiceRegistry is instantiated using `registry.NewServiceRegistry`
* A test acquires a service instance by calling `Get`, e.g. `echoRegistry.Get`
* A package can contain multiple service registries for different service types. 

## Running Black-box Tests Directly

The above tests can be run directly, much like white-box tests, by invoking `go test`.  This is because in addition to using the ServiceRegistry, the tests also define the constructor for a simple "local" service.  Running `go test` directly will construct this "local" service and run the tests against it.

### Rules for running Black-box tests directly

* A static initialization block `init()` can register a "local" constructor, for running the tests in white-box style.

## Running Black-box Tests Against the Compiled Application

When Blueprint compiles an application, it can automatically convert black-box tests into tests that can run against the real, compiled application.  To do this, however, the tests require a "real" client to the application.  For example, if the compiled application deploys EchoService over gRPC, then the tests will need the gRPC client in order to contact the service.

The gotests plugin provides this functionality.  If enabled, it will modify the black-box tests to register a client for the compiled application (e.g. it will add a gRPC client to the service registry).

Blueprint will auto-generate tests to `tests` in the output directory.  To run the compiled tests:
 1. start the compiled application
 2. invoke `go test` from the `tests` output package

Compiled tests may require additional command-line arguments depending on how the application is configured; for example, if the EchoService is deployed over gRPC, then the server address will be required.  If tests require additional arguments, they will fail and report which arguments are missing.

### Rules for compiling Black-box tests

* The `gotests` plugin must be used in the application's WiringSpec.  See [Wiring Spec](wiring.md) or the [gotests plugin documentation](../../plugins/gotests/) for more information.
* Compiled tests may require additional command-line arguments depending on how the application is configured.  See [Running an Application](running.md) for more information.  If tests require additional arguments, they will fail and report which arguments are missing.