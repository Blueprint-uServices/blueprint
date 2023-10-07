module gitlab.mpi-sws.org/cld/blueprint/examples/sockshop/workflow

go 1.20

replace gitlab.mpi-sws.org/cld/blueprint/runtime => ../../../runtime

require (
	github.com/google/uuid v1.3.1
	gitlab.mpi-sws.org/cld/blueprint/runtime v0.0.0-00010101000000-000000000000
	golang.org/x/crypto v0.14.0
)

require go.mongodb.org/mongo-driver v1.12.1 // indirect
