package wiring

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/goproc"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/grpc"
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
			leaf.grpc.addr = GolangServerAddress()
			leaf.grpc_client = GRPCClient(leaf.grpc.addr)
			nonleaf.grpc.addr = GolangServerAddress()
			nonleaf.grpc_client = GRPCClient(nonleaf.grpc.addr)
			leaf.handler.visibility
			leaf = TestLeafService()
			leaf.grpc_server = GRPCServer(leaf, leaf.grpc.addr)
			nonleaf.handler.visibility
			nonleaf = TestNonLeafService(leaf.grpc_client)
			nonleaf.grpc_server = GRPCServer(nonleaf, nonleaf.grpc.addr)
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
			leaf.grpc.addr = GolangServerAddress()
			leaf.handler.visibility
			nonleaf.grpc.addr = GolangServerAddress()
			nonleaf.handler.visibility
			myproc = GolangProcessNode(leaf.grpc.addr, nonleaf.grpc.addr) {
			  leaf = TestLeafService()
			  leaf.grpc_server = GRPCServer(leaf, leaf.grpc.addr)
			  leaf.grpc_client = GRPCClient(leaf.grpc.addr)
			  nonleaf = TestNonLeafService(leaf.grpc_client)
			  nonleaf.grpc_server = GRPCServer(nonleaf, nonleaf.grpc.addr)
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
			leaf.grpc.addr = GolangServerAddress()
			leaf.handler.visibility
			leafproc = GolangProcessNode(leaf.grpc.addr) {
			  leaf = TestLeafService()
			  leaf.grpc_server = GRPCServer(leaf, leaf.grpc.addr)
			}
			nonleaf.grpc.addr = GolangServerAddress()
			nonleaf.handler.visibility
			nonleafproc = GolangProcessNode(nonleaf.grpc.addr, leaf.grpc.addr) {
			  leaf.grpc_client = GRPCClient(leaf.grpc.addr)
			  nonleaf = TestNonLeafService(leaf.grpc_client)
			  nonleaf.grpc_server = GRPCServer(nonleaf, nonleaf.grpc.addr)
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
			leaf.grpc.addr = GolangServerAddress()
			leaf.handler.visibility
			leafproc = GolangProcessNode(leaf.grpc.addr) {
			  leaf = TestLeafService()
			  leaf.grpc_server = GRPCServer(leaf, leaf.grpc.addr)
			}
			nonleaf.handler.visibility
			nonleafproc = GolangProcessNode(leaf.grpc.addr) {
			  leaf.grpc_client = GRPCClient(leaf.grpc.addr)
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
			leaf.grpc.addr = GolangServerAddress()
			leaf.handler.visibility
			leafproc = GolangProcessNode(leaf.grpc.addr) {
			  leaf = TestLeafService()
			  leaf.grpc_server = GRPCServer(leaf, leaf.grpc.addr)
			}
			nonleaf.grpc.addr = GolangServerAddress()
			nonleaf.handler.visibility
			nonleafproc = GolangProcessNode(nonleaf.grpc.addr, leaf.grpc.addr) {
			  leaf.grpc_client = GRPCClient(leaf.grpc.addr)
			  nonleaf = TestNonLeafService(leaf.grpc_client)
			  nonleaf.grpc_server = GRPCServer(nonleaf, nonleaf.grpc.addr)
			}
			leafclient = GolangProcessNode(nonleaf.grpc.addr) {
			  nonleaf.grpc_client = GRPCClient(nonleaf.grpc.addr)
			}
		  }`)
}
