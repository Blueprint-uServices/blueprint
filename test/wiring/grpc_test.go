package wiring

import (
	"testing"

	"github.com/blueprint-uservices/blueprint/plugins/goproc"
	"github.com/blueprint-uservices/blueprint/plugins/grpc"
	"github.com/blueprint-uservices/blueprint/plugins/simple"
	"github.com/blueprint-uservices/blueprint/plugins/workflow"
	"github.com/blueprint-uservices/blueprint/test/workflow/cache"
	wf "github.com/blueprint-uservices/blueprint/test/workflow/workflow"
	"github.com/stretchr/testify/assert"
)

/*
Tests for correct IR layout from wiring spec helper functions for GRPC
*/

func TestServicesOverGRPCNoProcess(t *testing.T) {
	spec := newWiringSpec("TestServicesOverGRPCNoProcess")

	leaf := workflow.Service[*wf.TestLeafServiceImpl](spec, "leaf")
	nonleaf := workflow.Service[wf.TestNonLeafService](spec, "nonleaf", leaf)

	grpc.Deploy(spec, leaf)
	grpc.Deploy(spec, nonleaf)

	app := assertBuildSuccess(t, spec, leaf, nonleaf)

	assertIR(t, app,
		`TestServicesOverGRPCNoProcess = BlueprintApplication() {
			leaf = TestLeafService()
			leaf.client = leaf.grpc_client
			leaf.grpc.addr
			leaf.grpc.bind_addr = AddressConfig()
			leaf.grpc.dial_addr = AddressConfig()
			leaf.grpc_client = GRPCClient(leaf.grpc.dial_addr)
			leaf.grpc_server = GRPCServer(leaf, leaf.grpc.bind_addr)
			leaf.handler.visibility
			nonleaf = TestNonLeafService(leaf.client)
			nonleaf.client = nonleaf.grpc_client
			nonleaf.grpc.addr
			nonleaf.grpc.bind_addr = AddressConfig()
			nonleaf.grpc.dial_addr = AddressConfig()
			nonleaf.grpc_client = GRPCClient(nonleaf.grpc.dial_addr)
			nonleaf.grpc_server = GRPCServer(nonleaf, nonleaf.grpc.bind_addr)
			nonleaf.handler.visibility
		  }`)
}

func TestServicesOverGRPCSameProcess(t *testing.T) {
	spec := newWiringSpec("TestServicesOverGRPCSameProcess")

	leaf := workflow.Service[*wf.TestLeafServiceImpl](spec, "leaf")
	nonleaf := workflow.Service[wf.TestNonLeafService](spec, "nonleaf", leaf)

	grpc.Deploy(spec, leaf)
	grpc.Deploy(spec, nonleaf)

	myproc := goproc.CreateProcess(spec, "myproc", leaf, nonleaf)

	app := assertBuildSuccess(t, spec, myproc)

	assertIR(t, app,
		`TestServicesOverGRPCSameProcess = BlueprintApplication() {
			leaf.grpc.addr
			leaf.grpc.bind_addr = AddressConfig()
			leaf.grpc.dial_addr = AddressConfig()
			leaf.handler.visibility
			myproc = GolangProcessNode(leaf.grpc.bind_addr, leaf.grpc.dial_addr, nonleaf.grpc.bind_addr) {
			  leaf = TestLeafService()
			  leaf.client = leaf.grpc_client
			  leaf.grpc_client = GRPCClient(leaf.grpc.dial_addr)
			  leaf.grpc_server = GRPCServer(leaf, leaf.grpc.bind_addr)
			  myproc.logger = SLogger()
			  myproc.stdoutmetriccollector = StdoutMetricCollector()
			  nonleaf = TestNonLeafService(leaf.client)
			  nonleaf.grpc_server = GRPCServer(nonleaf, nonleaf.grpc.bind_addr)
			}
			nonleaf.grpc.addr
			nonleaf.grpc.bind_addr = AddressConfig()
			nonleaf.handler.visibility
		  }`)
}

