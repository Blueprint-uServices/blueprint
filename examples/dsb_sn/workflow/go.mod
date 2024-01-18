module github.com/blueprint-uservices/blueprint/examples/dsb_sn/workflow

go 1.20

replace github.com/blueprint-uservices/blueprint/runtime => ../../../runtime

require (
	github.com/blueprint-uservices/blueprint/runtime v0.0.0-00010101000000-000000000000
	github.com/dgrijalva/jwt-go v3.2.0+incompatible
	go.mongodb.org/mongo-driver v1.13.0
)

require (
	go.opentelemetry.io/otel v1.21.0 // indirect
	go.opentelemetry.io/otel/metric v1.21.0 // indirect
	go.opentelemetry.io/otel/trace v1.21.0 // indirect
	golang.org/x/exp v0.0.0-20230728194245-b0cb94b80691 // indirect
)
