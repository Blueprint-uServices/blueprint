# Blueprint Wiring

Blueprint's wiring is how an application is assembled out of its constituent pieces.  The underlying type `WiringSpec` is used to assemble a wiring spec, defined in [pkg/blueprint/wiring.go](pkg/blueprint/wiring.go).

## Getting Started

To begin constructing a wiring spec, create a new wiring spec, giving the application a name:

```
import "gitlab.mpi-sws.org/cld/blueprint/pkg/blueprint"
wiring := blueprint.NewWiringSpec("example")
```

## Typical Usage by Applications

The WiringSpec provides an API for building up an application, but the typical usage is not to use these methods directly, but to use plugin-specific utility methods that simplify things further.  Here are some example methods provided by different plugins.  For each plugin, these methods are by convention defined in the plugin's `wiring.go` file:

* `workflow.Define(serviceName, serviceType, serviceArgs...)` from the [workflow](pkg/plugins/workflow/wiring.go) plugin creates a service called `serviceName` that is an instance of `serviceType` which was defined in the application's workflow spec.  If the service takes arguments (e.g. to make calls to other services or to backends), the names of those other services can be provided as `serviceArgs`
* `opentelemetry.Instrument(serviceName)` from the [opentelemetry](pkg/plugins/opentelemetry/wiring.go) plugin will wrap existing service `serviceName` with an opentelemetry wrapper class that will create spans and propagate contexts
* `grpc.Deploy(serviceName)` from the [grpc](pkg/plugins/grpc/wiring.go) will deploy an existing service `serviceName` with GRPC such that callers to the service will now make RPC calls using a grpc client library
* `golang.CreateProcess(procName, children...)` from the [golang](pkg/plugins/golang/wiring.go) plugin creates a Golang process called `procName` that contains zero or more named child nodes `children`.  Children are typically services like `serviceName` defined with `workflow.Define`.
* `memcached.PrebuiltProcess(cacheName)` from the [memcached](pkg/plugins/memcached/wiring.go) plugin instantiates a standalone memcached process called `cacheName` that can be used by workflow services.

## Dependency Injection

In general, plugins define convenience methods for applications to build up the wiring spec.  Applications can directly call wiring spec methods, but this is more verbose and nuanced and generally not needed.

The general design of the Wiring spec is a dependency injection abstraction, extended to provide concepts of **hierarchy** and **addressing** (explained later).

Dependency injection is about controlling how objects get instantiated, and it is well suited to scenarios where one object doesn't care about the exact implementation of other objects.  For example, a service B might depend on a cache, but the exact cache implementation does not particularly matter -- it could be Redis, Memcached, or even an application-level dictionary.  This opaque interface between components is central to Blueprint, and thus dependency injection is a particularly well-suited design choice.

Using the WiringSpec there will be two distinct stages:

1. definitions are provided ***but not built*** -- for example, a service `foo` might be defined as being an instance of the `MyFooService` from the workflow spec, with dependencies to a service with name `bar`
2. once all definitions are provided, the actual nodes are built by invoking build functions

### Defining nodes

The first step of using the WiringSpec is to provide **definitions**.  Definitions describe *how* to build nodes and register a function that can be called to build the nodes, but in the first step those build functions are not yet invoked -- definitions are provided ***but not built***.  For example, we might define a service with the name `foo` as being an instance of the `MyFooService` from the workflow spec, with a dependency to a service named `bar`.

A definition thus comprises three pieces:

1. a name for the definition (e.g. `foo`)
2. a nodeType for the definition (e.g. `WorkflowSpecService`)
3. a build function for the definition, that is responsible for constructing the actual IR node representing `foo`

### Build Function

A build function has the following method signature:

```
func(blueprint.Scope) (blueprint.IRNode, error)
```

When invoked, this function is responsible for constructing and returning the IR node that represents `foo`.  In the case of a workflow spec service, it might look something like this:

```
func(scope blueprint.Scope) (blueprint.IRNode, error) {
    bar, err := scope.Get("bar")    // Get the node that represents `bar`
    if err != nil { return err }

    return newWorkflowService("foo", "MyFooService", bar)
}
```

Notice the argument `scope` of type `blueprint.Scope`.  In dependency injection parlance, this is a dependency injection container (however, we avoided overloading the word container, and use scope instead).  The purpose of the scope is to manage the process of building various nodes.  From within a build function, it is possible to look up the built nodes of dependencies like `bar` using the `scope.Get("bar")` function, where you pass in the name of the dependency that you want, and the scope will invoke the build function of that dependency then return the resulting built node.  Thus nodes are recursively instantiated following this chain of dependencies.  Instantiating the IR of an application as a whole entails building the root application node, which will then recursively get all dependencies before ultimately returning an IR node representing the application.

**Singletons.** One important aspect of scopes are that nodes within a scope are singletons -- if you call `scope.Get("bar")` multiple times, it will return the same instance each time.  Thus, within this scope, the build function of `bar` is only going to be called once.

