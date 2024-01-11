// Package grpc is a plugin for deploying an application-level Golang service to a gRPC server.
//
// # Prerequisites
//
// To compile a Blueprint application that uses grpc, the build machine needs to have protocol buffers
// and the grpc compiler installed.  Installation instructions can be found on the [gRPC Quick Start].
//
// # Wiring Spec Usage
//
// To use the grpc plugin in your wiring spec, instantiate a workflow service and then
// invoke [Deploy]:
//
//	grpc.Deploy(spec, "my_service")
//
// Any application-level service modifiers (e.g. tracing) should be applied to the service *before*
// deploying it with gRPC.
//
// After deploying a service to gRPC, you will probably want to deploy the service in a process.
//
// # Example
//
// The SockShop [grpc wiring spec] uses the grpc plugin.
//
// # Configuration and Arguments
//
// The gRPC server requires an argument `bind_addr` to know which interface and port to bind to.
// This is a host:port string, typically looking something like "0.0.0.0:12345"
//
// The gRPC client requires an argument `dial_addr` to know which hostname and port to connect to.
// This is a host:port string, typically looking something like "192.168.1.2:12345" or "myhost:12345"
//
// Blueprint can automatically generate these addresses in some circumstances, but usually they have
// to be specified by you when running the application, such as when running processes or containers.
// For example, the process and container plugins will complain if arguments are missing.
//
// # Artifacts Generated
//
// The plugin will generate a server-side handler that creates and runs a gRPC server, with the
// service as the request handling logic.  The plugin will also generate a client implementation
// for other services to call the service.  The plugin also generates marshalling code for packing
// arguments into protobuf structs and vice versa.  This is implemented within
// the [grpccodegen] package.
//
// To use this plugin requires the protocol buffers and grpc compilers are installed
// on the machine that is compiling the Blueprint wiring spec.  Installation instructions
// can be found on the [gRPC Quick Start].
//
// [grpccodegen]: https://github.com/Blueprint-uServices/blueprint/tree/main/plugins/grpc/grpccodegen
// [grpc wiring spec]: https://github.com/Blueprint-uServices/blueprint/tree/main/examples/sockshop/wiring/specs/grpc.go
// [gRPC Quick Start]: https://grpc.io/docs/languages/go/quickstart/
package grpc

import (
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/blueprint"
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/coreplugins/address"
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/coreplugins/pointer"
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/ir"
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/wiring"
	"github.com/blueprint-uservices/blueprint/plugins/golang"
	"golang.org/x/exp/slog"
)

// [Deploy] can be used by wiring specs to deploy a workflow service using gRPC.
//
// serviceName should be the name of an applciation-level service; typically one that
// was defined using workflow.Define.
//
// Like many other modifiers, GRPC modifies the service at the golang level, by generating
// server-side handler code and a client-side library.  However, GRPC should be the last
// golang-level modifier applied to a service, because thereafter communication between
// the client and server is no longer at the golang level, but at the network level.
//
// Deploying a service with GRPC increases the visibility of the service within the application.
// By default, any other service running in any other container or namespace can now contact
// this service.
func Deploy(spec wiring.WiringSpec, serviceName string) {
	// The nodes that we are defining
	grpcClient := serviceName + ".grpc_client"
	grpcServer := serviceName + ".grpc_server"
	grpcAddr := serviceName + ".grpc.addr"

	// Get the pointer metadata
	ptr := pointer.GetPointer(spec, serviceName)
	if ptr == nil {
		slog.Error("Unable to deploy " + serviceName + " using GRPC as it is not a pointer")
		return
	}

	// Define the address that will be used by clients and the server
	address.Define[*golangServer](spec, grpcAddr, grpcServer)

	// Add the client-side modifier
	//
	// The client-side modifier creates a gRPC client and dials the server address.
	// It assumes the next src modifier node will be a golangServer address.
	clientNext := ptr.AddSrcModifier(spec, grpcClient)
	spec.Define(grpcClient, &golangClient{}, func(namespace wiring.Namespace) (ir.IRNode, error) {
		addr, err := address.Dial[*golangServer](namespace, clientNext)
		if err != nil {
			return nil, blueprint.Errorf("GRPC client %s expected %s to be an address, but encountered %s", grpcClient, clientNext, err)
		}
		return newGolangClient(grpcClient, addr)
	})

	// Add the server-side modifier, which is an address that PointsTo the grpcServer
	serverNext := ptr.AddAddrModifier(spec, grpcAddr)
	spec.Define(grpcServer, &golangServer{}, func(namespace wiring.Namespace) (ir.IRNode, error) {
		var wrapped golang.Service
		if err := namespace.Get(serverNext, &wrapped); err != nil {
			return nil, blueprint.Errorf("GRPC server %s expected %s to be a golang.Service, but encountered %s", grpcServer, serverNext, err)
		}

		server, err := newGolangServer(grpcServer, wrapped)
		if err != nil {
			return nil, err
		}

		err = address.Bind[*golangServer](namespace, grpcAddr, server, &server.Bind)
		server.Bind.PreferredPort = 12345
		return server, err
	})
}
