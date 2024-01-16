# 2. Blueprint Overview

Blueprint is a compiler for microservice applications.  We recommend first familiarizing yourself with Blueprint's goals by reading our [SOSP 2023](https://blueprint-uservices.github.io/assets/pdf/anand2023blueprint.pdf) paper.

## Anatomy of an application

We describe the structure of a typical Blueprint application with reference to the [SockShop](../../examples/sockshop/) example.

Within the application directory, we can see the constituent components of a Blueprint application:

 * [workflow](../../examples/sockshop/workflow) is a Go module that implements the "business logic" of the application.  These are the services in the microservice application, such as the user service that stores user details.  Workflow services mostly make reference to each other (since services call each other), as well as backend services like databases, caches, etc.
 * [tests](../../examples/sockshop/tests) is a Go module that implements black-box tests of the workflow services.
 * [wiring](../../examples/sockshop/wiring) is a Go module that implements different configurations of the application, instantiating and combining instances in different configurations and topologies.  The wiring spec instantiates services that were declared in [workflow] and then applies different Blueprint plugins to those service instances.

## Anatomy of the compiler

* [blueprint](../../blueprint/) is the core compiler package 
* [plugins](../../plugins/) implements much of the functionality of Blueprint, in the form of compiler plugins
* [examples](../../examples/) contains a number of example Blueprint applications
* [runtime](../../runtime/) contains code that is used at runtime by compiled applications.  This includes the interfaces for backends used by workflow specs (in [runtime/core/backend](../../runtime/core/backend/)) and implementations that are automatically used by plugins (in [runtime/plugins](../../runtime/plugins))

## End-to-end workflow

 1. Define the workflow services of the application.  This code can make use of the interfaces defined in [runtime/core/backend](../../runtime/core/backend/).
 2. Create a wiring spec that instantiates services and applies plugins.  This code can make use of the compiler plugins defined in [plugins](../../plugins)
 3. Compile the application by invoking the compiler on the wiring spec
 4. Run the application by running the generated artifacts; this will vary depending on what you chose to generate (e.g. kubernetes manifests, dockerfiles, process scripts, etc.)
