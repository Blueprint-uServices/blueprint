package wiring

import (
	"testing"

	"github.com/blueprint-uservices/blueprint/plugins/goproc"
	"github.com/blueprint-uservices/blueprint/plugins/grpc"
	"github.com/blueprint-uservices/blueprint/plugins/workflow"
	wf "github.com/blueprint-uservices/blueprint/test/workflow/workflow"

	"github.com/stretchr/testify/assert"
)

/*
Tests for correct IR layout from wiring spec helper functions for goproc

Primarily want visibility tests for nodes that are in separate processes but not addressible
*/

func TestServicesWithinSameProcess(t *testing.T) {
	spec := newWiringSpec("TestServicesWithinSameProcess")

	leaf := workflow.Service[*wf.TestLeafServiceImpl](spec, "leaf")
	nonleaf := workflow.Service[wf.TestNonLeafService](spec, "nonleaf", leaf)

	myproc := goproc.CreateProcess(spec, "myproc", leaf, nonleaf)

	app := assertBuildSuccess(t, spec, myproc)

	assertIR(t, app,
		`TestServicesWithinSameProcess = BlueprintApplication() {
			leaf.handler.visibility
			myproc = GolangProcessNode() {
			  leaf = TestLeafService()
			  leaf.client = leaf
			  myproc.logger = SLogger()
			  myproc.stdoutmetriccollector = StdoutMetricCollector()
			  nonleaf = TestNonLeafService(leaf.client)
			}
			nonleaf.handler.visibility
          }`)
}

func TestSeparateServicesInSeparateProcesses(t *testing.T) {
	spec := newWiringSpec("TestSeparateServicesInSeparateProcesses")

	leaf1 := workflow.Service[*wf.TestLeafServiceImpl](spec, "leaf1")
	leaf2 := workflow.Service[*wf.TestLeafServiceImpl](spec, "leaf2")
	nonleaf := workflow.Service[wf.TestNonLeafService](spec, "nonleaf", leaf2)

	leaf1proc := goproc.CreateProcess(spec, "leaf1proc", leaf1)
	myproc := goproc.CreateProcess(spec, "myproc", leaf2, nonleaf)

	app := assertBuildSuccess(t, spec, leaf1proc, myproc)

	assertIR(t, app,
		`TestSeparateServicesInSeparateProcesses = BlueprintApplication() {
            leaf1.handler.visibility
            leaf1proc = GolangProcessNode() {
              leaf1 = TestLeafService()
			  leaf1proc.logger = SLogger()
			  leaf1proc.stdoutmetriccollector = StdoutMetricCollector()
            }
            leaf2.handler.visibility
            myproc = GolangProcessNode() {
              leaf2 = TestLeafService()
			  leaf2.client = leaf2
			  myproc.logger = SLogger()
			  myproc.stdoutmetriccollector = StdoutMetricCollector()
              nonleaf = TestNonLeafService(leaf2.client)
            }
            nonleaf.handler.visibility
          }`)
}

func TestAddChildrenToProcess(t *testing.T) {
	spec := newWiringSpec("TestAddChildrenToProcess")

	myproc := goproc.CreateProcess(spec, "myproc")

	leaf := workflow.Service[*wf.TestLeafServiceImpl](spec, "leaf")
	nonleaf := workflow.Service[wf.TestNonLeafService](spec, "nonleaf", leaf)

	goproc.AddToProcess(spec, myproc, leaf)
	goproc.AddToProcess(spec, myproc, nonleaf)

	app := assertBuildSuccess(t, spec, myproc)

	assertIR(t, app,
		`TestAddChildrenToProcess = BlueprintApplication() {
            leaf.handler.visibility
            myproc = GolangProcessNode() {
              leaf = TestLeafService()
			  leaf.client = leaf
			  myproc.logger = SLogger()
			  myproc.stdoutmetriccollector = StdoutMetricCollector()
              nonleaf = TestNonLeafService(leaf.client)
            }
            nonleaf.handler.visibility
          }`)

}

func TestReachabilityErrorForSeparateProcesses(t *testing.T) {
	spec := newWiringSpec("TestReachabilityErrorForSeparateProcesses")

	leaf := workflow.Service[*wf.TestLeafServiceImpl](spec, "leaf")
	nonleaf := workflow.Service[wf.TestNonLeafService](spec, "nonleaf", leaf)

	leafproc := goproc.CreateProcess(spec, "leafproc", leaf)
	nonleafproc := goproc.CreateProcess(spec, "nonleafproc", nonleaf)

	err := assertBuildFailure(t, spec, leafproc, nonleafproc)
	assert.Contains(t, err.Error(), "reachability error")
}

func TestClientWithinSameProcess(t *testing.T) {
	spec := newWiringSpec("TestClientWithinSameProcess")

	leaf := workflow.Service[*wf.TestLeafServiceImpl](spec, "leaf")
	nonleaf := workflow.Service[wf.TestNonLeafService](spec, "nonleaf", leaf)

	nonleafclient := goproc.CreateClientProcess(spec, "nonleafclient", nonleaf)

	app := assertBuildSuccess(t, spec, nonleafclient)

	assertIR(t, app,
		`TestClientWithinSameProcess = BlueprintApplication() {
            leaf.handler.visibility
            nonleaf.handler.visibility
            nonleafclient = GolangProcessNode() {
              leaf = TestLeafService()
			  leaf.client = leaf
              nonleaf = TestNonLeafService(leaf.client)
			  nonleaf.client = nonleaf
            }
          }`)
}

func TestImplicitServicesWithinSameProcess(t *testing.T) {
	spec := newWiringSpec("TestImplicitServicesWithinSameProcess")

	leaf := workflow.Service[*wf.TestLeafServiceImpl](spec, "leaf")
	nonleaf := workflow.Service[wf.TestNonLeafService](spec, "nonleaf", leaf)

	nonleafproc := goproc.CreateProcess(spec, "nonleafproc", nonleaf)

	app := assertBuildSuccess(t, spec, nonleafproc)

	assertIR(t, app,
		`TestImplicitServicesWithinSameProcess = BlueprintApplication() {
            leaf.handler.visibility
            nonleaf.handler.visibility
            nonleafproc = GolangProcessNode() {
              leaf = TestLeafService()
			  leaf.client = leaf
              nonleaf = TestNonLeafService(leaf.client)
			  nonleafproc.logger = SLogger()
			  nonleafproc.stdoutmetriccollector = StdoutMetricCollector()
            }
          }`)
}

func TestProcessModifier(t *testing.T) {
	spec := newWiringSpec("TestProcessModifier")

	leaf := workflow.Service[*wf.TestLeafServiceImpl](spec, "leaf")
	nonleaf := workflow.Service[wf.TestNonLeafService](spec, "nonleaf", leaf)

	grpc.Deploy(spec, leaf)
	grpc.Deploy(spec, nonleaf)

	goproc.CreateProcess(spec, "leafproc", leaf)
	nonleafproc := goproc.CreateProcess(spec, "nonleafproc", nonleaf)

	app := assertBuildSuccess(t, spec, nonleafproc)

	assertIR(t, app,
		`TestProcessModifier = BlueprintApplication() {
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
