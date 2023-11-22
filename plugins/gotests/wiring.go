// Package gotests provides a Blueprint plugin for automatically converting black-box workflow spec unit tests into
// tests that can run against a compiled Blueprint system.
//
// To use the gotests plugin in a wiring spec, call gotests.[Test] and specify the services to test.  During
// compilation the plugin will search for compatible black-box tests, modify them, and include the modified tests
// in the compiled output.
//
// Tests are only compatible if they meet the following requirements:
//   - tests are contained in a separate module from the workflow spec
//   - the test module is on the workflow spec search path (e.g. using [workflow.Init])
//   - tests use the [registry.ServiceRegistry] to acquire service clients
//
// To run the generated tests:
//  1. start the compiled application
//  2. navigate to the 'tests' directory in the compiled output
//  3. run `go test` passing any required command-line arguments such as addresses
//
// A test can potentially leave state within the application; running the test twice in succession
// may result in a failure the second time since the application is not in a clean state.
//
// See [Workflow Tests] for more information on writing workflow tests.
//
// [Workflow Tests]: https://github.com/Blueprint-uServices/blueprint/tree/main/docs/manual/workflow_tests.md
// [registry.ServiceRegistry]: https://github.com/Blueprint-uServices/blueprint/tree/main/runtime/core/registry
// [workflow.Init]: https://github.com/Blueprint-uServices/blueprint/tree/main/plugins/workflow
package gotests

import (
	"strings"

	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/blueprint"
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/ir"
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/wiring"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/golang"
)

// Auto-generates tests for servicesToTest by converting existing black-box workflow unit tests.
// After compilation, the output will contain a golang workspace called "tests" that will
// include modified versions of the source tests.
//
// servicesToTest should be the names of golang services instantiated in the wiring spec.
//
// The gotests plugin searches for any workflow packages with tests that make use of [registry.ServiceRegistry].
// Matching modules are copied to an output golang workspace caled "tests".
// Matching packges in the output workspace will have a file blueprint_clients.go that registers
// a service client.
//
// Returns the name "gotests" which must be included when later calling [wiring.WiringSpec.BuildIR]
//
// For more information about tests see [Workflow Tests].
//
// [Workflow Tests]: https://github.com/Blueprint-uServices/blueprint/tree/main/docs/manual/workflow_tests.md
// [registry.ServiceRegistry]: https://github.com/Blueprint-uServices/blueprint/tree/main/runtime/core/registry
func Test(spec wiring.WiringSpec, servicesToTest ...string) string {

	name := "gotests"

	// The output gotests package can include tests for multiple services
	for _, serviceName := range servicesToTest {
		spec.AddProperty(name, "Services", serviceName)
	}

	// Might redefine gotests multiple times; no big deal
	spec.Define("gotests", &testLibrary{}, func(namespace wiring.Namespace) (ir.IRNode, error) {
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
	IRNode *testLibrary
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
