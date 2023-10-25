# Blueprint Thrift plugin

The Blueprint Thrift plugin can wrap any golang service node and expose it over RPC.

[wiring.go](wiring.go) contains functions that can be called by a wiring spec to expose a service over RPC.

## Requirements

The plugin makes use of the thrift compiler.  Thus:

* The thrift compiler ``thrift`` must be installed ([link](https://thrift.apache.org/download); Installation instructions are [here](https://thrift.apache.org/docs/install/debian.html)].



## IR Internals

The Blueprint Thrift plugin introduces three node types: a client, a server, and an address.

## Code generation

Code generation for Thrift comprises the following pieces:

- mapping the datatypes of the exposed interface into Thrift datatypes
- generating a Thrift proto file
- invoking the Thrift compiler to generate Thrift code from the thrift file
- generating a client implementation
- generating a server implementation

The above steps happen during module generation -- the Thrift plugin generates source files and adds them to the module being constructed.

In addition, code generation has the following:

- instantiation logic for client
- instantiation logic for server

The above steps happen during DI code generation -- e.g. as performed by the goproc plugin