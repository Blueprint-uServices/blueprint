package workflow

import (
	"path/filepath"
	"runtime"

	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/blueprint"
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/coreplugins/pointer"
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/ir"
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/wiring"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/golang/gocode"
	"golang.org/x/exp/slog"
)

var workflowSpecModulePaths = make(map[string]struct{})
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
		workflowPath := filepath.Clean(filepath.Join(dir, path))
		if _, exists := workflowSpecModulePaths[workflowPath]; !exists {
			slog.Info("Added workflow spec path " + workflowPath)
			workflowSpecModulePaths[workflowPath] = struct{}{}
		}
	}
	workflowSpec = nil
}

func Reset() {
	workflowSpecModulePaths = make(map[string]struct{})
}

// Static initialization of the workflow spec
func GetSpec() (*WorkflowSpec, error) {
	if workflowSpec != nil {
		return workflowSpec, nil
	}

	if len(workflowSpecModulePaths) == 0 {
		return nil, blueprint.Errorf("workflow spec src directories haven't been specified; use workflow.Init(srcPath) to add your workflow spec")
	}

	var modulePaths []string
	for modulePath := range workflowSpecModulePaths {
		modulePaths = append(modulePaths, modulePath)
	}
	spec, err := NewWorkflowSpec(modulePaths...)
	if err != nil {
		return nil, err
	}
	workflowSpec = spec
	return workflowSpec, nil
}

var strtype = &gocode.BasicType{Name: "string"}

/*
This adds a service to the application, using a definition that was provided in the workflow spec.

`serviceType` must refer to a named service that was defined in the workflow spec.  If the service
doesn't exist, then this will result in a build error.

`serviceArgs` can be zero or more other named nodes that are provided as arguments to the service.

This call creates several definitions within the wiring spec.  In particular, `serviceName` is
defined as a pointer to the actual service, and can thus be modified and
*/
func Service(spec wiring.WiringSpec, serviceName, serviceType string, serviceArgs ...string) string {
	// Define the service
	handlerName := serviceName + ".handler"
	spec.Define(handlerName, &workflowHandler{}, func(namespace wiring.Namespace) (ir.IRNode, error) {
		// Create the IR node for the handler
		handler := &workflowHandler{}
		if err := handler.Init(serviceName, serviceType); err != nil {
			return nil, err
		}

		// Check the serviceArgs provided match the constructorArgs from the actual code
		constructorArgs := handler.ServiceInfo.Constructor.Arguments[1:]
		if len(constructorArgs) != len(serviceArgs) {
			return nil, blueprint.Errorf("mismatched constructor arguments for %s, expect %v, got %v", serviceName, handler.ServiceInfo.Constructor, serviceArgs)
		}

		// Determine if any of the arguments are hard-coded values
		handler.Args = make([]ir.IRNode, len(constructorArgs))
		for i, arg := range constructorArgs {
			if arg.Type.Equals(strtype) && spec.GetDef(serviceArgs[i]) == nil {
				// A string argument with no node definition in the wiring spec is assumed to be a hard-coded value
				handler.Args[i] = &ir.IRValue{Value: serviceArgs[i]}
			} else {
				if err := namespace.Get(serviceArgs[i], &handler.Args[i]); err != nil {
					return nil, err
				}
			}
		}

		return handler, nil
	})

	// Mandate that this service with this name must be unique within the application (although, this can be changed by namespaces)
	dstName := serviceName + ".dst"
	spec.Alias(dstName, handlerName)
	pointer.RequireUniqueness(spec, dstName, &ir.ApplicationNode{})

	// Define the pointer
	ptr := pointer.CreatePointer(spec, serviceName, &workflowNode{}, dstName)

	// Add a "service.client" node for convenience
	clientName := serviceName + ".client"
	clientNext := ptr.AddSrcModifier(spec, clientName)
	spec.Define(clientName, &workflowClient{}, func(namespace wiring.Namespace) (ir.IRNode, error) {
		client := &workflowClient{}
		if err := client.Init(serviceName, serviceType); err != nil {
			return nil, err
		}
		return client, namespace.Get(clientNext, &client.Wrapped)
	})

	return serviceName
}

/*
TODOs:

-  can also implement a different version of Define that requests all clients specified in serviceArgs are unique.  This is achievable
   by just opening a namespace when getting the args


*/
