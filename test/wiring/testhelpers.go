package wiring

import (
	"flag"
	"strings"
	"testing"

	"github.com/Blueprint-uServices/blueprint/blueprint/pkg/blueprint/logging"
	"github.com/Blueprint-uServices/blueprint/blueprint/pkg/ir"
	"github.com/Blueprint-uServices/blueprint/blueprint/pkg/wiring"
	"github.com/Blueprint-uServices/blueprint/plugins/workflow"
	"github.com/stretchr/testify/require"
)

var compilerLogging = flag.Bool("blueprint_verbose", false, "Enables verbose Blueprint compiler logging")

func newWiringSpec(name string) wiring.WiringSpec {
	if !*compilerLogging {
		logging.DisableCompilerLogging()
		defer logging.EnableCompilerLogging()
	}
	workflow.Reset()
	spec := wiring.NewWiringSpec(name)
	workflow.Init("../workflow")
	return spec
}

func build(t *testing.T, spec wiring.WiringSpec, toInstantiate ...string) (*ir.ApplicationNode, error) {
	if !*compilerLogging {
		logging.DisableCompilerLogging()
		defer logging.EnableCompilerLogging()
	}
	return spec.BuildIR(toInstantiate...)
}

func assertBuildFailure(t *testing.T, spec wiring.WiringSpec, toInstantiate ...string) error {
	app, err := build(t, spec, toInstantiate...)
	require.Error(t, err, "Expected a build error but did not get one.\nWiring Spec: %v\nApplication: %v", spec.String(), app.String())
	return err
}

func assertBuildSuccess(t *testing.T, spec wiring.WiringSpec, toInstantiate ...string) *ir.ApplicationNode {
	app, err := build(t, spec, toInstantiate...)
	require.NoError(t, err, "Unepxected error building application.\nWiring Spec: %v\nApplication: %v", spec.String(), app.String())
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
	require.Equal(t, b, a, "Got unexpected application\n%v", app.String())
	return true
}
