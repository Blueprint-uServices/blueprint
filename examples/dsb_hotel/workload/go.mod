module github.com/blueprint-uservices/blueprint/examples/dsb_hotel/workload

go 1.21

replace github.com/blueprint-uservices/blueprint/examples/dsb_hotel/workflow => ../workflow

require github.com/blueprint-uservices/blueprint/examples/dsb_hotel/workflow v0.0.0

require github.com/blueprint-uservices/blueprint/runtime v0.0.0-20240619221802-d064c5861c1e // indirect

require (
	github.com/hailocab/go-geoindex v0.0.0-20160127134810-64631bfe9711 // indirect
	go.mongodb.org/mongo-driver v1.15.0 // indirect
	go.opentelemetry.io/otel v1.26.0 // indirect
	go.opentelemetry.io/otel/metric v1.26.0 // indirect
	go.opentelemetry.io/otel/trace v1.26.0 // indirect
	golang.org/x/exp v0.0.0-20240416160154-fe59bbe5cc7f // indirect
)
