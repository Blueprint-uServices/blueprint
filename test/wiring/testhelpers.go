package wiring

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/blueprint/logging"
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/ir"
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/wiring"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/workflow"
	"golang.org/x/exp/slog"
)

func newWiringSpec(name string) wiring.WiringSpec {
	logging.DisableCompilerLogging()
	defer logging.EnableCompilerLogging()
	workflow.Reset()
	spec := wiring.NewWiringSpec(name)
	workflow.Init("../workflow")
	return spec
}

func build(t *testing.T, spec wiring.WiringSpec, toInstantiate ...string) (*ir.ApplicationNode, error) {
	logging.DisableCompilerLogging()
	defer logging.EnableCompilerLogging()
	return spec.BuildIR(toInstantiate...)
}

func assertBuildFailure(t *testing.T, spec wiring.WiringSpec, toInstantiate ...string) error {
	app, err := build(t, spec, toInstantiate...)

	if !assert.Error(t, err) {
		slog.Info("Expected a build error, but did not get one")
		slog.Info("Wiring Spec: \n" + spec.String())
		slog.Info("Application: \n" + app.String())
	}
	return err
}

func assertBuildSuccess(t *testing.T, spec wiring.WiringSpec, toInstantiate ...string) *ir.ApplicationNode {
	app, err := build(t, spec, toInstantiate...)

	if !assert.NoError(t, err) {
		slog.Info("Unexpected error building application")
		slog.Info("Wiring Spec: \n" + spec.String())
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
func assertIR(t *testing.T, app *ir.ApplicationNode, expected string) bool {
	a := splits(app.String())
	b := splits(expected)
	if !assert.Equal(t, b, a) {
		slog.Info("Application: \n" + app.String())
		return false
	}
	return true
}
