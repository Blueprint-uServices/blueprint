# blueprintv2 

This repository is a work-in-progress refactor and reimplementation of several components of Blueprint.  v1 of Blueprint can be found on the [MPI-SWS GitLab cld/blueprint](https://gitlab.mpi-sws.org/cld/blueprint).

**While this repository remains a work-in-progress, all documentation assumes that you are familiar with Blueprint's concepts and the previous version of Blueprint.**

Shortcuts:
* [IR documentation](docs/ir.md)
* [Wiring documentation](docs/wiring.md)

## Rationale

This refactor focuses on four things:

* Improving the modularity of Blueprint's wiring, workflow, and plugins.
* Streamlining and simplifying the experience of developing a plugin and integrating it with Blueprint's wiring and IR
* Simplifying Blueprint's IR 
* Removing hacks and strictly enforcing abstractions

Thus the refactor is primarily affecting the wiring spec, the API for plugins to integrate with the wiring spec, the Blueprint IR, the process for generating artifacts, and the generated artifact code itself.

## Compatibility

Applications written against Blueprint's v1 workflow spec and workflow abstractions will remain compatible.  However, the files of an application might need to be re-organized.  The wiring spec of an application will need to be rewritten using Blueprint v2's go-based wiring API.

## Overview

In this version of Blueprint, we have the following pieces:

1. A **Workflow Spec** is defined as before, as Golang services, and using interfaces like Caches, Databases, etc. that are defined by plugins.
2. A **Wiring Spec** is defined that instantiates and places components of an application.  This aspect has significantly changed from Blueprint v1.  Before, a wiring spec was a python-like DSL.  Now, a wiring spec is simply a Go program that we directly execute to build the application.  While the syntax of the wiring spec has changed to be more go-like, conceptually the Go-style wiring spec remains similar to the previous DSL.  It is significantly more flexible and powerful now, however.
3. When the wiring spec is executed, a **Blueprint IR** representation of the application is constructed.  This is also changed significantly from Blueprint v1.  Before, the IR was built using a visitor-like pattern.  Now, Blueprint IR nodes contain direct references to other, typed nodes, making compilation a simpler and more direct process of invoking methods on nodes.
4. Most of the generated code looks the same as in Blueprint v1.  An exception, however, is in the code blocks that instantiate objects, for example in the main method of a process.  In this refactored code we generally favor a *dependency injection* style of instantiating objects in generated code, as it makes for cleaner output code as well as cleaner plugin implementations that generate code.

## Example

A simple example wiring spec can be found in [examples/leaf/wiring/main.go](examples/leaf/wiring/main.go).  The example:

Creates a new Blueprint wiring spec
```
wiring := blueprint.NewWiringSpec("example")
```

Defines a cache and some services that call each other
```
b_cache := memcached.PrebuiltProcess(wiring, "b_cache")
b := workflow.Define(wiring, "b", "LeafService", b_cache)
a := workflow.Define(wiring, "a", "nonLeafService", b)
```

Applies some default modifiers to the services
```
func serviceDefaults(wiring blueprint.WiringSpec, serviceName string) string {
	procName := fmt.Sprintf("p%s", serviceName)
	opentelemetry.Instrument(wiring, serviceName)
	grpc.Deploy(wiring, serviceName)
	return golang.CreateProcess(wiring, procName, serviceName)
}
pa := serviceDefaults(wiring, a)
pb := serviceDefaults(wiring, b)
```

Instructs Blueprint to instantiate `a` and `b`
```
bp := wiring.GetBlueprint()
bp.Instantiate(pa, pb)
```

And finally builds the application
```
application, err := bp.Build()
```

Blueprint recursively instantiates all dependencies, such that in the built application, `b_cache` is also instantiated.  If we run the application and print the generated IR nodes, our output will look like this (some output lines have been removed):

```
example = BlueprintApplication() {
  a.grpc.addr = Address(-> a.grpc_server)
  b.grpc.addr = Address(-> b.grpc_server)
  b_cache.addr = Address(-> b_cache.process)
  ot_collector.addr = Address(-> ot_collector.proc)

  pa = GolangProcessNode(a.grpc.addr, ot_collector.addr, b.grpc.addr, ot_collector.proc) {
    b.grpc_client = GRPCClient(b.grpc.addr)
    b.client.ot = OTClientWrapper(b.grpc_client, ot_collector.client)
    a = nonLeafService(b.client.ot)
    ot_collector.client = OTClient(ot_collector.addr)
    a.server.ot = OTServerWrapper(a, ot_collector.client)
    a.grpc_server = GRPCServer(a.server.ot, a.grpc.addr)
  }

  pb = GolangProcessNode(b.grpc.addr, ot_collector.addr, b_cache.addr, b_cache.process) {
    b_cache.client = MemcachedClient(b_cache.addr)
    b = LeafService(b_cache.client)
    ot_collector.client = OTClient(ot_collector.addr)
    b.server.ot = OTServerWrapper(b, ot_collector.client)
    b.grpc_server = GRPCServer(b.server.ot, b.grpc.addr)
  }

  ot_collector.proc = OTCollector(ot_collector.addr)
  b_cache.process = MemcachedProcess(b_cache.addr)
}
```


## Layout

The code in this repository is currently laid out as follows:

* `cmd` contains an example wiring spec
* `pkg/blueprint` contains the implementation of Blueprint itself
* `pkg/core` contains Blueprint plugins that we consider to be basic required functionality.  Each plugin resides in a subdirectory.
* `pkg/plugins` contains Blueprint all remaining Blueprint plugins.  Each plugin resides in a subdirectory.

## Anatomy of a plugin

A plugin typically, at a minimum, defines two types of functionality.  What is described below is convention rather than strict requirement:

* `ir.go` defines Blueprint IR nodes for the plugin.  For example, in the [GRPC ir.go](plugins/grpc/ir.go), two IR nodes are defined: one representing the GRPC server and one representing the GRPC client.  IR nodes implement interfaces such as `golang.Node` and `golang.CodeGenerator` which are used by other IR nodes.  For example, the `GolangServer` node contains a reference to some other node, `Wrapped`, that must be a `golang.Service` node.  See [docs/ir.md](docs/ir.md) for more information about the IR.
* `wiring.go` integrates the plugin into Blueprint's wiring spec.  Blueprint's wiring API can be called directly by wiring spec implementations, but in practice plugins can provide utility methods that simplify things.  For example, in the [GRPC wiring.go](plugins/grpc/wiring.go), a method `Deploy` exists; this can be called from a user's wiring spec to wrap a service in GRPC.  See [docs/wiring.md](docs/wiring.md) for more information about Wiring.
