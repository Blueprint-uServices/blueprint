module github.com/blueprint-uservices/blueprint/examples/online_boutique/tests

go 1.21

require (
	github.com/blueprint-uservices/blueprint/examples/online_boutique/workflow v0.0.0
	github.com/blueprint-uservices/blueprint/runtime v0.0.0-20240619221802-d064c5861c1e
	github.com/stretchr/testify v1.9.0
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	go.mongodb.org/mongo-driver v1.15.0 // indirect
	go.opentelemetry.io/otel v1.26.0 // indirect
	go.opentelemetry.io/otel/metric v1.26.0 // indirect
	go.opentelemetry.io/otel/trace v1.26.0 // indirect
	golang.org/x/exp v0.0.0-20240416160154-fe59bbe5cc7f // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace github.com/blueprint-uservices/blueprint/examples/online_boutique/workflow => ../workflow
