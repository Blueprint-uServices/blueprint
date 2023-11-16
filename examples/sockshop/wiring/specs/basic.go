package specs

import (
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/wiring"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/simplenosqldb"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/workflow"
)

// Creates a basic sockshop wiring spec.
// Returns the names of the nodes to instantiate or an error
func BasicWiringSpec(spec wiring.WiringSpec) ([]string, error) {
	user_db := simplenosqldb.Define(spec, "user_db")
	user_service := workflow.Define(spec, "user_service", "UserService", user_db)

	return []string{user_service}, nil
}
