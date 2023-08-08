package golang_workflow

import (
	"os"

	"gitlab.mpi-sws.org/cld/blueprint/pkg/blueprint"
	"golang.org/x/exp/slog"
)

// Adds a service of type serviceType to the wiring spec, giving it the name specified.
// Services can have arguments which are other named nodes
func Define(wiring *blueprint.WiringSpec, name, serviceType string, args ...string) {
	// Eagerly look up the service in the workflow spec to make sure it exists
	details, err := findService(serviceType)
	if err != nil {
		slog.Error("Unable to resolve workflow spec services used by the wiring spec, exiting", "error", err)
		os.Exit(1)
	}

	wiring.Define(name, &GolangWorkflowSpecServiceNode{}, func(scope blueprint.Scope) (any, error) {
		// Get all of the argument nodes; can error out if the arguments weren't actually defined
		var arg_nodes []blueprint.IRNode
		for _, arg_name := range args {
			node, err := scope.Get(arg_name)
			if err != nil {
				return nil, err
			}
			arg_nodes = append(arg_nodes, node)
		}

		// Instantiate and return the service
		service := newGolangWorkflowSpecServiceNode(name, details, arg_nodes)
		return service, err
	})
}

// TODO: implement a unique-clients option for Define, that opens a scope when getting the args, causing the clients to be unique
