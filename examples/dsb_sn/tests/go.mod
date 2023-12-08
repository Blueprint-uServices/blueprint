module gitlab.mpi-sws.org/cld/blueprint/examples/dsb_sn/tests

go 1.20

require gitlab.mpi-sws.org/cld/blueprint/runtime v0.0.0

replace gitlab.mpi-sws.org/cld/blueprint/runtime => ../../../runtime

require (
	github.com/stretchr/testify v1.8.4
	gitlab.mpi-sws.org/cld/blueprint/examples/dsb_sn/workflow v0.0.0
	go.mongodb.org/mongo-driver v1.13.0
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/dgrijalva/jwt-go v3.2.0+incompatible // indirect
	github.com/golang/protobuf v1.5.3 // indirect
	github.com/golang/snappy v0.0.4 // indirect
	github.com/klauspost/compress v1.16.6 // indirect
	github.com/montanaflynn/stats v0.0.0-20171201202039-1bf9dbcd8cbe // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/tracingplane/tracingplane-go v0.0.0-20171025152126-8c4e6f79b148 // indirect
	github.com/xdg-go/pbkdf2 v1.0.0 // indirect
	github.com/xdg-go/scram v1.1.2 // indirect
	github.com/xdg-go/stringprep v1.0.4 // indirect
	github.com/youmark/pkcs8 v0.0.0-20181117223130-1be2e3e5546d // indirect
	go.opentelemetry.io/otel v1.20.0 // indirect
	go.opentelemetry.io/otel/trace v1.20.0 // indirect
	golang.org/x/crypto v0.11.0 // indirect
	golang.org/x/exp v0.0.0-20230728194245-b0cb94b80691 // indirect
	golang.org/x/sync v0.0.0-20220722155255-886fb9371eb4 // indirect
	golang.org/x/text v0.11.0 // indirect
	google.golang.org/protobuf v1.31.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace gitlab.mpi-sws.org/cld/blueprint/examples/dsb_sn/workflow => ../workflow
