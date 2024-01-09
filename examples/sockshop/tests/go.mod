module github.com/Blueprint-uServices/blueprint/examples/sockshop/tests

go 1.20

require github.com/Blueprint-uServices/blueprint/runtime v0.0.0

replace github.com/Blueprint-uServices/blueprint/runtime => ../../../runtime

require (
	github.com/pkg/errors v0.9.1
	github.com/stretchr/testify v1.8.4
	github.com/Blueprint-uServices/blueprint/examples/sockshop/workflow v0.0.0
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/golang/protobuf v1.5.3 // indirect
	github.com/google/uuid v1.4.0 // indirect
	github.com/jmoiron/sqlx v1.3.5 // indirect
	github.com/mattn/go-sqlite3 v1.14.17 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/tracingplane/tracingplane-go v0.0.0-20171025152126-8c4e6f79b148 // indirect
	go.mongodb.org/mongo-driver v1.12.1 // indirect
	go.opentelemetry.io/otel v1.20.0 // indirect
	go.opentelemetry.io/otel/trace v1.20.0 // indirect
	golang.org/x/exp v0.0.0-20230728194245-b0cb94b80691 // indirect
	google.golang.org/protobuf v1.31.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace github.com/Blueprint-uServices/blueprint/examples/sockshop/workflow => ../workflow
