module github.com/blueprint-uservices/blueprint/examples/leaf/workflow

go 1.20

require (
	github.com/blueprint-uservices/blueprint/runtime v0.0.0
	go.mongodb.org/mongo-driver v1.12.1
	go.opentelemetry.io/otel/metric v1.21.0
)

require (
	go.opentelemetry.io/otel v1.21.0 // indirect
	go.opentelemetry.io/otel/trace v1.21.0 // indirect
	golang.org/x/exp v0.0.0-20230728194245-b0cb94b80691 // indirect
)

replace github.com/blueprint-uservices/blueprint/runtime => ../../../runtime
