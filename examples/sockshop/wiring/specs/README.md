<!-- Code generated by gomarkdoc. DO NOT EDIT -->

# specs

```go
import "gitlab.mpi-sws.org/cld/blueprint/examples/sockshop/wiring/specs"
```

Package specs provides various different wiring specs for the SockShop application. These specs are used when running wiring/main.go.

## Index

- [Variables](<#variables>)


## Variables

<a name="Basic"></a>A simple wiring spec that compiles all services to a single process and therefore directly invoke each other. No RPC, containers, processes etc. are used.

```go
var Basic = wiringcmd.SpecOption{
    Name:        "basic",
    Description: "A basic single-process wiring spec with no modifiers",
    Build:       makeBasicSpec,
}
```

<a name="Docker"></a>A wiring spec that deploys each service into its own Docker container and using gRPC to communicate between services. The user, cart, shipping, and orders services using separate MongoDB instances to store their data. The catalogue service uses MySQL to store catalogue data. The shipping service and queue master service run within the same process \(TODO: separate processes\)

```go
var Docker = wiringcmd.SpecOption{
    Name:        "docker",
    Description: "Deploys each service in a separate container with gRPC, and uses mongodb as NoSQL database backends.",
    Build:       makeDockerSpec,
}
```

<a name="GRPC"></a>A wiring spec that deploys each service to a separate process, with services communicating over GRPC. The user, cart, shipping, and order services use simple in\-memory NoSQL databases to store their data. The catalogue service uses a simple in\-memory sqlite database to store its data. The shipping service and queue master service run within the same process \(TODO: separate processes\)

```go
var GRPC = wiringcmd.SpecOption{
    Name:        "grpc",
    Description: "Deploys each service in a separate process with gRPC.",
    Build:       makeGrpcSpec,
}
```

Generated by [gomarkdoc](<https://github.com/princjef/gomarkdoc>)