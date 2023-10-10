package wiring

import (
	"testing"

	"gitlab.mpi-sws.org/cld/blueprint/plugins/clientpool"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/goproc"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/grpc"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/workflow"
)

func TestBasicClientPool(t *testing.T) {
	wiring := newWiringSpec("TestBasicClientPool")

	leaf := workflow.Define(wiring, "leaf", "TestLeafServiceImpl")
	nonleaf := workflow.Define(wiring, "nonleaf", "TestNonLeafService", leaf)

	clientpool.Create(wiring, leaf, 7)

	grpc.Deploy(wiring, leaf)
	grpc.Deploy(wiring, nonleaf)

	leafproc := goproc.CreateProcess(wiring, "leafproc", leaf)
	nonleafproc := goproc.CreateProcess(wiring, "nonleafproc", nonleaf)

	app := assertBuildSuccess(t, wiring, leafproc, nonleafproc)

	assertIR(t, app,
		`TestBasicClientPool = BlueprintApplication() {
			leaf.grpc.addr = GolangServerAddress()
			leaf.handler.visibility
			leafproc = GolangProcessNode(leaf.grpc.addr) {
			  leaf = TestLeafService()
			  leaf.grpc_server = GRPCServer(leaf, leaf.grpc.addr)
			}
			nonleaf.grpc.addr = GolangServerAddress()
			nonleaf.handler.visibility
			nonleafproc = GolangProcessNode(nonleaf.grpc.addr, leaf.grpc.addr) {
			  leaf.clientpool = ClientPool(leaf.grpc_client, 7) {
				leaf.grpc_client = GRPCClient(leaf.grpc.addr)
			  }
			  nonleaf = TestNonLeafService(leaf.clientpool)
			  nonleaf.grpc_server = GRPCServer(nonleaf, nonleaf.grpc.addr)
			}
		  }`)

}

func TestInvalidModifierOrder(t *testing.T) {
	wiring := newWiringSpec("TestBasicClientPool")

	leaf := workflow.Define(wiring, "leaf", "TestLeafServiceImpl")
	nonleaf := workflow.Define(wiring, "nonleaf", "TestNonLeafService", leaf)

	grpc.Deploy(wiring, leaf)
	clientpool.Create(wiring, leaf, 7)

	leafproc := goproc.CreateProcess(wiring, "leafproc", leaf)
	nonleafproc := goproc.CreateProcess(wiring, "nonleafproc", nonleaf)

	assertBuildFailure(t, wiring, leafproc, nonleafproc)

}
