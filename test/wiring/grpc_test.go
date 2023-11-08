package wiring

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/goproc"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/grpc"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/simplecache"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/workflow"
)

/*
Tests for correct IR layout from wiring spec helper functions for GRPC
*/

func TestServicesOverGRPCNoProcess(t *testing.T) {
	wiring := newWiringSpec("TestServicesOverGRPCNoProcess")

	leaf := workflow.Define(wiring, "leaf", "TestLeafServiceImpl")
	nonleaf := workflow.Define(wiring, "nonleaf", "TestNonLeafService", leaf)

	grpc.Deploy(wiring, leaf)
	grpc.Deploy(wiring, nonleaf)

	app := assertBuildSuccess(t, wiring, leaf, nonleaf)

	assertIR(t, app,
		`TestServicesOverGRPCNoProcess = BlueprintApplication() {
			leaf.grpc.addr
			leaf.grpc.dial_addr = AddressConfig()
			leaf.grpc_client = GRPCClient(leaf.grpc.dial_addr)
			nonleaf.grpc.addr
			nonleaf.grpc.dial_addr = AddressConfig()
			nonleaf.grpc_client = GRPCClient(nonleaf.grpc.dial_addr)
			leaf.grpc.bind_addr = AddressConfig()
			leaf.handler.visibility
			leaf = TestLeafService()
			leaf.grpc_server = GRPCServer(leaf, leaf.grpc.bind_addr)
			nonleaf.grpc.bind_addr = AddressConfig()
			nonleaf.handler.visibility
			nonleaf = TestNonLeafService(leaf.grpc_client)
			nonleaf.grpc_server = GRPCServer(nonleaf, nonleaf.grpc.bind_addr)
		  }`)
}

func TestServicesOverGRPCSameProcess(t *testing.T) {
	wiring := newWiringSpec("TestServicesOverGRPCSameProcess")

	leaf := workflow.Define(wiring, "leaf", "TestLeafServiceImpl")
	nonleaf := workflow.Define(wiring, "nonleaf", "TestNonLeafService", leaf)

	grpc.Deploy(wiring, leaf)
	grpc.Deploy(wiring, nonleaf)

	myproc := goproc.CreateProcess(wiring, "myproc", leaf, nonleaf)

	app := assertBuildSuccess(t, wiring, myproc)

	assertIR(t, app,
		`TestServicesOverGRPCSameProcess = BlueprintApplication() {
			leaf.grpc.addr
			leaf.grpc.bind_addr = AddressConfig()
			leaf.handler.visibility
			nonleaf.grpc.addr
			nonleaf.grpc.bind_addr = AddressConfig()
			nonleaf.handler.visibility
			leaf.grpc.dial_addr = AddressConfig()
			myproc = GolangProcessNode(leaf.grpc.bind_addr, nonleaf.grpc.bind_addr, leaf.grpc.dial_addr) {
			  leaf = TestLeafService()
			  leaf.grpc_server = GRPCServer(leaf, leaf.grpc.bind_addr)
			  leaf.grpc_client = GRPCClient(leaf.grpc.dial_addr)
			  nonleaf = TestNonLeafService(leaf.grpc_client)
			  nonleaf.grpc_server = GRPCServer(nonleaf, nonleaf.grpc.bind_addr)
			}
		  }`)
}

func TestBasicServicesOverGRPCDifferentProcesses(t *testing.T) {
	wiring := newWiringSpec("TestBasicServicesOverGRPCDifferentProcesses")

	leaf := workflow.Define(wiring, "leaf", "TestLeafServiceImpl")
	nonleaf := workflow.Define(wiring, "nonleaf", "TestNonLeafService", leaf)

	grpc.Deploy(wiring, leaf)
	grpc.Deploy(wiring, nonleaf)

	leafproc := goproc.CreateProcess(wiring, "leafproc", leaf)
	nonleafproc := goproc.CreateProcess(wiring, "nonleafproc", nonleaf)

	app := assertBuildSuccess(t, wiring, leafproc, nonleafproc)

	assertIR(t, app,
		`TestBasicServicesOverGRPCDifferentProcesses = BlueprintApplication() {
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
			  leaf.grpc_client = GRPCClient(leaf.grpc.dial_addr)
			  nonleaf = TestNonLeafService(leaf.grpc_client)
			  nonleaf.grpc_server = GRPCServer(nonleaf, nonleaf.grpc.bind_addr)
			}
		  }`)
}

