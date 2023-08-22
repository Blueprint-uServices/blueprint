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

### Namespaces

A key aspect of Blueprint is its support for hierarchical namespaces -- that is, for example, an application might comprise a number of containers; within each container a number of processes; within a process a number of application-level objects.  Within Blueprint's IR, there are IR nodes to represent these concepts. For example, a golang process node contains golang object nodes for the objects within the process.

Namespace nodes like this are implemented by extending blueprint.Scope.  This is best explained through an example using the [golang process](pkg/plugins/golang) plugin as an example.  

Recall that a golang process is defined with a name, and it contains a number of golang objects.  For example, if we want to deploy `foo` in a process and expose it with GRPC, we might write the following wiring:

```
workflow.Define(wiring, "foo", "MyFooService")
grpc.Deploy(wiring, "foo")
golang.CreateProcess(wiring, "fooProc", "foo")
```

In this example we want to particularly focusing on the implementation of `golang.CreateProcess`.  This implementation makes use of a custom `golang.ProcessScope` which is defined in [scope.go](pkg/plugins/golang/scope.go) and used in [wiring.go](pkg/plugins/golang/wiring.go) in the definition of `CreateProcess`.

Its usage is rather straightforward.  In the build function for `fooProc`, the first thing that happens is to create a new scope as a child of the received scope:
```
process := NewGolangProcessScope(scope, wiring, procName)
```

Here, we say that `process` is a *child* scope of the received `scope`, and that `scope` is the *parent* scope of `process`.

Subsequently, in the build function of `fooProc`, there is a call to `Get("foo")` which will, internally, invoke foo's build function.  However, instead of getting foo from the parent `scope`, we get foo from the `process` scope: `process.Get("foo")`.  This has the same semantics: foo's build function will be called once, and the instance will be a singleton within the `process` scope.  

Internally, `process` will make note of nodes such as `foo` that get built, as well as any other nodes that `foo` builds recursively.  Eventually, once `foo` has been built, `fooProc`'s build function finishes.  It returns an IR node representing the process, that contains `foo` as well as any golang nodes that were built as a result of building `foo`.

**Types.** Scopes make use of IR nodeTypes to decide whether to, either, (a) build a node here, in this scope; or (b) just get the node from the parent scope instead.  Consider the example of a memcached container image.  It is nonsensical for a golang process to contain a memcached container image.  So the `golang.ProcessScope` does not support building nodes of type Container.  The Scope interface, which is defined in [pkg/blueprint/scope.go](pkg/blueprint/scope.go) and implemented by the [Golang plugin](pkg/plugins/golang/scope.go) allows scopes to specify which node types they actually support and can be built in this scope; for all other node types, a call to Get will just recursively call Get in the parent scope.  Recall, that in `wiring.Define`, one of the arguments is `nodeType` -- the type of node that this definition will build.
 that `golang.ProcessScope` *doesn't* support (e.g. to a memcached container image), then the node is instead gotten from the parent scope, recursively.  

**Singletons.** Although nodes are singletons within a scope, there can be multiple different scope instances, each with its own node instance.  For example, suppose we have services A, B, and C, all making RPC calls to service D over GRPC.  Then there will be an instance of D's GRPC client in each of A, B, and C's processes.

### Addresses

Sometimes, we don't want nodes to be immediately built, because we don't want them to live in the same scope as the caller.  Consider the example where two services, A and B, are both exposed over GRPC and deployed into a different process.  If A makes calls to B, then building A will require us to build the client library of B, and building the client library of B requires a corresponding server of B to make calls to.  We don't want to build the server of B in the same scope as A.

Addresses break this chain.  An address from A to B is little more than metadata that records the fact that an instance of B is required, but is not yet built.  For example, while building the client library of B, the GRPC client library will be built, followed by the address to the server, but at that point building ends.

When an address is defined, it is up to the plugin to decide the nodeType of the address -- that is, at what scope should it be built?  For most network addresses, they should be built in the root scope, because they should be accessible application-wide.

### Pointers

Pointers are a concept that are not directly part of Blueprint's wiring spec, but are built on top of it and widely used.  A pointer represents a chain of nodes, often with an address in the middle.  A pointer can be created for any defined node.  Modifers can be applied to pointers, which entails inserting extra nodes in the chain of nodes.  This is best illustrated by the [workflow plugin](pkg/plugins/workflow/wiring.go) which, for any service, defines both a handler and a pointer to the handler; and the [GRPC plugin](pkg/plugins/grpc/wiring.go), which adds client, server, and address nodes to a pointer.