package govector

import (
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/blueprint"
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/coreplugins/pointer"
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/ir"
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/wiring"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/golang"
	"golang.org/x/exp/slog"
)

// Instruments the service with an entry + exit point govector wrapper to generate govector logs.
func Instrument(spec wiring.WiringSpec, serviceName string) {
	clientWrapper := serviceName + ".client.govec"
	serverWrapper := serviceName + ".server.govec"

	ptr := pointer.GetPointer(spec, serviceName)
	if ptr == nil {
		slog.Error("Unable to deploy " + serviceName + " using GoVector as it is not a pointer")
	}

	clientNext := ptr.AddSrcModifier(spec, clientWrapper)

	spec.Define(clientWrapper, &govecClientWrapper{}, func(ns wiring.Namespace) (ir.IRNode, error) {
		var wrapped golang.Service
		if err := ns.Get(clientNext, &wrapped); err != nil {
			return nil, blueprint.Errorf("GoVector client %s expected %s to be a golang.Service, but encountered %s", clientWrapper, clientNext, err)
		}

		return newGovecClientWrapper(clientWrapper, wrapped)
	})

	serverNext := ptr.AddDstModifier(spec, serverWrapper)

	spec.Define(serverWrapper, &govecServerWrapper{}, func(ns wiring.Namespace) (ir.IRNode, error) {
		var wrapped golang.Service
		if err := ns.Get(serverNext, &wrapped); err != nil {
			return nil, blueprint.Errorf("GoVector server %s expected %s to be a golang.Service, but encountered %s", serverWrapper, serverNext, wrapped)
		}

		return newGovecServerWrappe(serverWrapper, wrapped)
	})
}
