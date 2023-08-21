package grpc

import (
	"fmt"

	"gitlab.mpi-sws.org/cld/blueprint/pkg/blueprint"
	"gitlab.mpi-sws.org/cld/blueprint/pkg/plugins/golang"
)

/*
Simply defines a GRPC server wrapper node
*/
func DefineGRPCServerWrapper(wiring blueprint.WiringSpec, name string, wrappedHandler string) {
	wiring.SetProperty(name, "wrappedHandler", wrappedHandler)
	wiring.Define(name, &GolangGRPCServerNode{}, func(scope blueprint.Scope) (blueprint.IRNode, error) {
		wrappedHandlerProp, err := scope.GetProperty(name, "wrappedHandler")
		if err != nil {
			return nil, err
		}

		wrappedHandlerName, is_string := wrappedHandlerProp.(string)
		if !is_string {
			return nil, fmt.Errorf("grpc server wrapper %s expected a string \"wrappedHandler\" property, but got %v", name, wrappedHandlerName)
		}

		wrappedHandlerNode, err := scope.Get(wrappedHandlerName)
		if err != nil {
			return nil, err
		}

		wrappedService, is_service := wrappedHandlerNode.(golang.Service)
		if !is_service {
			return nil, fmt.Errorf("grpc server wrapper %s only supports wrapping golang services, but got %v", name, wrappedService)
		}

		server := newGolangGRPCServerNode(wrappedService.Name()+".grpc_server", wrappedService)
		return server, nil
	})
}