func TestBasicServicesOverGRPCDifferentProcesses(t *testing.T) {
	spec := newWiringSpec("TestBasicServicesOverGRPCDifferentProcesses")

	leaf := workflow.Service[*wf.TestLeafServiceImpl](spec, "leaf")
	nonleaf := workflow.Service[wf.TestNonLeafService](spec, "nonleaf", leaf)

	grpc.Deploy(spec, leaf)
	grpc.Deploy(spec, nonleaf)

	leafproc := goproc.CreateProcess(spec, "leafproc", leaf)
	nonleafproc := goproc.CreateProcess(spec, "nonleafproc", nonleaf)

	app := assertBuildSuccess(t, spec, leafproc, nonleafproc)

	assertIR(t, app,
		`TestBasicServicesOverGRPCDifferentProcesses = BlueprintApplication() {
			leaf.grpc.addr
			leaf.grpc.bind_addr = AddressConfig()
			leaf.grpc.dial_addr = AddressConfig()
			leaf.handler.visibility
			leafproc = GolangProcessNode(leaf.grpc.bind_addr) {
			  leaf = TestLeafService()
			  leaf.grpc_server = GRPCServer(leaf, leaf.grpc.bind_addr)
			  leafproc.logger = SLogger()
			  leafproc.stdoutmetriccollector = StdoutMetricCollector()
			}
			nonleaf.grpc.addr
			nonleaf.grpc.bind_addr = AddressConfig()
			nonleaf.handler.visibility
			nonleafproc = GolangProcessNode(leaf.grpc.dial_addr, nonleaf.grpc.bind_addr) {
			  leaf.client = leaf.grpc_client
			  leaf.grpc_client = GRPCClient(leaf.grpc.dial_addr)
			  nonleaf = TestNonLeafService(leaf.client)
			  nonleaf.grpc_server = GRPCServer(nonleaf, nonleaf.grpc.bind_addr)
			  nonleafproc.logger = SLogger()
			  nonleafproc.stdoutmetriccollector = StdoutMetricCollector()
			}
		  }`)
}

func TestReachabilityErrorForServiceNotDeployedWithGRPC(t *testing.T) {
	spec := newWiringSpec("TestReachabilityErrorForServiceNotDeployedWithGRPC")

	leaf := workflow.Service[*wf.TestLeafServiceImpl](spec, "leaf")
	nonleaf := workflow.Service[wf.TestNonLeafService](spec, "nonleaf", leaf)

	grpc.Deploy(spec, nonleaf)

	leafproc := goproc.CreateProcess(spec, "leafproc", leaf)
	nonleafproc := goproc.CreateProcess(spec, "nonleafproc", nonleaf)

	err := assertBuildFailure(t, spec, leafproc, nonleafproc)
	assert.Contains(t, err.Error(), "reachability error")
}

func TestNoReachabilityErrorForServiceNotDeployedWithGRPC(t *testing.T) {
	spec := newWiringSpec("TestNoReachabilityErrorForServiceNotDeployedWithGRPC")

	leaf := workflow.Service[*wf.TestLeafServiceImpl](spec, "leaf")
	nonleaf := workflow.Service[wf.TestNonLeafService](spec, "nonleaf", leaf)

	grpc.Deploy(spec, leaf)

	leafproc := goproc.CreateProcess(spec, "leafproc", leaf)
	nonleafproc := goproc.CreateProcess(spec, "nonleafproc", nonleaf)

	app := assertBuildSuccess(t, spec, leafproc, nonleafproc)

	assertIR(t, app,
		`TestNoReachabilityErrorForServiceNotDeployedWithGRPC = BlueprintApplication() {
			leaf.grpc.addr
			leaf.grpc.bind_addr = AddressConfig()
			leaf.grpc.dial_addr = AddressConfig()
			leaf.handler.visibility
			leafproc = GolangProcessNode(leaf.grpc.bind_addr) {
			  leaf = TestLeafService()
			  leaf.grpc_server = GRPCServer(leaf, leaf.grpc.bind_addr)
			  leafproc.logger = SLogger()
			  leafproc.stdoutmetriccollector = StdoutMetricCollector()
			}
			nonleaf.handler.visibility
			nonleafproc = GolangProcessNode(leaf.grpc.dial_addr) {
			  leaf.client = leaf.grpc_client
			  leaf.grpc_client = GRPCClient(leaf.grpc.dial_addr)
			  nonleaf = TestNonLeafService(leaf.client)
			  nonleafproc.logger = SLogger()
			  nonleafproc.stdoutmetriccollector = StdoutMetricCollector()
			}
		  }`)
}

