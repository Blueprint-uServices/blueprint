package workflow

import (
	"fmt"
	"path/filepath"
	"runtime"

	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/blueprint"
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/core/pointer"
)

var workflowSpecModulePaths []string
var workflowSpec *WorkflowSpec

/*
The Golang workflow plugin must be initialized in the wiring spec with the location of the workflow
spec modules.

Workflow specs can be included from more than one source module.

The provided paths should be to the root of a go module (containing a go.mod file).  The arguments
are assumed to be **relative** to the calling file.

This can be called more than once, which will concatenate all provided srcModulePaths
*/
func Init(srcModulePaths ...string) {
	_, callingFile, _, _ := runtime.Caller(1)
	dir, _ := filepath.Split(callingFile)
	for _, path := range srcModulePaths {
		workflowPath := filepath.Join(dir, path)
		workflowSpecModulePaths = append(workflowSpecModulePaths, workflowPath)
	}
	workflowSpec = nil
}

// Static initialization of the workflow spec
func getSpec() (*WorkflowSpec, error) {
	if workflowSpec != nil {
		return workflowSpec, nil
	}

	if len(workflowSpecModulePaths) == 0 {
		return nil, fmt.Errorf("workflow spec src directories haven't been specified; use workflow.Init(srcPath) to add your workflow spec")
	}

	spec, err := NewWorkflowSpec(workflowSpecModulePaths...)
	if err != nil {
		return nil, err
	}
	workflowSpec = spec
	return workflowSpec, nil
}

/*
This adds a service to the application, using a definition that was provided in the workflow spec.

`serviceType` must refer to a named service that was defined in the workflow spec.  If the service
doesn't exist, then this will result in a build error.

`serviceArgs` can be zero or more other named nodes that are provided as arguments to the service.

This call creates several definitions within the wiring spec.  In particular, `serviceName` is
defined as a pointer to the actual service, and can thus be modified and
*/
func Define(wiring blueprint.WiringSpec, serviceName, serviceType string, serviceArgs ...string) string {
	// Define the service
	handlerName := serviceName + ".handler"
	wiring.Define(handlerName, &WorkflowService{}, func(scope blueprint.Scope) (blueprint.IRNode, error) {
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
		return newWorkflowService(serviceName, serviceType, arg_nodes)
	})

	// Mandate that this service with this name must be unique within the application (although, this can be changed by scopes)
	dstName := serviceName + ".dst"
	wiring.Alias(dstName, handlerName)
	pointer.RequireUniqueness(wiring, dstName, &blueprint.ApplicationNode{})

	// Lastly define the pointer
	pointer.CreatePointer(wiring, serviceName, &WorkflowService{}, dstName)

	return serviceName
}

/*
TODOs:

-  can also implement a different version of Define that requests all clients specified in serviceArgs are unique.  This is achievable
   by just opening a scope when getting the args


*/
