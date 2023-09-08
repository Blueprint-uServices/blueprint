# Blueprint GRPC plugin

The Blueprint GRPC plugin can wrap any golang service node and expose it over RPC.

[wiring.go](wiring.go) contains functions that can be called by a wiring spec to expose a service over RPC.

## Requirements

The plugin makes use of the protocol buffers compiler `protoc` and its Golang extensions.  Thus:

* The protocol buffers compiler `protoc` must be installed ([link](https://github.com/protocolbuffers/protobuf/releases))
* The golang and grpc extensions must also be installed ([link](https://grpc.io/docs/languages/go/quickstart/))



## IR Internals

The Blueprint GRPC plugin introduces three node types: a client, a server, and an address.

## Code generation

Code generation for GRPC comprises the following pieces:

- mapping the datatypes of the exposed interface into GRPC datatypes
- generating a GRPC proto file
- invoking the GRPC compiler to generate GRPC code from the proto file
- generating a client implementation
- generating a server implementation

The above steps happen during module generation -- the GPRC plugin generates source files and adds them to the module being constructed.

In addition, code generation has the following:

- instantiation logic for client
- instantiation logic for server

The above steps happen during DI code generation -- e.g. as performed by the goproc plugin