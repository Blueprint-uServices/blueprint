// Package gotests provides a plugin for automatically converting workflow spec unit tests into
// tests that can run against a compiled Blueprint system.
package gotests

import (
	"strings"

	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/blueprint"
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/ir"
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/wiring"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/golang"
)

// Specifies that any compatible unit dests defined for serviceName should be automatically
// converted into tests for the compiled Blueprint system.
//
// The gotests plugin will produce an output golang workspace called "tests" that will
// include modified versions of the source unit tests.  After running the compiled Blueprint
// system, you can run the tests with the usual 'go test'.
//
// Tests must meet the following criteria to be included:
//   - tests are written using go's standard testing library
//   - tests are contained in a separate module from the workflow spec
//     (and thus are black-box tests), needed to prevent circular module
//     dependencies
//   - tests make use of the core runtime registry.ServiceRegistry to acquire
//     client instances (as opposed to manually constructing them).
func Test(spec wiring.WiringSpec, servicesToTest ...string) string {

	name := "gotests"

	// The output gotests package can include tests for multiple services
	for _, serviceName := range servicesToTest {
		spec.AddProperty(name, "Services", serviceName)
	}

	// Might redefine gotests multiple times; no big deal
	spec.Define("gotests", &TestLibrary{}, func(namespace wiring.Namespace) (ir.IRNode, error) {
		testLib := newTestsNamespace(namespace, spec, name)

		var serviceNames []string
		if err := namespace.GetProperties(name, "Services", &serviceNames); err != nil {
			return nil, blueprint.Errorf("unable to build Golang process as the \"Services\" property is not defined: %s", err.Error())
		}
		testLib.Info("converting unit tests for %v services (%s)", len(serviceNames), strings.Join(serviceNames, ", "))

		// Instantiate service clients.  If the child node hasn't actually been defined, then this will error out
		for _, serviceName := range serviceNames {
			var client ir.IRNode
			if err := testLib.Get(serviceName, &client); err != nil {
				return nil, err
			}
			testLib.handler.IRNode.ServicesToTest[serviceName] = client
		}

		// Instantiate and return the service
		return testLib.handler.IRNode, nil
	})

	return name
}

type testsNamespace struct {
	wiring.SimpleNamespace
	handler *testsNamespaceHandler
}

type testsNamespaceHandler struct {
	wiring.DefaultNamespaceHandler
	IRNode *TestLibrary
}

func newTestsNamespace(parent wiring.Namespace, spec wiring.WiringSpec, name string) *testsNamespace {
	namespace := &testsNamespace{}
	namespace.handler = &testsNamespaceHandler{}
	namespace.handler.Init(&namespace.SimpleNamespace)
	namespace.handler.IRNode = newTestLibrary(name)
	namespace.Init(name, "GolangTests", parent, spec, namespace.handler)
	return namespace
}

func (handler *testsNamespaceHandler) Accepts(nodeType any) bool {
	_, ok := nodeType.(golang.Node)
	return ok
}

func (handler *testsNamespaceHandler) AddNode(name string, node ir.IRNode) error {
	handler.IRNode.ContainedNodes = append(handler.IRNode.ContainedNodes, node)
	return nil
}

func (handler *testsNamespaceHandler) AddEdge(name string, node ir.IRNode) error {
	handler.IRNode.ArgNodes = append(handler.IRNode.ArgNodes, node)
	return nil
}
