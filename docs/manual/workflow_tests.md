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
    "github.com/blueprint-uservices/blueprint/runtime/plugins/simplecache"
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

* Black-box tests must be written in a separate module to the workflow.

Typically this is a module called `tests` residing in a sibling directory to `workflow`.  The `tests` module will need a dependency on the workflow module.

### Writing tests

Test files can be written following typical [Golang conventions](https://go.dev/doc/tutorial/add-a-test).  However, black-box tests should **not** manually instantiate services.  Instead, a [ServiceRegistry](../../runtime/core/registry/) should be used.

```
package tests

import (
    "testing"
    "github.com/blueprint-uservices/blueprint/runtime/core/registry"
)

var echoRegistry = registry.NewServiceRegistry[echo.EchoService]("echo")

func init() {
    echoRegistry.Register("local", func(ctx context.Context) (echo.EchoService, error) {
        return echo.NewEchoService(ctx)
    })
}

// A test for the EchoService
func TestEcho(t *testing.T) {
    ctx := context.Background()
    echo, err := echoRegistry.Get(ctx)
    
    // ... test code here ...
}
```

A static initialization block can be used to provide a "local" service; this code is equivalent to how the services were instantiated in the white-box tests.


```
package tests

import (
    "testing"
    "github.com/blueprint-uservices/blueprint/runtime/core/registry"
    "github.com/blueprint-uservices/blueprint/runtime/plugins/simplecache"
)

var multiEchoerRegistry = registry.NewServiceRegistry[echo.MultiEchoerService]("multi_echoer")

func init() {
    multiEchoerRegistry.Register("local", func(ctx context.Context) (echo.MultiEchoerService, error) {
        echo, err := NewEchoService(ctx)
        cache, err := simplecache.NewSimpleCache(ctx)
        return NewMultiEchoer(ctx, echo, cache)
    })
}

// A test for the MultiEchoerService
func TestMultiEchoer(t *testing.T) {
    ctx := context.Background()
    multi, err := multiEchoerRegistry.Get(ctx)
}
```

### Rules for defining gotests tests

* A `ServiceRegistry` is used for definining and subsequently instantiating services.
* The ServiceRegistry instance is declared as a ***package variable*** (e.g. `var echoRegistry = ...`)
* The ServiceRegistry is instantiated using `registry.NewServiceRegistry`
* Tests call `Get` on the service registry to acquire the service instances, e.g. `echo, err := echoRegistry.Get(ctx)`
* A package can contain multiple service registries for different service types, e.g. `echoRegistry` and `multiEchoerRegistry`

## Running Black-box Tests Directly

Like white-box tests, black-box tests can be directly invoked using `go test` in the source package.  For this to work, however, at least one service must be registered with the service registry, otherwise the tests will fail.

This can be done using a static initialization block to provide a "local" service; this service creation code is equivalent to how the services were instantiated in the white-box tests.  Running the tests directly will construct this "local" service and run the tests against it.

### Rules for running Black-box tests directly

* A static initialization block `init()` should register a "local" constructor to enable directly invoking the tests in the source package.

## Running Black-box Tests Against the Compiled Application

When Blueprint compiles an application, the [gotests plugin](../../plugins/gotests/) can be used to automatically convert black-box tests into tests that can run against the real, compiled application.  To do this, however, the tests require a "real" client to the application.  For example, if the compiled application deploys EchoService over gRPC, then the tests will need the gRPC client in order to contact the service.

The gotests plugin provides this functionality.  If enabled, it will inject a "real" client to the compiled applications into the black-box tests, e.g. it will add a gRPC client to the service registry.

Blueprint will auto-generate tests to `tests` in the output directory.  To run the compiled tests:
 1. start the compiled application
 2. invoke `go test` from the `tests` output package

Compiled tests may require additional command-line arguments depending on how the application is configured; for example, if the EchoService is deployed over gRPC, then the server address will be required.  If tests require additional arguments, they will fail and report which arguments are missing.

### Rules for compiling Black-box tests

* The `gotests` plugin must be used in the application's WiringSpec.  See [Wiring Spec](wiring.md) or the [gotests plugin documentation](../../plugins/gotests/) for more information.
* Compiled tests may require additional command-line arguments depending on how the application is configured.  See [Running an Application](running.md) for more information.  If tests require additional arguments, they will fail and report which arguments are missing.