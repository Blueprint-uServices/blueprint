# Sockshop Microservices

This is a Blueprint re-implementation / translation of the [SockShop microservices demo](https://microservices-demo.github.io).

For the most part, the application directly re-uses the original SockShop code (for services that were written in Go) or does a mostly-direct translation of code (for services that were not written in Go).  Some aspects of the application (such as HTTP URLs) were tweaked from the original implementation, but in terms of APIs and functionality, this implementation is intended to be as close to unmodified from the original as possible.

* [workflow](workflow) contains service implementations
* [tests](tests) has tests of the workflow
* [wiring](wiring) configures the application's topology and deployment and is used to compile the application

## Getting started

Before running the example applications, make sure you have installed the recommended [prerequisites](requirements.md).

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

## Configure and run the application

To run the application, we will need to set a number of environment variables.  Blueprint will generate a `.local.env` file with some default values for these;
you can modify them if necessary.

Set the docker environment variables:

```
cd build
cp .local.env docker/.env
```

If this is your first time, you will need to build the containers:

```
cd docker
docker compose build
```

Run the application:

```
docker compose up
```

## Invoke the application

The SockShop application's [frontend API](workflow/frontend) is exposed by HTTP (when using the above configuration).
The port at which the service will be exposed is determined by the value of the variable `FRONTEND_HTTP_DIAL_ADDR` in the generated `.local.env` file.

For example, the value of the variable might be declared as `FRONTEND_HTTP_DIAL_ADDR=localhost:12356`.

We can invoke the `ListItems` API to list the socks in the application's catalogue:

```
curl http://localhost:12356/ListItems?pageSize=100\&pageNum=1
```

Alternatively in your web browser navigate to [localhost:12356/ListItems?pageSize=100&pageNum=1](http://localhost:12356/ListItems?pageSize=3&pageNum=1)

You should expect to see the following:

```
{
  "Ret0":[]
}
```

The result is empty as the catalogue databases are empty. Once they have been populated, the returned result will consist of more elements.

To populate the catalogue, use the following frontend API [localhost:12356/LoadCatalogue](http:/localhost:12356/LoadCatalogue)

After populating the databases, you should expect to see the following:

```
{
  "Ret0": [
    {
      "id": "3c59f984-80df-456c-8f56-6a2b57b7a342",
      "name": "Classic",
      "description": "Keep it simple.",
      "imageUrl": [
        "/catalogue/images/classic.jpg",
        "/catalogue/images/classic2.jpg"
      ],
      "price": 12,
      "quantity": 127,
      "tag": [
        "brown",
        "green"
      ]
    },
    {
      "id": "4999b899-e2c7-4e61-a799-68d0778aefe8",
      "name": "YouTube.sock",
      "description": "We were not paid to sell this sock. It's just a bit geeky.",
      "imageUrl": [
        "/catalogue/images/youtube_1.jpeg",
        "/catalogue/images/youtube_2.jpeg"
      ],
      "price": 10.99,
      "quantity": 801,
      "tag": [
        "formal",
        "geek"
      ]
    },
    {
      "id": "6f39c5c3-8ee8-47aa-ac7a-d5c14dcafb02",
      "name": "Nerd leg",
      "description": "For all those leg lovers out there. A perfect example of a swivel chair trained calf. Meticulously trained on a diet of sitting and Pina Coladas. Phwarr...",
      "imageUrl": [
        "/catalogue/images/bit_of_leg_1.jpeg",
        "/catalogue/images/bit_of_leg_2.jpeg"
      ],
      "price": 7.99,
      "quantity": 115,
      "tag": [
        "blue",
        "skin"
      ]
    }
  ]
}
```

## Viewing Traces

Navigate to Zipkin address defined in the `.local.env` file (eg: [http://localhost:12357](http://localhost:12357)) to view the Zipkin WebUI.

Click the "Query" button and you should see a trace with Root "frontend_proc: listitems start".  You can click "Show" to view the trace details.

## Testing the compiled application

You can run unit tests against the compiled application.  After starting the application, navigate to `build/gotests/tests` and run using `go test`.  You will need to set a number of address variables, which can either be passed as command line arguments, or sourced as environment variables.  We demonstrate the latter:

```
cd ./build
set -a
. ./.local.env
cd gotests/tests
go test .
```

The tests will also generate Zipkin traces which you can view in the Zipkin WebUI at [http://localhost:12357](http://localhost:12357).

## Running the workload generator

```
cd ./build
set -a
. ./.local.env
./wlgen/wlgen
```

## Next steps

### Custom Clients and Workloads

If you wish to write your own client application, you can import the `build/gotests/testclients` module and instantiate clients to any of the services as follows:

```
import "context"
import "blueprint/testclients/clients"
import "github.com/blueprint-uservices/blueprint/examples/sockshop/workflow/frontend"

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

Alternatively, client code to the SockShop frontend is also generated to `build/golang/golang/main.go`

### Changing the Application's Wiring Spec

The SockShop application comes with a number of out-of-the-box configurations; run `main.go` with the `-h` flag to list them, or view the documentation for the [wiring/specs](wiring/specs) package.

As a starting point for implementing your own custom wiring spec, we recommend duplicating and building off of the [basic.go](wiring/specs/basic.go) wiring spec.  After implementing your spec,
make sure that you add it to [wiring/main.go](wiring/main.go) so that it can be selected on the command line.