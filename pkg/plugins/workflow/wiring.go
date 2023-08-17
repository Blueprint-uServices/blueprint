package workflow

import (
	"os"

	"gitlab.mpi-sws.org/cld/blueprint/pkg/blueprint"
	"gitlab.mpi-sws.org/cld/blueprint/pkg/core/pointer"
	"golang.org/x/exp/slog"
)

// Adds a service of type serviceType to the wiring spec, giving it the name specified.
// Services can have arguments which are other named nodes
func Define(wiring blueprint.WiringSpec, serviceName, serviceType string, args ...string) {
	// Eagerly look up the service in the workflow spec to make sure it exists
	details, err := findService(serviceType)
	if err != nil {
		slog.Error("Unable to resolve workflow spec services used by the wiring spec, exiting", "error", err)
		os.Exit(1)
	}

	// First, define the handler that will be called
	handler := serviceName + ".handler"
	wiring.Define(handler, &WorkflowService{}, func(scope blueprint.Scope) (blueprint.IRNode, error) {
		// Get all of the argument nodes; can error out if the arguments weren't actually defined
		// For arguments that are pointer types, this will only get the caller-side of the pointer
		var arg_nodes []blueprint.IRNode
		for _, arg_name := range args {
			node, err := scope.Get(arg_name)
			if err != nil {
				return nil, err
			}
			arg_nodes = append(arg_nodes, node)
		}

		// Instantiate and return the service
		service := newWorkflowService(serviceName, details, arg_nodes)
		return service, err
	})

	// Next, define the pointer
	ptr := serviceName + ".ptr"
	pointer.DefinePointer(wiring, ptr, handler, &blueprint.ApplicationNode{}, &WorkflowService{})

	// Lastly, use the service name as an alias to the ptr for anybody wanting to call it
	wiring.Alias(serviceName, ptr)

}

// TODO: implement a unique-clients option for Define, that opens a scope when getting the args, causing the clients to be unique
