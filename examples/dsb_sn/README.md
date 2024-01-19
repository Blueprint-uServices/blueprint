# DeathStarBench Social Network

This is a Blueprint re-implementation / translation of the [social-network application](https://github.com/delimitrou/DeathStarBench/tree/master/socialNetwork) from the DeathStarBench microservices benchmark.

The application provides a mostly-direct translation of the original code. In terms of the APIs and functionality, this implementation is intended to be as close to unmodified from the original as possible.

* [workflow](workflow) contains service implementations
* [tests](tests) has tests of the workflow
* [wiring](wiring) configures the application's topology and deployment and is used to compile the application

## Getting started

Prerequisites for this tutorial:
* [thrift compiler](https://thrift.apache.org/download) is installed
* docker is installed

## Running unit tests prior to compilation

Local unit tests are tests of an application's workflow logic within a single process. You can run local unit tests with:

```
cd tests
go test
```

Running the unit tests defined in ```usertimelineservice_test``` require a locally running ```mongodb``` instance as the query operators used by the UserTimelineService are not currently supported by the ```simplenosqldb``` implementation of the ```nosqldb``` instance.

Local unit tests only work on Linux or WSL.

## Compiling the application

To compile the application, we execute `wiring/main.go` and specify which wiring spec to compile. To view options and list wiring specs, run:

```
go run wiring/main.go -h
```

The following will compile the `docker` wiring spec to the directory `build`. This will fail if the pre-requisite thrift compiler is not installed.

```
go run wiring/main.go -w docker -o build
```

If you navigate to the `build` directory, you will now see a number of build artifact.
* `build/docker` contains docker images for the various containers of the application, as well as a `docker-compose.yml` file for starting and stopping all containers.
* `build/docker/*`  contain the individual docker images for services, including a Dockerfile and the golang source code.
* `build/gotests` contains the unit tests that we ran in the previous section.

## Running the application

To run the application, navigate to `build/docker` and run `docker compose up`.  You might see Docker complain about missing environment variables.  Edit the `.env` file in `build/docker` and put the following:

```
COMPOSEPOST_SERVICE_THRIFT_BIND_ADDR=9000
HOMETIMELINE_CACHE_BIND_ADDR=11211
HOMETIMELINE_SERVICE_THRIFT_BIND_ADDR=9001
MEDIA_SERVICE_THRIFT_BIND_ADDR=9002
POST_CACHE_BIND_ADDR=11212
POST_DB_BIND_ADDR=27017
POST_STORAGE_SERVICE_THRIFT_BIND_ADDR=9003
SOCIAL_CACHE_BIND_ADDR=11213
SOCIAL_DB_BIND_ADDR=27018
SOCIALGRAPH_SERVICE_THRIFT_BIND_ADDR=9004
TEXT_SERVICE_THRIFT_BIND_ADDR=9005
UNIQUEID_SERVICE_THRIFT_BIND_ADDR=9006
URLSHORTEN_SERVICE_THRIFT_BIND_ADDR=9007
URLSHORTEN_DB_BIND_ADDR=27019
USER_CACHE_BIND_ADDR=11214
USER_SERVICE_THRIFT_BIND_ADDR=9008
USER_DB_BIND_ADDR=27020
USERID_SERVICE_THRIFT_BIND_ADDR=9009
USERMENTION_SERVICE_THRIFT_BIND_ADDR=9010
USERTIMELINE_CACHE_BIND_ADDR=11215
USERTIMELINE_DB_BIND_ADDR=27021
USERTIMELINE_SERVICE_THRIFT_BIND_ADDR=9012
WRK2API_SERVICE_HTTP_BIND_ADDR=9011
```

## Running tests in compiled application

__NOTE: CURRENTLY TESTS NOT PASSING__

 After running the application, you can run the unit tests against it.

 ```
 cd build/gotests/tests
go test -wrk2api_service.http.dial_addr=localhost:9011 -userid_service.thrift.dial_addr=localhost:9009 -text_service.thrift.dial_addr=localhost:9005 -usertimeline_service.thrift.dial_addr=localhost:9012 -usermention_service.thrift.dial_addr=localhost:9010 -post_storage_service.thrift.dial_addr=localhost:9003 -media_service.thrift.dial_addr=localhost:9002 -urlshorten_service.thrift.dial_addr=localhost:9007 -composepost_service.thrift.dial_addr=localhost:9000 -hometimeline_service.thrift.dial_addr=localhost:9001 -uniqueid_service.thrift.dial_addr=localhost:9006 -user_service.thrift.dial_addr=localhost:9008 -socialgraph_service.thrift.dial_addr=localhost:9004
 ```