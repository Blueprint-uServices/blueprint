module github.com/blueprint-uservices/blueprint/examples/petshop/wiring

go 1.25.0

require (
	github.com/blueprint-uservices/blueprint/blueprint v0.0.0-20240619221802-d064c5861c1e
	github.com/blueprint-uservices/blueprint/examples/petshop/tests v0.0.0
	github.com/blueprint-uservices/blueprint/examples/petshop/workflow v0.0.0
	github.com/blueprint-uservices/blueprint/plugins v0.0.0-20240619221802-d064c5861c1e
)

require (
	github.com/blueprint-uservices/blueprint/runtime v0.0.0-20240619221802-d064c5861c1e // indirect
	github.com/go-logr/logr v1.4.1 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/jmoiron/sqlx v1.4.0 // indirect
	github.com/mattn/go-sqlite3 v1.14.22 // indirect
	github.com/otiai10/copy v1.14.0 // indirect
	go.mongodb.org/mongo-driver v1.15.0 // indirect
	go.opentelemetry.io/otel v1.26.0 // indirect
	go.opentelemetry.io/otel/exporters/stdout/stdoutmetric v1.26.0 // indirect
	go.opentelemetry.io/otel/exporters/stdout/stdouttrace v1.26.0 // indirect
	go.opentelemetry.io/otel/metric v1.26.0 // indirect
	go.opentelemetry.io/otel/sdk v1.26.0 // indirect
	go.opentelemetry.io/otel/sdk/metric v1.26.0 // indirect
	go.opentelemetry.io/otel/trace v1.26.0 // indirect
	golang.org/x/exp v0.0.0-20240416160154-fe59bbe5cc7f // indirect
	golang.org/x/mod v0.38.0 // indirect
	golang.org/x/sync v0.22.0 // indirect
	golang.org/x/sys v0.47.0 // indirect
	golang.org/x/tools v0.48.0 // indirect
)

replace github.com/blueprint-uservices/blueprint/examples/petshop/workflow => ../workflow

replace github.com/blueprint-uservices/blueprint/examples/petshop/tests => ../tests
