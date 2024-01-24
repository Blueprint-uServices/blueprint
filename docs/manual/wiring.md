# Wiring Spec

A **Wiring Spec** or just **Wiring** instantiates services, connects them together, and specifies how they should be modified, configured, and deployed.  Wiring specs are separate from workflow specs, because the same workflow logic can be instantiated in many different ways.

Wiring specs are written as standalone go programs; invoking the program will compile the application.  To dive straight in to examples, see the [SockShop wiring](../../examples/sockshop/wiring).

## Project Layout

By convention we recommend placing wiring specs in a sibling module `wiring` alongside the workflow spec `workflow`.  The `wiring` subdirectory will be a golang module that contains your wiring spec implementation(s).  Your wiring module will likely want dependencies on the following modules:
 * `github.com/blueprint-uservices/blueprint/blueprint` - the core blueprint compiler
 * `github.com/blueprint-uservices/blueprint/plugins` - plugins that implement much of the wiring functionality

## Overview

A wiring spec does the following:

 1. It defines the instances that will exist when we run the program.
 2. It configures how those instances are deployed (e.g. within processes, containers, etc.)
 3. It connects instances together that communicate with each other
 4. It enables and configures features, such as tracing, that should be applied to instances.

When we execute a wiring spec, it *compiles* the application that was defined, generating all artifacts needed to actually run the application.  When we talk about instantiating services in the wiring spec, we mean defining service instances that will be compiled.

After executing a wiring spec, it will be up to you to actually run the application using the compiled artifacts.  Blueprint attempts to make this step as straightforward as possible, but it will typically still involve setting some environment variables and issuing a run command (e.g. a docker or kubernetes command).

## Writing a Wiring Spec, Simple Example

Imports

```
import "github.com/blueprint-uservices/blueprint/blueprint/pkg/wiring"
```

Initialize a new wiring spec

```
spec := wiring.NewWiringSpec("my_application")
```

Next, you would instantiate services using Blueprint [plugins](../../plugins), which we describe in more detail below.

After instantiating services we construct the IR

```
var nodesToBuild []string     // populated when instantiating services

applicationIR, err := spec.BuildIR(nodesToBuild...)  // Construct the IR
```

Constructing the IR might yield some errors if the wiring spec has errors, inconsistencies, or incompatible choices

We finally generate artifacts for the IR:

```
outputDir := "build"

err = applicationIR.GenerateArtifacts(outputDir)
```

If we then run the application, it will generate artifacts to `outputDir`.  (Note: the example above will not generate anything, because it doesn't instantiate any services).

## Wiring Spec Basics

### Create a Service

To define a service instance, we use Blueprint's [workflow](../../plugins/workflow) plugin.

Initialize a service that exists within the workflow spec.

```
echo_service := workflow.Service[workflow.EchoService](spec, "echo_service")
```

In the above example, we create an instance of "EchoService" and we call it "echo_service".  For more details on creating services, see the [workflow](../../plugins/workflow) plugin documentation.

To explicitly instantiate `echo_service` we will also need to add it to `nodesToBuild`:

```
nodesToBuild = append(nodesToBuild, "echo_service")
```

Note: it is unnecessary to explicitly add every service to `nodesToBuild`.  Blueprint can recursively discover service instances by following the dependencies between services, so typically the only service that needs to explicitly added to `nodesToBuild` is the front-end API service of the application (e.g. `nodesToBuild := []string{"frontend_service"}`).  This is usually sufficient to discover all other services in the application.

### Modify a Service

Elsewhere in the wiring spec, if we wish to refer to the echo service, then we will do so by (string) name (`"echo_service"`).  For example, the [grpc plugin](../../plugins/grpc) lets us deploy services using gRPC:

```
grpc.Deploy(spec, "echo_service")
```

In general, Blueprint plugins will expose a few methods that can be used within a wiring spec to define or modify service instances.  This functionality is entirely plugin-specific, so each plugin will define and document how it is supposed to be used.

### Plugins

We provide a non-exhaustive list of the more important Blueprint plugins on the [📝Wiring Spec Plugins](../plugins.md) page of the User Manual.  An exhaustive list of plugins can be found in the [plugins](../../plugins) module.

Examples of plugin usage can be found in the example applications, such as the [Leaf Application](../../examples/leaf/wiring/specs) and the [Sock Shop Application](../../examples/sockshop/wiring/specs).

## Cmdbuilder

It is usually useful to define multiple wiring specs for your application.  If this is the case, the [cmdbuilder](../../plugins/cmdbuilder) is a useful way of doing so.  All applications in the [examples](../../examples) directory make use of the cmdbuilder, and can be consulted for example usage.