// Package workflow instantiates services defined in the application's workflow spec.
//
// # Wiring Spec Usage
//
// The plugin needs to know where to look for workflow spec services.  The plugin assumes
// paths relative to the calling file.
//
//	workflow.Init("../workflow", "../other_path")
//
// You can instantiate a service and give it a name by specifying either the name of the
// service's interface, implementation, or constructor.
//
//	payment_service := workflow.Service(spec, "payment_service", "PaymentService")
//
// If a service has arguments (e.g. another service, a backend), then those arguments
// can be added to the call:
//
//	user_db := simple.NoSQLDB(spec, "user_db")
//	user_service := workflow.Service(spec, "user_service", "UserService", user_db)
//
// If a service has configuration value arguments (e.g. a timeout) then string
// values can be provided for those arguments:
//
//	payment_service := workflow.Service(spec, "payment_service", "PaymentService", "500")
//
// The arguments provided to a service must match the arguments needed by the service's
// constructor in the workflow spec.  If they do not match, you will see a compilation error.
//
// # Generated Artifacts
//
// The workflow spec service implementation will be copied into the output directory.
// Where appropriate, the plugin generates constructor invocations, passing clients of
// other services/backends as constructor arguments.
package workflow

import (
	"path/filepath"
	"runtime"

	"github.com/blueprint-uservices/blueprint/blueprint/pkg/blueprint"
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/coreplugins/pointer"
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/ir"
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/wiring"
	"github.com/blueprint-uservices/blueprint/plugins/golang/gocode"
	"golang.org/x/exp/slog"
)

var workflowSpecModulePaths = make(map[string]struct{})
var workflowSpec *WorkflowSpec

// [Init] can be used by wiring specs to point the workflow plugin at the correct location of the
// application's workflow spec code.  It is required to be able to instantiate any services from
// that workflow spec.
//
// srcModulePaths should be paths to the *root* of a go module, ie. a directory containing a go.mod file.
//
// srcModulePaths are assumed to be *relative* to the calling file.  Typically this means
// calling something like:
//
//	workflow.Init("../workflow")
//
// Workflow specs can be included from more than one source module.  If also using the [gotests] plugin,
// then the location of the tests directory should also be provided, e.g.
//
//	workflow.Init("../workflow", "../tests")
//
// [Init] can be called more than once, which will concatenate all provided srcModulePaths.  [Reset]
// can be used to clear any previously provided module paths.
//
// [gotests]: https://github.com/Blueprint-uServices/blueprint/tree/main/plugins/gotests
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

// [Reset] can be used by wiring specs to clear any srcModulePaths given by previous calls to [Init]
func Reset() {
	workflowSpecModulePaths = make(map[string]struct{})
}

// [GetSpec] exists to enable other Blueprint plugins to access the parsed [*WorkflowSpec].
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
*/
// [Service] is used by wiring specs to instantiate services from the workflow spec.
//
// `serviceName`` is a unique name for the service instance.
//
// `serviceType` must refer to a named service that was defined in the workflow spec.  If the service
// doesn't exist, then this will result in a build error.  serviceType can be the name of an interface,
// an implementing struct, or a constructor.
//
// `serviceArgs` must correspond to the arguments of the service's constructor within the workflow spec.
// They can either be the names of other nodes that exist within this wiring spec, or string values for
// configuration arguments.  This determination is made by looking at the argument types of the constructor
// within the workflow spec (string arguments are treated as configuration; everything else is treated as
// a service instance).
//
// After calling [Service], serviceName is an application-level golang service.  Application-level modifiers
// can be applied to it, or it can be further deployed into e.g. a goproc, a linuxcontainer, etc.
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

	// Create a pointer to the handler
	ptr := pointer.CreatePointer[*workflowNode](spec, serviceName, handlerName)

	// Add a "service.client" node for convenience
	clientName := serviceName + ".client"
	clientNext := ptr.AddSrcModifier(spec, clientName)
	spec.Define(clientName, &workflowClient{}, func(namespace wiring.Namespace) (ir.IRNode, error) {
		client := &workflowClient{}
		if err := client.Init(clientName, serviceType); err != nil {
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
