package wiring

import (
	"testing"

	"gitlab.mpi-sws.org/cld/blueprint/plugins/goproc"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/grpc"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/workflow"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/workload"
)

func TestBasicWorkloadGenerator(t *testing.T) {
	spec := newWiringSpec("TestBasicWorkloadGenerator")

	leaf := workflow.Define(spec, "leaf", "TestLeafServiceImpl")
	nonleaf := workflow.Define(spec, "nonleaf", "TestNonLeafService", leaf)

	grpc.Deploy(spec, leaf)
	grpc.Deploy(spec, nonleaf)

	leafproc := goproc.CreateProcess(spec, "leafproc", leaf)
	nonleafproc := goproc.CreateProcess(spec, "nonleafproc", nonleaf)

	wlgen := workload.Generator(spec, nonleaf)

	app := assertBuildSuccess(t, spec, wlgen, leafproc, nonleafproc)

	assertIR(t, app,
		`TestBasicWorkloadGenerator = BlueprintApplication() {
      nonleaf.grpc.addr
      nonleaf.grpc.dial_addr = AddressConfig()
      clientnonleaf = GolangProcessNode(nonleaf.grpc.dial_addr) {
        nonleaf.grpc_client = GRPCClient(nonleaf.grpc.dial_addr)
        nonleaf.workloadgen.client = WorkloadGenerator(nonleaf.grpc_client)
      }
      leaf.grpc.addr
      leaf.grpc.bind_addr = AddressConfig()
      leaf.handler.visibility
      leafproc = GolangProcessNode(leaf.grpc.bind_addr) {
        leaf = TestLeafService()
        leaf.grpc_server = GRPCServer(leaf, leaf.grpc.bind_addr)
      }
      nonleaf.grpc.bind_addr = AddressConfig()
      nonleaf.handler.visibility
      leaf.grpc.dial_addr = AddressConfig()
      nonleafproc = GolangProcessNode(nonleaf.grpc.bind_addr, leaf.grpc.dial_addr) {
        leaf.grpc_client = GRPCClient(leaf.grpc.dial_addr)
        nonleaf = TestNonLeafService(leaf.grpc_client)
        nonleaf.grpc_server = GRPCServer(nonleaf, nonleaf.grpc.bind_addr)
      }
    }`)
}
