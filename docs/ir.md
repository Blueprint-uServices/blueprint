# Blueprint IR

Blueprint's IR is a representation of a blueprint application that derives from a wiring spec and the workflow spec and plugins that it uses.  Blueprint's IR uses **typed Nodes** that plugins extend with specialized functionality.  The base type is `blueprint.IRNode` which is defined in [pkg/blueprint/ir.go](pkg/blueprint/ir.go).  All nodes must have a name and a string representation.

The root of a Blueprint application is a `blueprint.ApplicationNode`.  Building a Blueprint wiring spec will produce a blueprint application node.

## IR Node types

Blueprint's IR forms a type hierarchy building on the base `blueprint.IRNode`.  There are several further interfaces that extend the base `blueprint.IRNode`, such as (non-exhaustive):

* `service.ProcessNode` defined in [core/process/ir.go](pkg/core/process/ir.go) provides a generic representation of a process.
* `service.ServiceNode` defined in [core/service/ir.go](pkg/core/service/ir.go) provides a generic representation of a service with a synchronous call API.

Plugins can introduce further extensions of nodes as well as hierarchies of their own.  For example, the `golang` plugin introduces generic interfaces:

* `golang.Node` defined in [plugins/golang/ir.go](pkg/plugins/golang/ir.go) provides a generic representation of a golang object.
* `golang.Service` defined in [plugins/golang/ir.go](pkg/plugins/golang/ir.go) provides a generic representation of a golang service, extending the `ServiceNode` above.
* `golang.Process` defined in [plugins/golang/ir.go](pkg/plugins/golang/ir.go) provides a generic representation of a golang process, extending the `ProcessNode` above.

## Example

As an example, consider the [GRPC ir.go](pkg/plugins/grpc/ir.go), which defines two nodes, `GolangServer` and `GolangClient`.

* `GolangClient` is a GRPC client object that can be instantiated within a Go process, and it can invoke methods of a remote service.  For this, it needs the address of the remote service; this is a field `ServerAddr *pointer.Address` which is an `Address` IRNode.  GolangClient itself is a `golang.Service` node, because it provides methods that can be invoked by callers.  It is a `golang.Node` because it exists within a Go process.  It is also a `golang.ArtifactGenerator` and `golang.CodeGenerator`, which are node types with methods for generating, collecting, and packaging Go code.

* `GolangServer` is the corresponding GRPC server object that is also instantiated within a Go process.  A GolangServer also needs to know the address that is exposes, which is a field `Addr *pointer.Address` pointing to the same `Address` IRNode as the corresponding client would point to.  A `GolangServer` itself needs a handler that provides the actual methods that are exposed over RPC; this is a field `Wrapped golang.Service` which can point to any other node that implements the `golang.Service` interface.  GolangServer is a `golang.Node` because it exists within a Go process, and like the client, it is both an Artifact and a Code generator.  However, it is ***not*** a `golang.Service` node, because it does not have an interface that can be called directly by golang nodes -- callers would have to interact with the server's methods via the client.

## Code Generation etc.

IR node types provide the implementation of things like code generation that is used by compilation.  **The current IR implementation is partial and needs fleshing out**.  At the time of writing, golang code generation serves as the main example of how the IR hierarchy incorporates different concepts.





## Manual construction

In principle, Blueprint's wiring spec can be completely circumvented, and a caller could (laboriously) construct an IR manually by constructing nodes of an application.  In this regard, Blueprint's wiring is intended as a convenient API for programmatically constructing the IR nodes more easily.

## Current Status

