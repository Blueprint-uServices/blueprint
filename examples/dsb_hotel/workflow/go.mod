module github.com/blueprint-uservices/blueprint/examples/dsb_hotel/workflow

go 1.21

toolchain go1.22.1

require github.com/hailocab/go-geoindex v0.0.0-20160127134810-64631bfe9711

replace github.com/blueprint-uservices/blueprint/runtime => ../../../runtime

require (
	github.com/blueprint-uservices/blueprint/runtime v0.0.0-20240405152959-f078915d2306
	go.mongodb.org/mongo-driver v1.15.0
)

require (
	go.opentelemetry.io/otel v1.26.0 // indirect
	go.opentelemetry.io/otel/metric v1.26.0 // indirect
	go.opentelemetry.io/otel/trace v1.26.0 // indirect
	golang.org/x/exp v0.0.0-20240416160154-fe59bbe5cc7f // indirect
)
