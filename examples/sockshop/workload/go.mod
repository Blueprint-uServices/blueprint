module github.com/blueprint-uservices/blueprint/examples/sockshop/workload

go 1.21

toolchain go1.22.1

replace github.com/blueprint-uservices/blueprint/runtime => ../../../runtime

replace github.com/blueprint-uservices/blueprint/examples/sockshop/workflow => ../workflow

require github.com/blueprint-uservices/blueprint/examples/sockshop/workflow v0.0.0-20240405152959-f078915d2306

require (
	github.com/blueprint-uservices/blueprint/runtime v0.0.0 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	go.mongodb.org/mongo-driver v1.15.0 // indirect
	go.opentelemetry.io/otel v1.26.0 // indirect
	go.opentelemetry.io/otel/metric v1.26.0 // indirect
	go.opentelemetry.io/otel/trace v1.26.0 // indirect
	golang.org/x/exp v0.0.0-20240416160154-fe59bbe5cc7f // indirect
)