func TestReachabilityErrorForServiceNotDeployedWithGRPC(t *testing.T) {
	wiring := newWiringSpec("TestReachabilityErrorForServiceNotDeployedWithGRPC")

	leaf := workflow.Define(wiring, "leaf", "TestLeafServiceImpl")
	nonleaf := workflow.Define(wiring, "nonleaf", "TestNonLeafService", leaf)

	grpc.Deploy(wiring, nonleaf)

	leafproc := goproc.CreateProcess(wiring, "leafproc", leaf)
	nonleafproc := goproc.CreateProcess(wiring, "nonleafproc", nonleaf)

	err := assertBuildFailure(t, wiring, leafproc, nonleafproc)
	assert.Contains(t, err.Error(), "reachability error")
}

func TestNoReachabilityErrorForServiceNotDeployedWithGRPC(t *testing.T) {
	wiring := newWiringSpec("TestNoReachabilityErrorForServiceNotDeployedWithGRPC")

	leaf := workflow.Define(wiring, "leaf", "TestLeafServiceImpl")
	nonleaf := workflow.Define(wiring, "nonleaf", "TestNonLeafService", leaf)

	grpc.Deploy(wiring, leaf)

	leafproc := goproc.CreateProcess(wiring, "leafproc", leaf)
	nonleafproc := goproc.CreateProcess(wiring, "nonleafproc", nonleaf)

	app := assertBuildSuccess(t, wiring, leafproc, nonleafproc)

	assertIR(t, app,
		`TestNoReachabilityErrorForServiceNotDeployedWithGRPC = BlueprintApplication() {
			leaf.grpc.addr
			leaf.grpc.bind_addr = AddressConfig()
			leaf.handler.visibility
			leafproc = GolangProcessNode(leaf.grpc.bind_addr) {
			  leaf = TestLeafService()
			  leaf.grpc_server = GRPCServer(leaf, leaf.grpc.bind_addr)
			}
			nonleaf.handler.visibility
			leaf.grpc.dial_addr = AddressConfig()
			nonleafproc = GolangProcessNode(leaf.grpc.dial_addr) {
			  leaf.grpc_client = GRPCClient(leaf.grpc.dial_addr)
			  nonleaf = TestNonLeafService(leaf.grpc_client)
			}
		  }`)
}

func TestClientProc(t *testing.T) {
	wiring := newWiringSpec("TestNoReachabilityErrorForServiceNotDeployedWithGRPC")

	leaf := workflow.Define(wiring, "leaf", "TestLeafServiceImpl")
	nonleaf := workflow.Define(wiring, "nonleaf", "TestNonLeafService", leaf)

	grpc.Deploy(wiring, leaf)
	grpc.Deploy(wiring, nonleaf)

	leafproc := goproc.CreateProcess(wiring, "leafproc", leaf)
	nonleafproc := goproc.CreateProcess(wiring, "nonleafproc", nonleaf)

	leafclient := goproc.CreateClientProcess(wiring, "leafclient", nonleaf)

	app := assertBuildSuccess(t, wiring, leafproc, nonleafproc, leafclient)

	assertIR(t, app,
		`TestNoReachabilityErrorForServiceNotDeployedWithGRPC = BlueprintApplication() {
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
			  leaf.grpc_client = GRPCClient(leaf.grpc.dial_addr)
			  nonleaf = TestNonLeafService(leaf.grpc_client)
			  nonleaf.grpc_server = GRPCServer(nonleaf, nonleaf.grpc.bind_addr)
			}
			nonleaf.grpc.dial_addr = AddressConfig()
			leafclient = GolangProcessNode(nonleaf.grpc.dial_addr) {
			  nonleaf.grpc_client = GRPCClient(nonleaf.grpc.dial_addr)
			}
		  }`)
}

