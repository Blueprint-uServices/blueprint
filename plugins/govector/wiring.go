// Package govector provides two plugins:
//
// (i) a plugin to wrap the client and server side of a service with a GoVector wrapper to maintain and propagate vector clocks for each process.
// A vector clock is a logical clock maintained by every process in a distributed system which can then be used to establish partial order between distributed operations. The plugin generates a log file for each process where the incremental vector timestamps are stored to track the propagation of requests.
//
// (ii) a plugin to install a GoVector logger for a given process. The log messages are appended with vector timestamps and placed in a log file in chronological order. Log files from all processes can be combined to visualize the full execution of a distributed system.
//
// GoVector is a vector clock logging library developed for educational purposes by researchers at UBC Systopia.
// More information on GoVector: https://github.com/DistributedClocks/GoVector
//
// Example Usage (for full instrumentation):
//
// import "github.com/blueprint-uservices/blueprint/plugins/govector"
//
// for _, service := range serviceNames {
//     govector.Instrument(spec, service) // Instrument the service to propagate vector clocks
// }
//
// for _, proc := range procNames {
//     logger := govector.DefineLogger(spec, proc) // Define a logger for the process
//	   goproc.SetLogger(spec, proc, logger) // Set the logger for the process
// }
package govector

import (
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/blueprint"
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/coreplugins/pointer"
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/ir"
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/wiring"
	"github.com/blueprint-uservices/blueprint/plugins/golang"
	"golang.org/x/exp/slog"
)

// Instruments the client and server side of a service with govector-instrumentation to initialize, maintain, and propagate vector clocks.
// The instrumentation generates logging events appended with vector clock timestamps.
// Ensures that the logs are sent to a GoVector logger defined with name `logger`
func Instrument(spec wiring.WiringSpec, serviceName string) {
	clientWrapper := serviceName + ".client.govec"
	serverWrapper := serviceName + ".server.govec"

	ptr := pointer.GetPointer(spec, serviceName)
	if ptr == nil {
		slog.Error("Unable to deploy " + serviceName + " using GoVector as it is not a pointer")
	}

	clientNext := ptr.AddSrcModifier(spec, clientWrapper)

	spec.Define(clientWrapper, &GovecClientWrapper{}, func(ns wiring.Namespace) (ir.IRNode, error) {
		var wrapped golang.Service
		if err := ns.Get(clientNext, &wrapped); err != nil {
			return nil, blueprint.Errorf("GoVector client %s expected %s to be a golang.Service, but encountered %s", clientWrapper, clientNext, err)
		}
		return newGovecClientWrapper(clientWrapper, wrapped)
	})

	serverNext := ptr.AddDstModifier(spec, serverWrapper)

	spec.Define(serverWrapper, &GovecServerWrapper{}, func(ns wiring.Namespace) (ir.IRNode, error) {
		var wrapped golang.Service
		if err := ns.Get(serverNext, &wrapped); err != nil {
			return nil, blueprint.Errorf("GoVector server %s expected %s to be a golang.Service, but encountered %s", serverWrapper, serverNext, wrapped)
		}
		return newGovecServerWrapper(serverWrapper, wrapped)
	})
}

// Defines a logger with name `loggerName`. The logger can then be used in process and service nodes.
func DefineLogger(spec wiring.WiringSpec, loggerName string) string {
	logger := loggerName + ".goveclogger"
	spec.Define(logger, &GoVecLoggerClient{}, func(ns wiring.Namespace) (ir.IRNode, error) {
		return newGoVecLoggerClient(logger)
	})
	return logger
}
