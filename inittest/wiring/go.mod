module github.com/blueprint-uservices/blueprint/inittest/wiring

go 1.20

require github.com/blueprint-uservices/blueprint/blueprint v0.0.0

require github.com/blueprint-uservices/blueprint/plugins v0.0.0

require (
	github.com/blueprint-uservices/blueprint/examples/leaf/wiring v0.0.0-20240120085724-a66c24cd32b1
	github.com/blueprint-uservices/blueprint/examples/leaf/workflow v0.0.0-20240120085724-a66c24cd32b1
	github.com/blueprint-uservices/blueprint/inittest/workflow v0.0.0
	github.com/blueprint-uservices/blueprint/inittest/workflow2 v0.0.0
	golang.org/x/tools v0.15.0
)

require (
	github.com/blueprint-uservices/blueprint/runtime v0.0.0 // indirect
	github.com/otiai10/copy v1.11.0 // indirect
	go.mongodb.org/mongo-driver v1.12.1 // indirect
	go.opentelemetry.io/otel v1.21.0 // indirect
	go.opentelemetry.io/otel/metric v1.21.0 // indirect
	go.opentelemetry.io/otel/trace v1.21.0 // indirect
	golang.org/x/exp v0.0.0-20231110203233-9a3e6036ecaa // indirect
	golang.org/x/mod v0.14.0 // indirect
	golang.org/x/sys v0.14.0 // indirect
)

replace github.com/blueprint-uservices/blueprint/blueprint => ../../blueprint

replace github.com/blueprint-uservices/blueprint/plugins => ../../plugins

replace github.com/blueprint-uservices/blueprint/runtime => ../../runtime

replace github.com/blueprint-uservices/blueprint/inittest/workflow => ../workflow

replace github.com/blueprint-uservices/blueprint/inittest/workflow2 => ../workflow2
