# blueprintv2

This repository is a work-in-progress refactor / reimplementation of the core wiring and IR of blueprint's compiler.

## Lifecycle

The lifecycle of a blueprint application is as follows.

1. Workflow spec is defined as before, as Golang services.
2. Wiring spec is defined as Go programs; see [cmd/example/main.go] for an example.  It is semantically similar to the wiring files from before, but now instead of interpreting the wiring file we directly execute it.  Many of the functions used to construct the application in the wiring file are syntatic sugar defined by the plugins being used.
3. The wiring spec is written using a dependency-injection style
4. Building the wiring spec will produce a blueprint application comprising typed nodes
5. Compiling the application entails compiling the root application node.

## Wiring Spec

To create a wiring spec, import blueprint and call

```
wiring := blueprint.NewWiringSpec("example")
```

The returned struct, `wiring`, is a **dependency injection container**.  