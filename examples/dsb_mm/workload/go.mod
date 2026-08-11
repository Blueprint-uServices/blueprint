module github.com/blueprint-uservices/blueprint/examples/dsb_mm/workload

go 1.22.0

require (
	github.com/blueprint-uservices/blueprint/examples/dsb_mm/workflow v0.0.0
	github.com/blueprint-uservices/blueprint/runtime v0.0.0-20260810164039-83c6feca6c74
)

require (
	github.com/dgrijalva/jwt-go v3.2.0+incompatible // indirect
	github.com/pkg/errors v0.9.1 // indirect
	go.mongodb.org/mongo-driver v1.15.0 // indirect
	go.opentelemetry.io/otel v1.26.0 // indirect
	go.opentelemetry.io/otel/metric v1.26.0 // indirect
	go.opentelemetry.io/otel/trace v1.26.0 // indirect
	golang.org/x/exp v0.0.0-20240416160154-fe59bbe5cc7f // indirect
	gonum.org/v1/gonum v0.15.1 // indirect
)

replace github.com/blueprint-uservices/blueprint/examples/dsb_mm/workflow => ../workflow
