module gitlab.mpi-sws.org/cld/blueprint/examples/sockshop/workflow

go 1.20

require (
	github.com/google/uuid v1.4.0
	github.com/stretchr/testify v1.8.4
	gitlab.mpi-sws.org/cld/blueprint/runtime v0.0.0
	go.mongodb.org/mongo-driver v1.12.1
	golang.org/x/exp v0.0.0-20230728194245-b0cb94b80691
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang/protobuf v1.5.3 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/tracingplane/tracingplane-go v0.0.0-20171025152126-8c4e6f79b148 // indirect
	go.opentelemetry.io/otel v1.20.0 // indirect
	go.opentelemetry.io/otel/trace v1.20.0 // indirect
	google.golang.org/protobuf v1.31.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace gitlab.mpi-sws.org/cld/blueprint/runtime => ../../../runtime
