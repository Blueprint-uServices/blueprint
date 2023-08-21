package workflow

import (
	"gitlab.mpi-sws.org/cld/blueprint/pkg/blueprint"
	"gitlab.mpi-sws.org/cld/blueprint/pkg/core/pointer"
)

/*
This adds a service to the application, using a definition that was provided in the workflow spec.

`serviceType` must refer to a named service that was defined in the workflow spec.  If the service
doesn't exist, then this will result in a build error.

`serviceArgs` can be zero or more other named nodes that are provided as arguments to the service.

This call creates several definitions within the wiring spec.  In particular, `serviceName` is
defined as a pointer to the actual service, and can thus be modified and
*/
func Define(wiring blueprint.WiringSpec, serviceName, serviceType string, serviceArgs ...string) {

	// Define the service
	handlerName := serviceName + ".handler"
	wiring.Define(handlerName, &WorkflowService{}, func(scope blueprint.Scope) (blueprint.IRNode, error) {
		// Look up the service details; errors out if the service doesn't exist
		details, err := findService(serviceType)
		if err != nil {
			return nil, err
		}

		// Get all of the argument nodes; can error out if the arguments weren't actually defined
		// For arguments that are pointer types, this will only get the caller-side of the pointer
		var arg_nodes []blueprint.IRNode
		for _, arg_name := range serviceArgs {
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

	// Mandate that this service with this name must be unique within the application (although, this can be changed by scopes)
	dstName := serviceName + ".dst"
	wiring.Alias(dstName, handlerName)
	pointer.RequireUniqueness(wiring, dstName, &blueprint.ApplicationNode{})

	// Lastly define the pointer
	pointer.CreatePointer(wiring, serviceName, &WorkflowService{}, dstName)
}

/*
TODOs:

-  can also implement a different version of Define that requests all clients specified in serviceArgs are unique.  This is achievable
   by just opening a scope when getting the args


*/
