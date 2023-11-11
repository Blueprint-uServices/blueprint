package workload

import (
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/ir"
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/wiring"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/goproc"
)

// Creates a workload generator process that will invoke the specified service
func Generator(spec wiring.WiringSpec, service string) string {
	// We will define a process and, within it, a client
	workloadProcName := "client" + service
	workloadClientName := service + ".workloadgen.client"

	// Create a workload generator process that contains the client
	goproc.CreateProcess(spec, workloadProcName, workloadClientName)

	// Define the workload generator client node
	spec.Define(workloadClientName, &WorkloadgenClient{}, func(namespace wiring.Namespace) (ir.IRNode, error) {
		var client ir.IRNode
		if err := namespace.Get(service, &client); err != nil {
			return nil, err
		}

		return NewWorkloadGenerator(workloadClientName, client)
	})

	return workloadProcName
}
