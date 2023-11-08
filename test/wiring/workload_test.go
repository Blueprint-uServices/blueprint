package wiring

import (
	"testing"

	"gitlab.mpi-sws.org/cld/blueprint/plugins/goproc"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/grpc"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/workflow"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/workload"
)

func TestBasicWorkloadGenerator(t *testing.T) {
	wiring := newWiringSpec("TestBasicWorkloadGenerator")

	leaf := workflow.Define(wiring, "leaf", "TestLeafServiceImpl")
	nonleaf := workflow.Define(wiring, "nonleaf", "TestNonLeafService", leaf)

	grpc.Deploy(wiring, leaf)
	grpc.Deploy(wiring, nonleaf)

	leafproc := goproc.CreateProcess(wiring, "leafproc", leaf)
	nonleafproc := goproc.CreateProcess(wiring, "nonleafproc", nonleaf)

	wlgen := workload.Generator(wiring, nonleaf)

	app := assertBuildSuccess(t, wiring, wlgen, leafproc, nonleafproc)

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
