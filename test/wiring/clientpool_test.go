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
	spec := newWiringSpec("TestBasicClientPool")

	leaf := workflow.Define(spec, "leaf", "TestLeafServiceImpl")
	nonleaf := workflow.Define(spec, "nonleaf", "TestNonLeafService", leaf)

	clientpool.Create(spec, leaf, 7)

	grpc.Deploy(spec, leaf)
	grpc.Deploy(spec, nonleaf)

	leafproc := goproc.CreateProcess(spec, "leafproc", leaf)
	nonleafproc := goproc.CreateProcess(spec, "nonleafproc", nonleaf)

	app := assertBuildSuccess(t, spec, leafproc, nonleafproc)

	assertIR(t, app,
		`TestBasicClientPool = BlueprintApplication() {
			leaf.grpc.addr
			leaf.grpc.bind_addr = AddressConfig()
			leaf.handler.visibility
			leafproc = GolangProcessNode(leaf.grpc.bind_addr) {
			  leaf = TestLeafService()
			  leaf.grpc_server = GRPCServer(leaf, leaf.grpc.bind_addr)
			}
			nonleaf.grpc.addr
			nonleaf.grpc.bind_addr = AddressConfig()
			nonleaf.handler.visibility
			leaf.grpc.dial_addr = AddressConfig()
			nonleafproc = GolangProcessNode(nonleaf.grpc.bind_addr, leaf.grpc.dial_addr) {
			  leaf.clientpool = ClientPool(leaf.grpc_client, 7) {
				leaf.grpc_client = GRPCClient(leaf.grpc.dial_addr)
			  }
			  nonleaf = TestNonLeafService(leaf.clientpool)
			  nonleaf.grpc_server = GRPCServer(nonleaf, nonleaf.grpc.bind_addr)
			}
		  }`)

}

func TestBasicClientPoolInnerModifier(t *testing.T) {
	spec := newWiringSpec("TestBasicClientPoolInnerModifier")

	leaf := workflow.Define(spec, "leaf", "TestLeafServiceImpl")
	nonleaf := workflow.Define(spec, "nonleaf", "TestNonLeafService", leaf)

	clientpool.Create(spec, leaf, 7)
	retries.AddRetries(spec, leaf, 10)

	grpc.Deploy(spec, leaf)
	grpc.Deploy(spec, nonleaf)

	leafproc := goproc.CreateProcess(spec, "leafproc", leaf)
	nonleafproc := goproc.CreateProcess(spec, "nonleafproc", nonleaf)

	app := assertBuildSuccess(t, spec, leafproc, nonleafproc)

	assertIR(t, app,
		`TestBasicClientPoolInnerModifier = BlueprintApplication() {
			leaf.grpc.addr
			leaf.grpc.bind_addr = AddressConfig()
			leaf.handler.visibility
			leafproc = GolangProcessNode(leaf.grpc.bind_addr) {
			  leaf = TestLeafService()
			  leaf.grpc_server = GRPCServer(leaf, leaf.grpc.bind_addr)
			}
			nonleaf.grpc.addr
			nonleaf.grpc.bind_addr = AddressConfig()
			nonleaf.handler.visibility
			leaf.grpc.dial_addr = AddressConfig()
			nonleafproc = GolangProcessNode(nonleaf.grpc.bind_addr, leaf.grpc.dial_addr) {
			  leaf.clientpool = ClientPool(leaf.client.retrier, 7) {
				leaf.grpc_client = GRPCClient(leaf.grpc.dial_addr)
				leaf.client.retrier = Retrier(leaf.grpc_client)
			  }
			  nonleaf = TestNonLeafService(leaf.clientpool)
			  nonleaf.grpc_server = GRPCServer(nonleaf, nonleaf.grpc.bind_addr)
			}
		  }`)

}
func TestBasicClientPoolOuterModifier(t *testing.T) {
	spec := newWiringSpec("TestBasicClientPoolOuterModifier")

	leaf := workflow.Define(spec, "leaf", "TestLeafServiceImpl")
	nonleaf := workflow.Define(spec, "nonleaf", "TestNonLeafService", leaf)

	retries.AddRetries(spec, leaf, 10)
	clientpool.Create(spec, leaf, 7)

	grpc.Deploy(spec, leaf)
	grpc.Deploy(spec, nonleaf)

	leafproc := goproc.CreateProcess(spec, "leafproc", leaf)
	nonleafproc := goproc.CreateProcess(spec, "nonleafproc", nonleaf)

	app := assertBuildSuccess(t, spec, leafproc, nonleafproc)

	assertIR(t, app,
		`TestBasicClientPoolOuterModifier = BlueprintApplication() {
			leaf.grpc.addr
			leaf.grpc.bind_addr = AddressConfig()
			leaf.handler.visibility
			leafproc = GolangProcessNode(leaf.grpc.bind_addr) {
			  leaf = TestLeafService()
			  leaf.grpc_server = GRPCServer(leaf, leaf.grpc.bind_addr)
			}
			nonleaf.grpc.addr
			nonleaf.grpc.bind_addr = AddressConfig()
			nonleaf.handler.visibility
			leaf.grpc.dial_addr = AddressConfig()
			nonleafproc = GolangProcessNode(nonleaf.grpc.bind_addr, leaf.grpc.dial_addr) {
			  leaf.clientpool = ClientPool(leaf.grpc_client, 7) {
				leaf.grpc_client = GRPCClient(leaf.grpc.dial_addr)
			  }
			  leaf.client.retrier = Retrier(leaf.clientpool)
			  nonleaf = TestNonLeafService(leaf.client.retrier)
			  nonleaf.grpc_server = GRPCServer(nonleaf, nonleaf.grpc.bind_addr)
			}
		  }`)

}

func TestInvalidModifierOrder(t *testing.T) {
	spec := newWiringSpec("TestBasicClientPool")

	leaf := workflow.Define(spec, "leaf", "TestLeafServiceImpl")
	nonleaf := workflow.Define(spec, "nonleaf", "TestNonLeafService", leaf)

	grpc.Deploy(spec, leaf)
	clientpool.Create(spec, leaf, 7)

	leafproc := goproc.CreateProcess(spec, "leafproc", leaf)
	nonleafproc := goproc.CreateProcess(spec, "nonleafproc", nonleaf)

	assertBuildFailure(t, spec, leafproc, nonleafproc)

}
