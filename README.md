# blueprintv2

This repository is a work-in-progress refactor / reimplementation of the core wiring and IR of blueprint's compiler.

## Lifecycle

The lifecycle of a blueprint application is as follows.

1. Workflow spec is defined as before, as Golang services.
2. Wiring spec is defined as Go programs; see [cmd/example/main.go](cmd/example/main.go) for an example.  It is semantically similar to the wiring files from before, but now instead of interpreting the wiring file we directly execute it.  Many of the functions used to construct the application in the wiring file are syntatic sugar defined by the plugins being used.
3. The wiring spec is written using a dependency-injection style
4. Building the wiring spec will produce a blueprint application comprising typed nodes
5. Compiling the application entails compiling the root application node.

## Wiring Spec

To create a wiring spec, import blueprint and call

```
wiring := blueprint.NewWiringSpec("example")
```

The returned struct, `wiring`, is a **dependency injection container**.

## Plugins

Blueprint is a modular framework and most of its functionality is provided by plugins.  Plugins interact with the wiring spec and help to define dependencies in the dependency injection container.

For example, the `workflow` plugin provides a method `workflow.Define` to instantiate workflow spec services ([link](pkg/plugins/workflow/wiring.go)).  e.g.

```
workflow.Define(wiring, "serviceA", "LeafService")
```

The above snippet defines `serviceA` to be an instantiation of the `LeafService` from the application's workflow spec.

Different plugins do different things.  For example, the `memcached` plugin provides a method `PrebuiltProcess` to instantiate a prebuilt Memcached process ([link](pkg/plugins/memcached/wiring.go)).

```
memcached.PrebuiltProcess(wiring, "cacheB")
```

Some plugins wrap or modify nodes that are defined by other plugins.  For example, the `grpc` plugin provides a method `Deploy` to serve some named service using GRPC ([link](pkg/plugins/grpc/wiring.go)).

```
grpc.Deploy(wiring, "serviceA")
```

Similarly the `opentelemetry` plugin provides a method `Instrument` to wrap client and server code with OpenTelemetry instrumentation ([link](pkg/plugins/opentelemetry/wiring.go))

```
opentelemetry.Instrument(wiring, serviceName)
```

Almost all of the time, an application's wiring spec should interact with the Wiring struct via the utility methods offered by plugins.

## Building

Once the application has been defined, it can be built by calling

```
bp := wiring.GetBlueprint()
```

From here, different nodes of the application can be explicitly instantiated by name:

```
bp.Instantiate("serviceA")
```

Lastly, to build the application, we call

```
application, err := bp.Build()
```

This will actually invoke all of the build functions, check types and compatibility between nodes of the application, and return a Blueprint IR Node that represents the application as a whole.

# Working with the Wiring Spec

TODO: describe how plugins use the wiring spec

TODO: describe how plugins extend the IR

TODO: describe how the IR, once constructed, is then used for artifact generation