func TestClientProc(t *testing.T) {
	spec := newWiringSpec("TestNoReachabilityErrorForServiceNotDeployedWithGRPC")

	leaf := workflow.Service[*wf.TestLeafServiceImpl](spec, "leaf")
	nonleaf := workflow.Service[wf.TestNonLeafService](spec, "nonleaf", leaf)

	grpc.Deploy(spec, leaf)
	grpc.Deploy(spec, nonleaf)

	leafproc := goproc.CreateProcess(spec, "leafproc", leaf)
	nonleafproc := goproc.CreateProcess(spec, "nonleafproc", nonleaf)

	appclient := goproc.CreateClientProcess(spec, "appclient", nonleaf)

	app := assertBuildSuccess(t, spec, leafproc, nonleafproc, appclient)

	assertIR(t, app,
		`TestNoReachabilityErrorForServiceNotDeployedWithGRPC = BlueprintApplication() {
			appclient = GolangProcessNode(nonleaf.grpc.dial_addr) {
			  nonleaf.client = nonleaf.grpc_client
			  nonleaf.grpc_client = GRPCClient(nonleaf.grpc.dial_addr)
			}
			leaf.grpc.addr
			leaf.grpc.bind_addr = AddressConfig()
			leaf.grpc.dial_addr = AddressConfig()
			leaf.handler.visibility
			leafproc = GolangProcessNode(leaf.grpc.bind_addr) {
			  leaf = TestLeafService()
			  leaf.grpc_server = GRPCServer(leaf, leaf.grpc.bind_addr)
			  leafproc.logger = SLogger()
			  leafproc.stdoutmetriccollector = StdoutMetricCollector()
			}
			nonleaf.grpc.addr
			nonleaf.grpc.bind_addr = AddressConfig()
			nonleaf.grpc.dial_addr = AddressConfig()
			nonleaf.handler.visibility
			nonleafproc = GolangProcessNode(leaf.grpc.dial_addr, nonleaf.grpc.bind_addr) {
			  leaf.client = leaf.grpc_client
			  leaf.grpc_client = GRPCClient(leaf.grpc.dial_addr)
			  nonleaf = TestNonLeafService(leaf.client)
			  nonleaf.grpc_server = GRPCServer(nonleaf, nonleaf.grpc.bind_addr)
			  nonleafproc.logger = SLogger()
			  nonleafproc.stdoutmetriccollector = StdoutMetricCollector()
			}
		  }`)
}

func TestImplicitServicesInSameProcWithGRPC(t *testing.T) {
	spec := newWiringSpec("TestImplicitServicesInSameProcWithGRPC")

	leaf := workflow.Service[*wf.TestLeafServiceImpl](spec, "leaf")
	nonleaf := workflow.Service[wf.TestNonLeafService](spec, "nonleaf", leaf)

	grpc.Deploy(spec, leaf)
	grpc.Deploy(spec, nonleaf)

	leafclient := goproc.CreateClientProcess(spec, "leafclient", nonleaf)

	app := assertBuildSuccess(t, spec, leafclient)

	assertIR(t, app,
		`TestImplicitServicesInSameProcWithGRPC = BlueprintApplication() {
			leaf.grpc.addr
			leaf.grpc.bind_addr = AddressConfig()
			leaf.grpc.dial_addr = AddressConfig()
			leaf.handler.visibility
			leafclient = GolangProcessNode(leaf.grpc.bind_addr, leaf.grpc.dial_addr, nonleaf.grpc.bind_addr, nonleaf.grpc.dial_addr) {
			  leaf = TestLeafService()
			  leaf.client = leaf.grpc_client
			  leaf.grpc_client = GRPCClient(leaf.grpc.dial_addr)
			  leaf.grpc_server = GRPCServer(leaf, leaf.grpc.bind_addr)
			  nonleaf = TestNonLeafService(leaf.client)
			  nonleaf.client = nonleaf.grpc_client
			  nonleaf.grpc_client = GRPCClient(nonleaf.grpc.dial_addr)
			  nonleaf.grpc_server = GRPCServer(nonleaf, nonleaf.grpc.bind_addr)
			}
			nonleaf.grpc.addr
			nonleaf.grpc.bind_addr = AddressConfig()
			nonleaf.grpc.dial_addr = AddressConfig()
			nonleaf.handler.visibility
		  }`)
}

