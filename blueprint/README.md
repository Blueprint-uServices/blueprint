# Blueprint

This package contains the core of Blueprint's compiler abstractions.

`pkg/blueprint` has definitions of the WiringSpec API, the base types of Blueprint's IR, and the core of Blueprint's building and compilation process.

`pkg/core` has plugin implementations that are sufficiently important that we consider them core plugins.  This includes stdlib interfaces for common backends like caches and databases; addresses and pointers; containers, processes, services, and artifact-generating nodes.

For an example of Blueprint's compiler usage, see the [wiring spec of the Leaf example](examples/leaf/wiring/main.go)