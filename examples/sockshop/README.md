# Sockshop Microservices

This is a Blueprint re-implementation / translation of the [SockShop microservices demo](https://microservices-demo.github.io).

For the most part, the application directly re-uses the original SockShop code (for services that were written in Go) or does a mostly-direct translation of code (for services that were not written in Go).  Some aspects of the application (such as HTTP URLs) were tweaked from the original implementation, but in terms of APIs and functionality, this implementation is intended to be as close to unmodified from the original as possible.

* [workflow](workflow) contains service implementations
* [tests](tests) has tests of the workflow
* [wiring](wiring) configures the application's topology and deployment and is used to compile the application

## Getting started

Prerequisites for this tutorial: 
 * gRPC and protocol buffers compiler are installed - https://grpc.io/docs/protoc-installation/
 * docker is installed

## Running unit tests prior to compilation

Local unit tests are tests of an application's workflow logic within a single process.  You can run local unit tests with:

```
cd tests
go test
```

Local unit tests only work on Linux or WSL.


## Compiling the application

To compile the application, we execute `wiring/main.go` and specify which wiring spec to compile.  To view options and list wiring specs, run:

```
go run wiring/main.go -h
```

The following will compile the `docker` wiring spec to the directory `build`.  This will fail if the pre-requisite gRPC and protocol buffers compilers aren't installed.

```
go run wiring/main.go -o build -w docker
```

If you navigate to the `build` directory, you will now see a number of build artifacts.
* `build/docker` contains docker images for the various containers of the application, as well as a `docker-compose.yml` file for starting and stopping all containers
* `build/docker/*` contain the individual docker images for services, including a Dockerfile and the golang source code
* `build/gotests` contain the unit tests that we ran in the "Running unit tests" section

## Running the application

To run the application, navigate to `build/docker` and run `docker-compose up`.  You might see Docker complain about missing environment variables.  Edit the `.env` file in `build/docker` and put the following:

```
USER_DB_BIND_ADDR=0.0.0.0:12345
PAYMENT_SERVICE_GRPC_BIND_ADDR=0.0.0.0:12346
USER_SERVICE_GRPC_BIND_ADDR=0.0.0.0:12347
CART_SERVICE_GRPC_BIND_ADDR=0.0.0.0:12348
ORDER_SERVICE_GRPC_BIND_ADDR=0.0.0.0:12349
ORDER_DB_BIND_ADDR=0.0.0.0:12350
SHIPPING_SERVICE_GRPC_BIND_ADDR=0.0.0.0:12351
SHIPPING_DB_BIND_ADDR=0.0.0.0:12352
CART_DB_BIND_ADDR=0.0.0.0:12353
CATALOGUE_DB_BIND_ADDR=0.0.0.0:12354
CATALOGUE_SERVICE_GRPC_BIND_ADDR=0.0.0.0:12355
FRONTEND_GRPC_BIND_ADDR=0.0.0.0:12356
ZIPKIN_BIND_ADDR=0.0.0.0:12357
```


## Running tests in compiled application

After running the application, you can run the unit tests against it.

```
cd build/gotests/tests
go test --payment_service.grpc.dial_addr=localhost:12346 --user_service.grpc.dial_addr=localhost:12347 --cart_service.grpc.dial_addr=localhost:12348 --order_service.grpc.dial_addr=localhost:12349 --shipping_service.grpc.dial_addr=localhost:12351 --catalogue_service.grpc.dial_addr=localhost:12355 --frontend.grpc.dial_addr=localhost:12356 --zipkin.dial_addr=localhost:12357
```

To see traces of the requests issued by the tests, navigate to [http://localhost:12357](http://localhost:12357)

## Next steps

If you wish to write your own client application, you can import the `build/gotests/testclients` module and instantiate clients to any of the services as follows:

```
import "context"
import "blueprint/testclients/clients"
import "gitlab.mpi-sws.org/cld/blueprint/examples/sockshop/workflow/frontend"

func getSockshopFrontendClient(ctx context.Context) (frontend.Frontend, error) {
    builder := clients.NewClientLibrary("myclient")

    // Optional: manually set addresses here; otherwise they are taken from cmd line args automatically when Build is called
    // builder.Set("user_service.grpc.dial_addr", "localhost:12347")

    clientlib := builder.Build(ctx)

    // Get the client to the FrontEnd (or, any other service defined in the wiring spec)
    var client frontend.FrontEnd
    err := clientlib.Get("frontend", &client)
    return client, err
}
```