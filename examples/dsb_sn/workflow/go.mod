module gitlab.mpi-sws.org/cld/blueprint/examples/dsb_sn/workflow

go 1.20

replace gitlab.mpi-sws.org/cld/blueprint/runtime => ../../../runtime

require (
	github.com/dgrijalva/jwt-go v3.2.0+incompatible
	gitlab.mpi-sws.org/cld/blueprint/runtime v0.0.0-00010101000000-000000000000
	go.mongodb.org/mongo-driver v1.13.0
)

require (
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang/protobuf v1.5.3 // indirect
	github.com/tracingplane/tracingplane-go v0.0.0-20171025152126-8c4e6f79b148 // indirect
	go.opentelemetry.io/otel v1.20.0 // indirect
	go.opentelemetry.io/otel/trace v1.20.0 // indirect
	google.golang.org/protobuf v1.31.0 // indirect
)
