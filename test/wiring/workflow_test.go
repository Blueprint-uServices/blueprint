package wiring

import (
	"testing"

	"gitlab.mpi-sws.org/cld/blueprint/plugins/workflow"
	"golang.org/x/exp/slog"

	"github.com/stretchr/testify/assert"
)

/*
Tests for correct IR layout from wiring spec helper functions

The workflow services used in this test exercise the following:
 - use of various types in service methods;
	* basic types that are aliased
	* nested structs
	* pointers and values
	* named and implicit imports
 - constructors:
    * that return the interface type (only instantiable by specifying the interface name)
	* that return the implementation type (instantiable both by specifying the interface and by specifying the implementation)
*/

func TestBasicServices(t *testing.T) {
	spec := newWiringSpec("TestBasicServices")

	leaf := workflow.Define(spec, "leaf", "TestLeafServiceImpl")
	nonleaf := workflow.Define(spec, "nonleaf", "TestNonLeafService", leaf)

	app := assertBuildSuccess(t, spec, leaf, nonleaf)

	assertIR(t, app,
		`TestBasicServices = BlueprintApplication() {
			leaf.handler.visibility
			leaf = TestLeafService()
			nonleaf.handler.visibility
			nonleaf = TestNonLeafService(leaf)
		}`)
}

func TestImplicitInstantiation(t *testing.T) {
	spec := newWiringSpec("TestImplicitInstantiation")

	leaf := workflow.Define(spec, "leaf", "TestLeafServiceImpl")
	nonleaf := workflow.Define(spec, "nonleaf", "TestNonLeafService", leaf)

	app := assertBuildSuccess(t, spec, nonleaf)

	assertIR(t, app,
		`TestImplicitInstantiation = BlueprintApplication() {
			nonleaf.handler.visibility
			leaf.handler.visibility
			leaf = TestLeafService()
			nonleaf = TestNonLeafService(leaf)
		  }`)
}

func TestBadServiceConstructor(t *testing.T) {
	spec := newWiringSpec("TestBadServiceConstructor")

	leaf := workflow.Define(spec, "leaf", "TestLeafServiceImpl")
	nonleaf := workflow.Define(spec, "nonleaf", "TestNonLeafServiceImpl", leaf) // non-leaf service constructor returns the interface type; matching the impl not currently supported

	app, err := build(t, spec, leaf, nonleaf)
	if !assert.Error(t, err) {
		slog.Info("Wiring Spec: \n" + spec.String())
		slog.Info("Application: \n" + app.String())
	}
}

func TestBadServiceConstructor2(t *testing.T) {
	spec := newWiringSpec("TestBadServiceConstructor2")

	leaf := workflow.Define(spec, "leaf", "TestLeafService") // leaf service constructor returns an *impl; matching the interface not currently supported
	nonleaf := workflow.Define(spec, "nonleaf", "TestNonLeafService", leaf)

	app, err := build(t, spec, leaf, nonleaf)
	if !assert.Error(t, err) {
		slog.Info("Wiring Spec: \n" + spec.String())
		slog.Info("Application: \n" + app.String())
	}
}
