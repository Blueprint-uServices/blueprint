package wiring

import (
	"testing"

	"gitlab.mpi-sws.org/cld/blueprint/plugins/clientpool"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/goproc"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/grpc"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/retries"
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

func TestBasicClientPoolInnerModifier(t *testing.T) {
	wiring := newWiringSpec("TestBasicClientPoolInnerModifier")

	leaf := workflow.Define(wiring, "leaf", "TestLeafServiceImpl")
	nonleaf := workflow.Define(wiring, "nonleaf", "TestNonLeafService", leaf)

	clientpool.Create(wiring, leaf, 7)
	retries.AddRetries(wiring, leaf, 10)

	grpc.Deploy(wiring, leaf)
	grpc.Deploy(wiring, nonleaf)

	leafproc := goproc.CreateProcess(wiring, "leafproc", leaf)
	nonleafproc := goproc.CreateProcess(wiring, "nonleafproc", nonleaf)

	app := assertBuildSuccess(t, wiring, leafproc, nonleafproc)

	assertIR(t, app,
		`TestBasicClientPoolInnerModifier = BlueprintApplication() {
			leaf.grpc.addr = GolangServerAddress()
			leaf.handler.visibility
			leafproc = GolangProcessNode(leaf.grpc.addr) {
			  leaf = TestLeafService()
			  leaf.grpc_server = GRPCServer(leaf, leaf.grpc.addr)
			}
			nonleaf.grpc.addr = GolangServerAddress()
			nonleaf.handler.visibility
			nonleafproc = GolangProcessNode(nonleaf.grpc.addr, leaf.grpc.addr) {
			  leaf.clientpool = ClientPool(leaf.client.retrier, 7) {
				leaf.grpc_client = GRPCClient(leaf.grpc.addr)
				leaf.client.retrier = Retrier(leaf.grpc_client)
			  }
			  nonleaf = TestNonLeafService(leaf.clientpool)
			  nonleaf.grpc_server = GRPCServer(nonleaf, nonleaf.grpc.addr)
			}
		  }`)

}
func TestBasicClientPoolOuterModifier(t *testing.T) {
	wiring := newWiringSpec("TestBasicClientPoolOuterModifier")

	leaf := workflow.Define(wiring, "leaf", "TestLeafServiceImpl")
	nonleaf := workflow.Define(wiring, "nonleaf", "TestNonLeafService", leaf)

	retries.AddRetries(wiring, leaf, 10)
	clientpool.Create(wiring, leaf, 7)

	grpc.Deploy(wiring, leaf)
	grpc.Deploy(wiring, nonleaf)

	leafproc := goproc.CreateProcess(wiring, "leafproc", leaf)
	nonleafproc := goproc.CreateProcess(wiring, "nonleafproc", nonleaf)

	app := assertBuildSuccess(t, wiring, leafproc, nonleafproc)

	assertIR(t, app,
		`TestBasicClientPoolOuterModifier = BlueprintApplication() {
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
			  leaf.client.retrier = Retrier(leaf.clientpool)
			  nonleaf = TestNonLeafService(leaf.client.retrier)
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
