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
 * (Optional, for critical path analysis) Python 3 and Docker for building the CRISP plugin

## Running unit tests prior to compilation

Local unit tests are tests of an application's workflow logic within a single process.  You can run local unit tests with:

```
cd tests
go test
```

Local unit tests only work on Linux or WSL.

## Compiling the application (including new wiring specs)

To compile the application, execute `wiring/main.go` and specify which wiring spec to compile. To view options and list wiring specs, run:

```
go run wiring/main.go -h
```

To compile a specific wiring spec (e.g., `original`, `chain`, `fanin`, `fanout`), run:

```
go run wiring/main.go -w <spec> -o build_<spec>
```

**Important:**
- If you add or modify a wiring spec (e.g., in `wiring/specs/chain.go`), you must recompile using the above command.
- If you add or update a plugin (e.g., CRISP), you must rebuild its Docker image (see below).

If you navigate to the `build_<spec>` directory, you will now see a number of build artifacts.
* `build_<spec>/docker` contains docker images for the various containers of the application, as well as a `docker-compose.yml` file for starting and stopping all containers.
* `build_<spec>/docker/*`  contain the individual docker images for services, including a Dockerfile and the golang source code.
* `build_<spec>/gotests` contains the unit tests that we ran in the previous section.

## Building the CRISP plugin (for critical path analysis)

If you are using the CRISP plugin for critical path analysis, you must build its Docker image before running the application:

```
cd plugins/crisp
docker build -t crisp:latest .
```

## Running the application

To run the application, navigate to `build_<spec>/docker` and run:

```
docker compose up
```

You might see Docker complain about missing environment variables.  Edit the `.env` file in `build_<spec>/docker` and put the following:

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

## Running different topologies (experiments)

You can experiment with different service topologies by selecting the appropriate wiring spec:
- `original`: Full DeathStarBench topology
- `chain`: Frontend → Search → Geo
- `fanin`: Search → Geo, Profile → Geo
- `fanout` (star): Frontend calls all backends directly

Compile and run as described above, substituting `<spec>` with the desired topology.

**Note:**
- You can specify the wiring spec to use by passing the `-w <spec>` flag to `wiring/main.go`.
- If you are editing the wiring code directly, ensure the correct spec name is set in the wiring file or in your build/run command.
- Alternatively, you can change the default experiment by editing the spec in `wiring/main.go` (e.g., change `specs.Fanout` to `specs.Chain`, `specs.Original`, or `specs.Fanin`).

## Analyzing traces and critical path

- Access Jaeger UI at [http://localhost:16686](http://localhost:16686) to view traces.
- If using CRISP, send a request to the CRISP API to analyze a trace:
  ```
  curl http://localhost:8000/analyze/<trace_id>
  ```
- The output will show the critical path and its total latency for the selected trace.