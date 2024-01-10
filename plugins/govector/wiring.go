// Package GoVector adds support for instrumenting applications with GoVector logger.
// GoVector logger maintains a vector clock for each process and implements the vector clock
// transmission between multiple processes.
// GoVector is a vector clock logging library developed for educational purposes.
// More information on GoVector: https://github.com/DistributedClocks/GoVector
package govector

import (
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/blueprint"
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/coreplugins/pointer"
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/ir"
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/wiring"
	"github.com/blueprint-uservices/blueprint/plugins/golang"
	"golang.org/x/exp/slog"
)

// Instruments the service with an entry + exit point govector wrapper to generate govector logs.
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
	// TODO: Require uniqueness for each logger
	return logger
}
