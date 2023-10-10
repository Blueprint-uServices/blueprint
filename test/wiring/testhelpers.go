package wiring

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/blueprint"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/workflow"
	"golang.org/x/exp/slog"
)

func newWiringSpec(name string) blueprint.WiringSpec {
	blueprint.DisableCompilerLogging()
	defer blueprint.EnableCompilerLogging()
	workflow.Reset()
	wiring := blueprint.NewWiringSpec(name)
	workflow.Init("../workflow")
	return wiring
}

func build(t *testing.T, wiring blueprint.WiringSpec, toInstantiate ...string) (*blueprint.ApplicationNode, error) {
	blueprint.DisableCompilerLogging()
	defer blueprint.EnableCompilerLogging()

	bp, err := wiring.GetBlueprint()
	if err != nil {
		return nil, err
	}

	bp.Instantiate(toInstantiate...)

	return bp.Build()
}

func assertBuildFailure(t *testing.T, wiring blueprint.WiringSpec, toInstantiate ...string) error {
	app, err := build(t, wiring, toInstantiate...)

	if !assert.Error(t, err) {
		slog.Info("Expected a build error, but did not get one")
		slog.Info("Wiring Spec: \n" + wiring.String())
		slog.Info("Application: \n" + app.String())
	}
	return err
}

func assertBuildSuccess(t *testing.T, wiring blueprint.WiringSpec, toInstantiate ...string) *blueprint.ApplicationNode {
	app, err := build(t, wiring, toInstantiate...)

	if !assert.NoError(t, err) {
		slog.Info("Unexpected error building application")
		slog.Info("Wiring Spec: \n" + wiring.String())
		slog.Info("Application: \n" + app.String())
		slog.Error(err.Error())
	}
	return app
}

func splits(str string) []string {
	ss := strings.Split(str, "\n")
	for i := range ss {
		ss[i] = strings.TrimSpace(ss[i])
	}
	return ss
}

/*
The easiest / most convenient way of checking an IR matches expectations.
This is brittle; there are plenty of reasons why this is a terrible idea.
But it's also far less time consuming and convenient for now than constructing
and testing IR objects
*/
func assertIR(t *testing.T, app *blueprint.ApplicationNode, expected string) bool {
	a := splits(app.String())
	b := splits(expected)
	if !assert.Equal(t, b, a) {
		slog.Info("Application: \n" + app.String())
		return false
	}
	return true
}
