# Workflow Tests

After writing some workflow services, you will want to write tests for those services.

## White-box unit tests

White-box unit tests can be written alongside the workflow spec using typical [Golang conventions](https://go.dev/doc/tutorial/add-a-test).  White-box unit tests can be run in the usual way using `go test`.

Services should be manually instantiated in white-box unit tests.  For services that use backends such as `backend.Cache` there are corresponding simple implementations that can be used, e.g. [simplecache](../../runtime/plugins/simplecache/cache.go) can be instantiated with `simplecache.NewSimpleCache`.

```
import (
    "testing"
    "gitlab.mpi-sws.org/cld/blueprint/runtime/plugins/simplecache"
)

// A test for the EchoService
func TestEcho(t *testing.T) {
    ctx := context.Background()
    echo, err := NewEchoService(ctx)
    
    // ... test code here ...
}

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

Notice that the code to instantiate the "local" instance of a service is similar to how the white-box test instantiates the service.


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

### Rules for gotests tests

* A service registry package variable is declared and instantiated using `registry.NewServiceRegistry`
* Tests acquire the echo service instance by calling `Get`, e.g. `echoRegistry.Get`
* To run the tests without compiling the application, implement a static initialization block `init()` and register a constructor.  This constructor can create a service instance directly, much like white-box tests.
* The `gotests` plugin must be used in the application's WiringSpec.  See [Wiring Spec](wiring.md) or the [gotests plugin documentation](../../plugins/gotests/) for more information.
* A package can contain multiple service registries for different service types. 

### Running tests

To run the tests without compiling the application, invoke `go test` as usual.  This will run the tests using a simple service instance much like white-box tests.

### Compiling tests

The `gotests` plugin must be used in the application's WiringSpec.  For this, see [Wiring Spec](wiring.md) or the [gotests plugin documentation](../../plugins/gotests/) for more information.

If the `gotests` plugin is wired, then during compilation, Blueprint will copy the black-box tests to the application's output and add code that automatically registers a "real" client to the service under test.  For example, if the compiled application deploys EchoService over gRPC, then the gotests plugin will update the code to use a gRPC client.

To run the compiled tests:
 1. start the compiled application
 2. invoke `go test` from the `tests` output package

Compiled tests may require additional command-line arguments depending on how the application is configured; for example, if the EchoService is deployed over gRPC, then the server address will be required.  The tests will fail and report the missing arguments if any are omitted.