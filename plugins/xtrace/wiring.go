// Package xtrace provides a Blueprint plugin for instrumenting services and collecting X-Trace traces. It provides APIs to be used by the wiring spec to do the following:
//
//  1. to wrap the service with an XTrace wrapper to generate XTrace compatible traces/logs by correctly propagating baggage across service boundaries. Automatically generates and include an xtrace-server instance in a Blueprint application.
//  2. to define an xtrace-based logger for a process. Log events are added as reports to the currently active xtrace task, if one exists. If no such task exists, then no events are logged.
//
// Once the application is instrumented with these plugins, traces will be generated and collected by the xtrace-server.
//
// # Wiring Usage
//
// To instrument the client and server side of a service with x-trace instrumentation:
//
//	xtrace.Instrument(spec, "my_service")
//
// Calling [Instrument] will automatically create an xtrace server container instance (called "xtrace_server") that will be responsible for receiving trace data from all instrumented processes.
//
// To redirect a process's logging statements to X-Trace:
//
//	xtrace.Logger(spec, "my_process")
//
// In order to generate complete end-to-end traces of the application, all services of the application need to be instrumented with XTrace.
// If the plugin is only applied to a subset of services, the application will run, but the traces it produces won't be end-to-end and won't be useful.
//
// # Artifacts Generated
//
//  1. The package generates a built-in xtrace container that provides the server-side implementation and a go-client for connecting to the server.
//  2. Generates client and server side wrappers for instrumented servers that contain xtrace instrumentation (baggage propagation, xtrace event generation)
//
// # Full Wiring Example
//
// The following example creates an x-trace server, instruments all services, and uses the xtrace logger in all processes.
//
//	func applyXtraceOptions() {
//		for _, service := range serviceNames {
//			xtrace.Instrument(spec, service)
//		}
//		for _, proc := range processNames {
//			xtrace.Logger(spec, proc)
//		}
//	}
//
// See the [xtrace_logger] wiring spec for the Leaf application for a further example
//
// # Accessing Traces
//
// The traces are generated and sent to the xtrace-server. To access traces, navigate to xtrace-server:4080 to view all generated traces. (Assuming that the xtrace server container is running at address xtrace-server with its internal port 4080 bound to host port 4080).
//
// [xtrace_logger]: https://github.com/Blueprint-uServices/blueprint/tree/main/examples/leaf/wiring/specs/custom_logger.go
package xtrace

import (
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/blueprint"
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/coreplugins/address"
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/coreplugins/pointer"
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/ir"
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/wiring"
	"github.com/blueprint-uservices/blueprint/plugins/golang"
	"github.com/blueprint-uservices/blueprint/plugins/goproc"
	"golang.org/x/exp/slog"
)

var default_xtrace_server_name = "xtrace_server"

// [Instrument] can be invoked by a wiring spec to instrument the client and server side of the service with name `serviceName` to add xtrace context propagation.
// Also generates xtrace events in the instrumented code.
//
// Defines and adds an xtrace server if not already defined.
//
// Usage:
//
//	xtrace.Instrument(spec, "serviceA")
func Instrument(spec wiring.WiringSpec, serviceName string) {
	xtraceServer := container(spec, default_xtrace_server_name)
	clientWrapper := serviceName + ".client.xtrace"
	serverWrapper := serviceName + ".server.xtrace"

	ptr := pointer.GetPointer(spec, serviceName)
	if ptr == nil {
		slog.Error("Unable to deploy " + serviceName + " using XTrace as it is not a pointer")
	}

	clientNext := ptr.AddSrcModifier(spec, clientWrapper)
	spec.Define(clientWrapper, &XtraceClientWrapper{}, func(ns wiring.Namespace) (ir.IRNode, error) {
		var wrapped golang.Service
		if err := ns.Get(clientNext, &wrapped); err != nil {
			return nil, blueprint.Errorf("XTrace client %s expected %s to be a golang.Service, but encountered %s", clientWrapper, clientNext, err)
		}

		var xtraceClient *XTraceClient
		err := ns.Get(xtraceServer, &xtraceClient)
		if err != nil {
			return nil, err
		}

		return newXtraceClientWrapper(clientWrapper, wrapped, xtraceClient)
	})

	serverNext := ptr.AddDstModifier(spec, serverWrapper)
	spec.Define(serverWrapper, &XtraceServerWrapper{}, func(ns wiring.Namespace) (ir.IRNode, error) {
		var wrapped golang.Service
		if err := ns.Get(serverNext, &wrapped); err != nil {
			return nil, blueprint.Errorf("XTrace server %s expected %s to be a golang.Service, but encountered %s", serverWrapper, serverNext, wrapped)
		}

		var xtraceClient *XTraceClient
		err := ns.Get(xtraceServer, &xtraceClient)
		if err != nil {
			return nil, err
		}

		return newXtraceServerWrapper(serverWrapper, wrapped, xtraceClient)
	})
}

