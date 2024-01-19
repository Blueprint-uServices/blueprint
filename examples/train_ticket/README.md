# Train Ticket

This is a Blueprint re-implementation / translation of the [Train Ticket application](https://github.com/FudanSELab/train-ticket/tree/master).

The application provides a mostly-direct translation of the original code. In terms of the APIs and functionality, this implementation is intended to be as close to unmodified from the original as possible.

* [workflow](workflow) contains service implementations
* [tests](tests) has tests of the workflow
* [wiring](wiring) configures the application's topology and deployment and is used to compile the application

## Running unit tests prior to compilation

Local unit tests are tests of an application's workflow logic within a single process. You can run local unit tests with:

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

The following will compile the `docker` wiring spec to the directory `build`. 

```
go run wiring/main.go -o build -w docker
```

If you navigate to the `build` directory, you will now see a number of build artifacts.
* `build/docker` contains docker images for the various containers of the application, as well as a `docker-compose.yml` file for starting and stopping all containers
* `build/docker/*` contain the individual docker images for services, including a Dockerfile and the golang source code
* `build/gotests` contain the unit tests that we ran in the "Running unit tests" section

## Running the application

To run the application, navigate to `build/docker` and run `docker compose up`.  You might see Docker complain about missing environment variables.  Edit the `.env` file in `build/docker` and put the following:

```
ASSURANCE_SERVICE_HTTP_BIND_ADDR=9003
ASSURANCE_DB_BIND_ADDR=9004
CONFIG_SERVICE_HTTP_BIND_ADDR=9005
CONFIG_DB_BIND_ADDR=9006
CONSIGNPRICE_SERVICE_HTTP_BIND_ADDR=9007
CONSIGNPRICE_DB_BIND_ADDR=9008
CONTACTS_SERVICE_HTTP_BIND_ADDR=9009
CONTACTS_DB_BIND_ADDR=9010
DELIVERY_DB_BIND_ADDR=9011
DELIVERY_Q_BIND_ADDR=9012
MONEY_DB_BIND_ADDR=9013
NEWS_SERVICE_HTTP_BIND_ADDR=9014
PAYMENTS_SERVICE_HTTP_BIND_ADDR=9015
PAYMENTS_DB_BIND_ADDR=9016
PRICE_SERVICE_HTTP_BIND_ADDR=9017
PRICE_DB_BIND_ADDR=9018
ROUTE_SERVICE_HTTP_BIND_ADDR=9019
ROUTE_DB_BIND_ADDR=9020
STATION_SERVICE_HTTP_BIND_ADDR=9021
STATION_DB_BIND_ADDR=9022
STATIONFOOD_SERVICE_HTTP_BIND_ADDR=9023
STATIONFOOD_DB_BIND_ADDR=9024
TRAIN_SERVICE_HTTP_BIND_ADDR=9025
TRAIN_DB_BIND_ADDR=9026
TRAINFOOD_SERVICE_HTTP_BIND_ADDR=9027
TRAINFOOD_DB_BIND_ADDR=9028
USER_SERVICE_HTTP_BIND_ADDR=9029
USER_DB_BIND_ADDR=9030
```

## Running tests in compiled application

After running the application, you can run the unit tests against it.

```
cd build/gotests/tests
go test -v -user_service.http.dial_addr=localhost:9029 -route_service.http.dial_addr=localhost:9019 -assurance_service.http.dial_addr=localhost:9003 -station_service.http.dial_addr=localhost:9021 -consignprice_service.http.dial_addr=localhost:9007 -stationfood_service.http.dial_addr=localhost:9023 -contacts_service.http.dial_addr=localhost:9009 -train_service.http.dial_addr=localhost:9025 -trainfood_service.http.dial_addr=localhost:9027 -payments_service.http.dial_addr=localhost:9015 -config_service.http.dial_addr=localhost:9005 -price_service.http.dial_addr=localhost:9017
```