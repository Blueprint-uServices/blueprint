package workload

import (
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/blueprint"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/goproc"
)

// Creates a workload generator process that will invoke the specified service
func Generator(wiring blueprint.WiringSpec, service string) string {
	// We will define a process and, within it, a client
	workloadProcName := "client" + service
	workloadClientName := service + ".workloadgen.client"

	// Create a workload generator process that contains the client
	goproc.CreateProcess(wiring, workloadProcName, workloadClientName)

	// Define the workload generator client node
	wiring.Define(workloadClientName, &WorkloadgenClient{}, func(scope blueprint.Scope) (blueprint.IRNode, error) {
		client, err := scope.Get(service)
		if err != nil {
			return nil, err
		}

		return NewWorkloadGenerator(workloadClientName, client)
	})

	return workloadProcName
}