// Adds an xtrace docker container that uses the latest xtrace image to the application
// along with the default client needed by the generated application to communicate with the server.
//
// The generated container has the name `serviceName`.
// Usage:
//
//  xtrace.container(spec, "xtrace_server")
func container(spec wiring.WiringSpec, serverName string) string {
	// The nodes that we are defining
	xtraceAddr := serverName + ".addr"
	xtraceClient := serverName + ".client"
	xtraceCtr := serverName + ".ctr"

	// Define the X-Trace server container
	spec.Define(xtraceCtr, &XTraceServerContainer{}, func(ns wiring.Namespace) (ir.IRNode, error) {
		xtrace, err := newXTraceServerContainer(xtraceCtr)
		if err != nil {
			return nil, err
		}

		err = address.Bind[*XTraceServerContainer](ns, xtraceAddr, xtrace, &xtrace.BindAddr)
		return xtrace, err
	})

	// Create a pointer to the server
	ptr := pointer.CreatePointer[*XTraceClient](spec, serverName, xtraceCtr)

	// Define the address that points to the X-Trace collector
	address.Define[*XTraceServerContainer](spec, xtraceAddr, xtraceCtr)

	// Add the address to the pointer
	ptr.AddAddrModifier(spec, xtraceAddr)

	// Define the X-Trace client and add it to the client side of the pointer
	clientNext := ptr.AddSrcModifier(spec, xtraceClient)
	spec.Define(xtraceClient, &XTraceClient{}, func(ns wiring.Namespace) (ir.IRNode, error) {
		addr, err := address.Dial[*XTraceServerContainer](ns, clientNext)
		if err != nil {
			return nil, err
		}

		return newXTraceClient(xtraceClient, addr.Dial)
	})

	// Return the pointer; anybody who wants to access the X-Trace server should do so through the pointer
	return serverName
}

// Defines and installs an xtrace-based logger to the process with name `processName`. Replaces the existing logger installed for the process.
// Instantiates the logger, registers the logger as the default logger for the desired process, and returns the instantiated logger's name.
// Logged events at runtime are added as reports to the currently active XTrace task, if available. If no such task exists, then no log events are generated.
// Log messages are not printed to stdout as they are captured by the xtrace library and attached to the trace.
//
// Defines and adds an xtrace server if not already defined.
//
// All services in the process must also be instrumented with `Instrument` to get log statements associated with a given xtrace task.
// Process with name `processName` must already be defined.
//
// # Wiring Spec Usage:
//
//   xtrace.Logger(spec, "my_process") // Define an xtrace-logger for the process `my_process`
func Logger(spec wiring.WiringSpec, processName string) string {
	logger := "xtrace_logger"
	xtrace_server := container(spec, default_xtrace_server_name)
	xtrace_addr := xtrace_server + ".addr"
	spec.Define(logger, &XTraceLogger{}, func(ns wiring.Namespace) (ir.IRNode, error) {
		addr, err := address.Dial[*XTraceServerContainer](ns, xtrace_addr)
		if err != nil {
			return nil, err
		}

		return newXTraceLogger(logger, addr.Dial)
	})
	goproc.SetLogger(spec, processName, logger)
	return logger
}
