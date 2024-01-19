module github.com/blueprint-uservices/blueprint/examples/sockshop/workflow

go 1.20

require (
	github.com/blueprint-uservices/blueprint/runtime v0.0.0
	github.com/google/uuid v1.4.0
	github.com/pkg/errors v0.9.1
	github.com/stretchr/testify v1.8.4
	go.mongodb.org/mongo-driver v1.12.1
	golang.org/x/exp v0.0.0-20230728194245-b0cb94b80691
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	go.opentelemetry.io/otel v1.21.0 // indirect
	go.opentelemetry.io/otel/metric v1.21.0 // indirect
	go.opentelemetry.io/otel/trace v1.21.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace github.com/blueprint-uservices/blueprint/runtime => ../../../runtime
