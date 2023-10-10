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
	wiring := newWiringSpec("TestBasicServices")

	leaf := workflow.Define(wiring, "leaf", "TestLeafServiceImpl")
	nonleaf := workflow.Define(wiring, "nonleaf", "TestNonLeafService", leaf)

	app, err := build(t, wiring, leaf, nonleaf)
	if !assert.NoError(t, err) {
		slog.Info("Wiring Spec: \n" + wiring.String())
		slog.Info("Application: \n" + app.String())
	}

	assertIR(t, app,
		`TestBasicServices = BlueprintApplication() {
			leaf.handler.visibility
			leaf = TestLeafService()
			nonleaf.handler.visibility
			nonleaf = TestNonLeafService(leaf)
		}`)
}

func TestBadServiceConstructor(t *testing.T) {
	wiring := newWiringSpec("TestBadServiceConstructor")

	leaf := workflow.Define(wiring, "leaf", "TestLeafServiceImpl")
	nonleaf := workflow.Define(wiring, "nonleaf", "TestNonLeafServiceImpl", leaf) // non-leaf service constructor returns the interface type; matching the impl not currently supported

	app, err := build(t, wiring, leaf, nonleaf)
	if !assert.Error(t, err) {
		slog.Info("Wiring Spec: \n" + wiring.String())
		slog.Info("Application: \n" + app.String())
	}
}

func TestBadServiceConstructor2(t *testing.T) {
	wiring := newWiringSpec("TestBadServiceConstructor2")

	leaf := workflow.Define(wiring, "leaf", "TestLeafService") // leaf service constructor returns an *impl; matching the interface not currently supported
	nonleaf := workflow.Define(wiring, "nonleaf", "TestNonLeafService", leaf)

	app, err := build(t, wiring, leaf, nonleaf)
	if !assert.Error(t, err) {
		slog.Info("Wiring Spec: \n" + wiring.String())
		slog.Info("Application: \n" + app.String())
	}
}