func TestImplicitServicesInSameProcWithGRPC(t *testing.T) {
	wiring := newWiringSpec("TestImplicitServicesInSameProcWithGRPC")

	leaf := workflow.Define(wiring, "leaf", "TestLeafServiceImpl")
	nonleaf := workflow.Define(wiring, "nonleaf", "TestNonLeafService", leaf)

	grpc.Deploy(wiring, leaf)
	grpc.Deploy(wiring, nonleaf)

	leafclient := goproc.CreateClientProcess(wiring, "leafclient", nonleaf)

	app := assertBuildSuccess(t, wiring, leafclient)

	assertIR(t, app,
		`TestImplicitServicesInSameProcWithGRPC = BlueprintApplication() {
			nonleaf.grpc.addr
			nonleaf.grpc.dial_addr = AddressConfig()
			leafclient = GolangProcessNode(nonleaf.grpc.dial_addr, nonleaf.grpc.bind_addr, leaf.grpc.dial_addr, leaf.grpc.bind_addr) {
			  nonleaf.grpc_client = GRPCClient(nonleaf.grpc.dial_addr)
			  leaf.grpc_client = GRPCClient(leaf.grpc.dial_addr)
			  nonleaf = TestNonLeafService(leaf.grpc_client)
			  nonleaf.grpc_server = GRPCServer(nonleaf, nonleaf.grpc.bind_addr)
			  leaf = TestLeafService()
			  leaf.grpc_server = GRPCServer(leaf, leaf.grpc.bind_addr)
			}
			nonleaf.grpc.bind_addr = AddressConfig()
			nonleaf.handler.visibility
			leaf.grpc.addr
			leaf.grpc.dial_addr = AddressConfig()
			leaf.grpc.bind_addr = AddressConfig()
			leaf.handler.visibility
		  }`)
}

func TestImplicitServicesInSameProcPartialGRPC(t *testing.T) {
	wiring := newWiringSpec("TestImplicitServicesInSameProcPartialGRPC")

	leaf := workflow.Define(wiring, "leaf", "TestLeafServiceImpl")
	nonleaf := workflow.Define(wiring, "nonleaf", "TestNonLeafService", leaf)

	grpc.Deploy(wiring, nonleaf)

	leafclient := goproc.CreateClientProcess(wiring, "leafclient", nonleaf)

	app := assertBuildSuccess(t, wiring, leafclient)

	assertIR(t, app,
		`TestImplicitServicesInSameProcPartialGRPC = BlueprintApplication() {
			nonleaf.grpc.addr
			nonleaf.grpc.dial_addr = AddressConfig()
			leafclient = GolangProcessNode(nonleaf.grpc.dial_addr, nonleaf.grpc.bind_addr) {
			  nonleaf.grpc_client = GRPCClient(nonleaf.grpc.dial_addr)
			  leaf = TestLeafService()
			  nonleaf = TestNonLeafService(leaf)
			  nonleaf.grpc_server = GRPCServer(nonleaf, nonleaf.grpc.bind_addr)
			}
			nonleaf.grpc.bind_addr = AddressConfig()
			nonleaf.handler.visibility
			leaf.handler.visibility
		  }`)
}

func TestImplicitCacheInSameProc(t *testing.T) {
	wiring := newWiringSpec("TestImplicitCacheInSameProc")

	leaf_cache := simplecache.Define(wiring, "leaf_cache")
	leaf := workflow.Define(wiring, "leaf", "TestLeafServiceImplWithCache", leaf_cache)
	nonleaf := workflow.Define(wiring, "nonleaf", "TestNonLeafService", leaf)

	grpc.Deploy(wiring, leaf)
	grpc.Deploy(wiring, nonleaf)

	leafproc := goproc.CreateProcess(wiring, "leafproc", leaf)
	nonleafproc := goproc.CreateProcess(wiring, "nonleafproc", nonleaf)

	leafclient := goproc.CreateClientProcess(wiring, "leafclient", nonleaf)

	app := assertBuildSuccess(t, wiring, leafproc, nonleafproc, leafclient)

	assertIR(t, app,
		`TestImplicitCacheInSameProc = BlueprintApplication() {
			leaf.grpc.addr
			leaf.grpc.bind_addr = AddressConfig()
			leaf.handler.visibility
			leaf_cache.backend.visibility
			leafproc = GolangProcessNode(leaf.grpc.bind_addr) {
			  leaf_cache = SimpleCache()
			  leaf = TestLeafService(leaf_cache)
			  leaf.grpc_server = GRPCServer(leaf, leaf.grpc.bind_addr)
			}
			nonleaf.grpc.addr
			nonleaf.grpc.bind_addr = AddressConfig()
			nonleaf.handler.visibility
			leaf.grpc.dial_addr = AddressConfig()
			nonleafproc = GolangProcessNode(nonleaf.grpc.bind_addr, leaf.grpc.dial_addr) {
			  leaf.grpc_client = GRPCClient(leaf.grpc.dial_addr)
			  nonleaf = TestNonLeafService(leaf.grpc_client)
			  nonleaf.grpc_server = GRPCServer(nonleaf, nonleaf.grpc.bind_addr)
			}
			nonleaf.grpc.dial_addr = AddressConfig()
			leafclient = GolangProcessNode(nonleaf.grpc.dial_addr) {
			  nonleaf.grpc_client = GRPCClient(nonleaf.grpc.dial_addr)
			}
		  }`)

}
