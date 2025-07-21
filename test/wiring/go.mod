module github.com/blueprint-uservices/blueprint/test/wiring

go 1.22.0

toolchain go1.24.2

require (
	github.com/blueprint-uservices/blueprint/test/workflow v0.0.0
	github.com/stretchr/testify v1.9.0
	golang.org/x/exp v0.0.0-20241009180824-f66d83c29e7c
)

require (
	github.com/blueprint-uservices/blueprint/runtime v0.0.0-20240619221802-d064c5861c1e // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/kr/text v0.2.0 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	go.mongodb.org/mongo-driver v1.15.0 // indirect
	go.opentelemetry.io/otel v1.26.0 // indirect
	go.opentelemetry.io/otel/metric v1.26.0 // indirect
	go.opentelemetry.io/otel/trace v1.26.0 // indirect
	gopkg.in/check.v1 v1.0.0-20201130134442-10cb98267c6c // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace github.com/blueprint-uservices/blueprint/test/workflow => ../workflow
