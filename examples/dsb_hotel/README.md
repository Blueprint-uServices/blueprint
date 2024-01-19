# DeathStarBench Hotel Reservation

This is a Blueprint re-implementation of the [hotel-reservation application](https://github.com/delimitrou/DeathStarBench/tree/master/hotelReservation).

For the most part, the application directly re-uses the original Hotel Reservation code. Some aspects of the application (such as HTTP URLs) were tweaked from the original implementation, but in terms of APIs and functionality, this implementation is intended to be as close to unmodified from the original as possible.

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

To compile the application, we execute `wiring/main.go` and specify which wiring spec to compile. To view options and list wiring specs, run:

```
go run wiring/main.go -h
```

The following will compile the `original` wiring spec to the directory `build`. This will fail if the pre-requisite grpc compiler is not installed.

```
go run wiring/main.go -w original -o build
```

If you navigate to the `build` directory, you will now see a number of build artifact.
* `build/docker` contains docker images for the various containers of the application, as well as a `docker-compose.yml` file for starting and stopping all containers.
* `build/docker/*`  contain the individual docker images for services, including a Dockerfile and the golang source code.
* `build/gotests` contains the unit tests that we ran in the previous section.

## Running the application

To run the application, navigate to `build/docker` and run `docker compose up`.  You might see Docker complain about missing environment variables.  Edit the `.env` file in `build/docker` and put the following:

```
FRONTEND_SERVICE_HTTP_BIND_ADDR=9000
GEO_DB_BIND_ADDR=27018
GEO_SERVICE_GRPC_BIND_ADDR=9001
JAEGER_BIND_ADDR=14268
PROFILE_CACHE_BIND_ADDR=11212
PROFILE_DB_BIND_ADDR=27019
PROFILE_SERVICE_GRPC_BIND_ADDR=9002
RATE_CACHE_BIND_ADDR=11213
RATE_DB_BIND_ADDR=27020
RATE_SERVICE_GRPC_BIND_ADDR=9003
RECOMD_DB_BIND_ADDR=27021
RECOMD_SERVICE_GRPC_BIND_ADDR=9004
RESERV_DB_BIND_ADDR=27022
RESERV_CACHE_BIND_ADDR=11214
RESERV_SERVICE_GRPC_BIND_ADDR=9005
SEARCH_SERVICE_GRPC_BIND_ADDR=9006
USER_DB_BIND_ADDR=27023
USER_SERVICE_GRPC_BIND_ADDR=9007
```

## Running tests in compiled application

 After running the application, you can run the unit tests against it.

 ```
 go test -v -profile_service.grpc.dial_addr=localhost:9002 -search_service.grpc.dial_addr=localhost:9006 -recomd_service.grpc.dial_addr=localhost:9004 -geo_service.grpc.dial_addr=localhost:9001 -user_service.grpc.dial_addr=localhost:9007 -rate_service.grpc.dial_addr=localhost:9003 -reserv_service.grpc.dial_addr=localhost:9005 -frontend_service.http.dial_addr=localhost:9000 -jaeger.dial_addr=localhost:14628
 ```