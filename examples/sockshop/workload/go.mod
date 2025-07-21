module github.com/blueprint-uservices/blueprint/examples/sockshop/workload

go 1.22.0

toolchain go1.24.2

require github.com/blueprint-uservices/blueprint/runtime v0.0.0-20240619221802-d064c5861c1e // indirect

replace github.com/blueprint-uservices/blueprint/examples/sockshop/workflow => ../workflow

require github.com/blueprint-uservices/blueprint/examples/sockshop/workflow v0.0.0-20240405152959-f078915d2306

require (
	github.com/google/uuid v1.6.0 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	go.mongodb.org/mongo-driver v1.15.0 // indirect
	go.opentelemetry.io/otel v1.26.0 // indirect
	go.opentelemetry.io/otel/metric v1.26.0 // indirect
	go.opentelemetry.io/otel/trace v1.26.0 // indirect
	golang.org/x/exp v0.0.0-20241009180824-f66d83c29e7c // indirect
)
