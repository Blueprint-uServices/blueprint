package grpc

import (
	"gitlab.mpi-sws.org/cld/blueprint/pkg/blueprint"
)

func Apply(wiring *blueprint.WiringSpec, name string) {
	// wiring.Advertise(name, &blueprint.ApplicationNode{}, &blueprint.ApplicationNode{})
	// wiring.Wrap(name, &GolangGRPCServerNode{}, func(scope blueprint.Scope, wrapDef *blueprint.WiringDef) (blueprint.IRNode, error) {
	// 	// Build the wrapped node and keep it
	// 	wrapped := wrapDef.Build(scope)
	// 	scope.Put(wrapped.Name()+".grpc_handler", wrapped)

	// 	server := newGolangGRPCServerNode(wrapped.Name()+".grpc_server", wrapped)
	// 	return server, nil
	// })
}
