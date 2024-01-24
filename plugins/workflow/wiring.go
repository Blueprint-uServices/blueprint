// Package workflow instantiates services defined in the application's workflow spec.
//
// A "Workflow Spec" or just "Workflow" defines the core business logic of an application.
// For example, in a social network application, the workflow defines how users can upload
// posts, view their timeline feed, follow other users, etc.
//
// Users of Blueprint are responsible for writing their application's workflow spec.  They
// then make use of Blueprint's compiler, and this workflow plugin, to compile the workflow
// spec into an application.
//
// See the [Workflow User Manual Page] for more details on writing workflows.
//
// # Wiring Spec Usage
//
// You can instantiate a service and give it a name, providing the service's interface or
// implementation type as the type parameter:
//
//	payment_service := workflow.Service[payment.PaymentService](spec, "payment_service")
//
// If a service has arguments (e.g. another service, a backend), then those arguments
// can be added to the call:
//
//	user_db := simple.NoSQLDB(spec, "user_db")
//	user_service := workflow.Service[user.UserService](spec, "user_service", user_db)
//
// If a service has configuration value arguments (e.g. a timeout) then string
// values can be provided for those arguments:
//
//	payment_service := workflow.Service[payment.PaymentService](spec, "payment_service", "500")
//
// The arguments provided to a service must match the arguments needed by the service's
// constructor in the workflow spec.  If they do not match, you will see a compilation error.
//
// # Generated Artifacts
//
// The workflow spec service implementation will be copied into the output directory.
// Where appropriate, the plugin generates constructor invocations, passing clients of
// other services/backends as constructor arguments.
//
// [Workflow User Manual Page]: https://github.com/Blueprint-uServices/blueprint/blob/main/docs/manual/workflow.md
package workflow

import (
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/blueprint"
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/coreplugins/pointer"
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/ir"
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/wiring"
	"github.com/blueprint-uservices/blueprint/plugins/golang/gocode"
)

var strtype = &gocode.BasicType{Name: "string"}

// [Service] is used by wiring specs to instantiate services from the workflow spec.
//
// Type parameter [ServiceType] is used to specify the type of the service.  It can be the name of an interface
// or an implementing struct.  [ServiceType] must be a valid workflow service: all interface methods must
// have [context.Context] arguments and [error] return values, and a constructor must be defined.
//
// `serviceName` is a unique name for the service instance.
//
// `serviceArgs` must correspond to the arguments of the service's constructor within the workflow spec.
// They can either be the names of other nodes that exist within this wiring spec, or string values for
// configuration arguments.  This determination is made by looking at the argument types of the constructor
// within the workflow spec (string arguments are treated as configuration; everything else is treated as
// a service instance).
//
// After calling [Service], serviceName is an application-level golang service.  Application-level modifiers
// can be applied to it, or it can be further deployed into e.g. a goproc, a linuxcontainer, etc.
func Service[ServiceType any](spec wiring.WiringSpec, serviceName string, serviceArgs ...string) string {
	// Define the service
	handlerName := serviceName + ".handler"
	spec.Define(handlerName, &workflowHandler{}, func(namespace wiring.Namespace) (ir.IRNode, error) {
		// Create the IR node for the handler
		handler := &workflowHandler{}
		if err := initWorkflowNode[ServiceType](&handler.workflowNode, serviceName); err != nil {
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
		if err := initWorkflowNode[ServiceType](&client.workflowNode, clientName); err != nil {
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