func TestImplicitServicesInSameProcPartialGRPC(t *testing.T) {
	spec := newWiringSpec("TestImplicitServicesInSameProcPartialGRPC")

	leaf := workflow.Service[*wf.TestLeafServiceImpl](spec, "leaf")
	nonleaf := workflow.Service[wf.TestNonLeafService](spec, "nonleaf", leaf)

	grpc.Deploy(spec, nonleaf)

	leafclient := goproc.CreateClientProcess(spec, "leafclient", nonleaf)

	app := assertBuildSuccess(t, spec, leafclient)

	assertIR(t, app,
		`TestImplicitServicesInSameProcPartialGRPC = BlueprintApplication() {
			leaf.handler.visibility
			leafclient = GolangProcessNode(nonleaf.grpc.bind_addr, nonleaf.grpc.dial_addr) {
			  leaf = TestLeafService()
			  leaf.client = leaf
			  nonleaf = TestNonLeafService(leaf.client)
			  nonleaf.client = nonleaf.grpc_client
			  nonleaf.grpc_client = GRPCClient(nonleaf.grpc.dial_addr)
			  nonleaf.grpc_server = GRPCServer(nonleaf, nonleaf.grpc.bind_addr)
			}
			nonleaf.grpc.addr
			nonleaf.grpc.bind_addr = AddressConfig()
			nonleaf.grpc.dial_addr = AddressConfig()
			nonleaf.handler.visibility
		  }`)
}

func TestImplicitCacheInSameProc(t *testing.T) {
	spec := newWiringSpec("TestImplicitCacheInSameProc")

	leaf_cache := simple.Cache(spec, "leaf_cache")
	leaf := workflow.Service[*cache.TestLeafServiceImplWithCache](spec, "leaf", leaf_cache)
	nonleaf := workflow.Service[wf.TestNonLeafService](spec, "nonleaf", leaf)

	grpc.Deploy(spec, leaf)
	grpc.Deploy(spec, nonleaf)

	leafproc := goproc.CreateProcess(spec, "leafproc", leaf)
	nonleafproc := goproc.CreateProcess(spec, "nonleafproc", nonleaf)

	leafclient := goproc.CreateClientProcess(spec, "leafclient", nonleaf)

	app := assertBuildSuccess(t, spec, leafproc, nonleafproc, leafclient)

	assertIR(t, app,
		`TestImplicitCacheInSameProc = BlueprintApplication() {
			leaf.grpc.addr
			leaf.grpc.bind_addr = AddressConfig()
			leaf.grpc.dial_addr = AddressConfig()
			leaf.handler.visibility
			leaf_cache.backend.visibility
			leafclient = GolangProcessNode(nonleaf.grpc.dial_addr) {
			  nonleaf.client = nonleaf.grpc_client
			  nonleaf.grpc_client = GRPCClient(nonleaf.grpc.dial_addr)
			}
			leafproc = GolangProcessNode(leaf.grpc.bind_addr) {
			  leaf = TestLeafService(leaf_cache)
			  leaf.grpc_server = GRPCServer(leaf, leaf.grpc.bind_addr)
			  leaf_cache = SimpleCache()
			  leafproc.logger = SLogger()
			  leafproc.stdoutmetriccollector = StdoutMetricCollector()
			}
			nonleaf.grpc.addr
			nonleaf.grpc.bind_addr = AddressConfig()
			nonleaf.grpc.dial_addr = AddressConfig()
			nonleaf.handler.visibility
			nonleafproc = GolangProcessNode(leaf.grpc.dial_addr, nonleaf.grpc.bind_addr) {
			  leaf.client = leaf.grpc_client
			  leaf.grpc_client = GRPCClient(leaf.grpc.dial_addr)
			  nonleaf = TestNonLeafService(leaf.client)
			  nonleaf.grpc_server = GRPCServer(nonleaf, nonleaf.grpc.bind_addr)
			  nonleafproc.logger = SLogger()
			  nonleafproc.stdoutmetriccollector = StdoutMetricCollector()
			}
		  }`)

}
