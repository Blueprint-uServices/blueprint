package wiring

import (
	"testing"

	"github.com/blueprint-uservices/blueprint/plugins/clientpool"
	"github.com/blueprint-uservices/blueprint/plugins/goproc"
	"github.com/blueprint-uservices/blueprint/plugins/grpc"
	"github.com/blueprint-uservices/blueprint/plugins/retries"
	"github.com/blueprint-uservices/blueprint/plugins/workflow"
	wf "github.com/blueprint-uservices/blueprint/test/workflow/workflow"
)

func TestBasicClientPool(t *testing.T) {
	spec := newWiringSpec("TestBasicClientPool")

	leaf := workflow.Service[*wf.TestLeafServiceImpl](spec, "leaf")
	nonleaf := workflow.Service[wf.TestNonLeafService](spec, "nonleaf", leaf)

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
			  leaf.client = leaf.clientpool
			  leaf.clientpool = ClientPool(leaf.grpc_client, 7) {
				leaf.grpc_client = GRPCClient(leaf.grpc.dial_addr)
			  }
			  nonleaf = TestNonLeafService(leaf.client)
			  nonleaf.grpc_server = GRPCServer(nonleaf, nonleaf.grpc.bind_addr)
			  nonleafproc.logger = SLogger()
			  nonleafproc.stdoutmetriccollector = StdoutMetricCollector()
			}
		  }`)

}

func TestBasicClientPoolInnerModifier(t *testing.T) {
	spec := newWiringSpec("TestBasicClientPoolInnerModifier")

	leaf := workflow.Service[*wf.TestLeafServiceImpl](spec, "leaf")
	nonleaf := workflow.Service[wf.TestNonLeafService](spec, "nonleaf", leaf)

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
			  leaf.client = leaf.clientpool
			  leaf.clientpool = ClientPool(leaf.client.retrier, 7) {
				leaf.grpc_client = GRPCClient(leaf.grpc.dial_addr)
				leaf.client.retrier = Retrier(leaf.grpc_client)
			  }
			  nonleaf = TestNonLeafService(leaf.client)
			  nonleaf.grpc_server = GRPCServer(nonleaf, nonleaf.grpc.bind_addr)
			  nonleafproc.logger = SLogger()
			  nonleafproc.stdoutmetriccollector = StdoutMetricCollector()
			}
		  }`)

}
func TestBasicClientPoolOuterModifier(t *testing.T) {
	spec := newWiringSpec("TestBasicClientPoolOuterModifier")

	leaf := workflow.Service[*wf.TestLeafServiceImpl](spec, "leaf")
	nonleaf := workflow.Service[wf.TestNonLeafService](spec, "nonleaf", leaf)

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
			  leaf.client = leaf.client.retrier
			  leaf.client.retrier = Retrier(leaf.clientpool)
			  leaf.clientpool = ClientPool(leaf.grpc_client, 7) {
				leaf.grpc_client = GRPCClient(leaf.grpc.dial_addr)
			  }
			  nonleaf = TestNonLeafService(leaf.client)
			  nonleaf.grpc_server = GRPCServer(nonleaf, nonleaf.grpc.bind_addr)
			  nonleafproc.logger = SLogger()
			  nonleafproc.stdoutmetriccollector = StdoutMetricCollector()
			}
		  }`)

}

func TestInvalidModifierOrder(t *testing.T) {
	spec := newWiringSpec("TestBasicClientPool")

	leaf := workflow.Service[*wf.TestLeafServiceImpl](spec, "leaf")
	nonleaf := workflow.Service[wf.TestNonLeafService](spec, "nonleaf", leaf)

	grpc.Deploy(spec, leaf)
	clientpool.Create(spec, leaf, 7)

	leafproc := goproc.CreateProcess(spec, "leafproc", leaf)
	nonleafproc := goproc.CreateProcess(spec, "nonleafproc", nonleaf)

	assertBuildFailure(t, spec, leafproc, nonleafproc)

}
