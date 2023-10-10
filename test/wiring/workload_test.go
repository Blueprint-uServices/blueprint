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
            nonleaf.grpc.addr = GolangServerAddress()
            clientnonleaf = GolangProcessNode(nonleaf.grpc.addr) {
              nonleaf.grpc_client = GRPCClient(nonleaf.grpc.addr)
              nonleaf.workloadgen.client = WorkloadGenerator(nonleaf.grpc_client)
            }
            leaf.grpc.addr = GolangServerAddress()
            leaf.handler.visibility
            leafproc = GolangProcessNode(leaf.grpc.addr) {
              leaf = TestLeafService()
              leaf.grpc_server = GRPCServer(leaf, leaf.grpc.addr)
            }
            nonleaf.handler.visibility
            nonleafproc = GolangProcessNode(nonleaf.grpc.addr, leaf.grpc.addr) {
              leaf.grpc_client = GRPCClient(leaf.grpc.addr)
              nonleaf = TestNonLeafService(leaf.grpc_client)
              nonleaf.grpc_server = GRPCServer(nonleaf, nonleaf.grpc.addr)
            }
          }`)
}
