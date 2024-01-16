module github.com/blueprint-uservices/blueprint/examples/sockshop/workload

go 1.20

replace github.com/blueprint-uservices/blueprint/runtime => ../../../runtime

replace github.com/blueprint-uservices/blueprint/examples/sockshop/workflow => ../workflow

require github.com/blueprint-uservices/blueprint/examples/sockshop/workflow v0.0.0-00010101000000-000000000000

require (
	github.com/blueprint-uservices/blueprint/runtime v0.0.0 // indirect
	github.com/golang/protobuf v1.5.3 // indirect
	github.com/google/uuid v1.4.0 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/tracingplane/tracingplane-go v0.0.0-20171025152126-8c4e6f79b148 // indirect
	go.mongodb.org/mongo-driver v1.12.1 // indirect
	go.opentelemetry.io/otel v1.21.0 // indirect
	go.opentelemetry.io/otel/metric v1.21.0 // indirect
	go.opentelemetry.io/otel/trace v1.21.0 // indirect
	golang.org/x/exp v0.0.0-20230728194245-b0cb94b80691 // indirect
	google.golang.org/protobuf v1.31.0 // indirect
)
