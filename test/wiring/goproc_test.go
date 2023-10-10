package wiring

import (
	"testing"

	"gitlab.mpi-sws.org/cld/blueprint/plugins/goproc"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/workflow"

	"github.com/stretchr/testify/assert"
)

/*
Tests for correct IR layout from wiring spec helper functions for goproc

Primarily want visibility tests for nodes that are in separate processes but not addressible
*/

func TestServicesWithinSameProcess(t *testing.T) {
	wiring := newWiringSpec("TestServicesWithinSameProcess")

	leaf := workflow.Define(wiring, "leaf", "TestLeafServiceImpl")
	nonleaf := workflow.Define(wiring, "nonleaf", "TestNonLeafService", leaf)

	myproc := goproc.CreateProcess(wiring, "myproc", leaf, nonleaf)

	app := assertBuildSuccess(t, wiring, myproc)

	assertIR(t, app,
		`TestServicesWithinSameProcess = BlueprintApplication() {
            leaf.handler.visibility
            nonleaf.handler.visibility
            myproc = GolangProcessNode() {
              leaf = TestLeafService()
              nonleaf = TestNonLeafService(leaf)
            }
          }`)
}

func TestSeparateServicesInSeparateProcesses(t *testing.T) {
	wiring := newWiringSpec("TestSeparateServicesInSeparateProcesses")

	leaf1 := workflow.Define(wiring, "leaf1", "TestLeafServiceImpl")
	leaf2 := workflow.Define(wiring, "leaf2", "TestLeafServiceImpl")
	nonleaf := workflow.Define(wiring, "nonleaf", "TestNonLeafService", leaf2)

	leaf1proc := goproc.CreateProcess(wiring, "leaf1proc", leaf1)
	myproc := goproc.CreateProcess(wiring, "myproc", leaf2, nonleaf)

	app := assertBuildSuccess(t, wiring, leaf1proc, myproc)

	assertIR(t, app,
		`TestSeparateServicesInSeparateProcesses = BlueprintApplication() {
            leaf1.handler.visibility
            leaf1proc = GolangProcessNode() {
              leaf1 = TestLeafService()
            }
            leaf2.handler.visibility
            nonleaf.handler.visibility
            myproc = GolangProcessNode() {
              leaf2 = TestLeafService()
              nonleaf = TestNonLeafService(leaf2)
            }
          }`)
}

func TestAddChildrenToProcess(t *testing.T) {
	wiring := newWiringSpec("TestAddChildrenToProcess")

	myproc := goproc.CreateProcess(wiring, "myproc")

	leaf := workflow.Define(wiring, "leaf", "TestLeafServiceImpl")
	nonleaf := workflow.Define(wiring, "nonleaf", "TestNonLeafService", leaf)

	goproc.AddChildToProcess(wiring, myproc, leaf)
	goproc.AddChildToProcess(wiring, myproc, nonleaf)

	app := assertBuildSuccess(t, wiring, myproc)

	assertIR(t, app,
		`TestAddChildrenToProcess = BlueprintApplication() {
            leaf.handler.visibility
            nonleaf.handler.visibility
            myproc = GolangProcessNode() {
              leaf = TestLeafService()
              nonleaf = TestNonLeafService(leaf)
            }
          }`)

}

func TestReachabilityErrorForSeparateProcesses(t *testing.T) {
	wiring := newWiringSpec("TestReachabilityErrorForSeparateProcesses")

	leaf := workflow.Define(wiring, "leaf", "TestLeafServiceImpl")
	nonleaf := workflow.Define(wiring, "nonleaf", "TestNonLeafService", leaf)

	leafproc := goproc.CreateProcess(wiring, "leafproc", leaf)
	nonleafproc := goproc.CreateProcess(wiring, "nonleafproc", nonleaf)

	err := assertBuildFailure(t, wiring, leafproc, nonleafproc)
	assert.Contains(t, err.Error(), "reachability error")
}

func TestClientWithinSameProcess(t *testing.T) {
	wiring := newWiringSpec("TestClientWithinSameProcess")

	leaf := workflow.Define(wiring, "leaf", "TestLeafServiceImpl")
	nonleaf := workflow.Define(wiring, "nonleaf", "TestNonLeafService", leaf)

	nonleafclient := goproc.CreateClientProcess(wiring, "nonleafclient", nonleaf)

	app := assertBuildSuccess(t, wiring, nonleafclient)

	assertIR(t, app,
		`TestClientWithinSameProcess = BlueprintApplication() {
            nonleaf.handler.visibility
            leaf.handler.visibility
            nonleafclient = GolangProcessNode() {
              leaf = TestLeafService()
              nonleaf = TestNonLeafService(leaf)
            }
          }`)
}

func TestImplicitServicesWithinSameProcess(t *testing.T) {
	wiring := newWiringSpec("TestImplicitServicesWithinSameProcess")

	leaf := workflow.Define(wiring, "leaf", "TestLeafServiceImpl")
	nonleaf := workflow.Define(wiring, "nonleaf", "TestNonLeafService", leaf)

	nonleafproc := goproc.CreateProcess(wiring, "nonleafproc", nonleaf)

	app := assertBuildSuccess(t, wiring, nonleafproc)

	assertIR(t, app,
		`TestImplicitServicesWithinSameProcess = BlueprintApplication() {
            nonleaf.handler.visibility
            leaf.handler.visibility
            nonleafproc = GolangProcessNode() {
              leaf = TestLeafService()
              nonleaf = TestNonLeafService(leaf)
            }
          }`)
}
