package specs

import (
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/wiring"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/simple"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/wiringcmd"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/workflow"
)

// Used by main.go
var Basic = wiringcmd.SpecOption{
	Name:        "basic",
	Description: "A basic single-process wiring spec with no modifiers",
	Build:       makeBasicSpec,
}

// Creates a basic sockshop wiring spec.
// Returns the names of the nodes to instantiate or an error
func makeBasicSpec(spec wiring.WiringSpec) ([]string, error) {
	user_db := simple.NoSQLDB(spec, "user_db")
	user_service := workflow.Define(spec, "user_service", "UserService", user_db)

	return []string{user_service}, nil
}
