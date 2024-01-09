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
// [Workflow Tests]: https://github.com/blueprint-uservices/blueprint/tree/main/docs/manual/workflow_tests.md
// [registry.ServiceRegistry]: https://github.com/blueprint-uservices/blueprint/tree/main/runtime/core/registry
// [workflow.Init]: https://github.com/blueprint-uservices/blueprint/tree/main/plugins/workflow
package gotests

import (
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/coreplugins/namespaceutil"
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/ir"
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/wiring"
	"github.com/blueprint-uservices/blueprint/plugins/golang"
)

var prop_SERVICESTOTEST = "Services"

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
// [Workflow Tests]: https://github.com/blueprint-uservices/blueprint/tree/main/docs/manual/workflow_tests.md
// [registry.ServiceRegistry]: https://github.com/blueprint-uservices/blueprint/tree/main/runtime/core/registry
func Test(spec wiring.WiringSpec, servicesToTest ...string) string {

	name := "gotests"

	for _, serviceName := range servicesToTest {
		spec.AddProperty(name, prop_SERVICESTOTEST, serviceName)
	}

	spec.Define(name, &testLibrary{}, func(namespace wiring.Namespace) (ir.IRNode, error) {
		lib := newTestLibrary(name)
		libNamespace, err := namespaceutil.InstantiateNamespace(namespace, &gotests{lib})
		if err != nil {
			return nil, err
		}

		var servicesToTest []string
		if err := namespace.GetProperties(name, prop_SERVICESTOTEST, &servicesToTest); err != nil {
			return nil, err
		}

		for _, serviceName := range servicesToTest {
			var service ir.IRNode
			if err := libNamespace.Get(serviceName, &service); err != nil {
				return nil, err
			}
			lib.ServicesToTest[serviceName] = service
		}
		return lib, err
	})

	return name
}

// A [wiring.NamespaceHandler] used to build the test library
type gotests struct {
	*testLibrary
}

// Implements [wiring.NamespaceHandler]
func (*gotests) Accepts(nodeType any) bool {
	_, isGolangNode := nodeType.(golang.Node)
	return isGolangNode
}

// Implements [wiring.NamespaceHandler]
func (lib *gotests) AddEdge(name string, edge ir.IRNode) error {
	lib.Edges = append(lib.Edges, edge)
	return nil
}

// Implements [wiring.NamespaceHandler]
func (lib *gotests) AddNode(name string, node ir.IRNode) error {
	lib.Nodes = append(lib.Nodes, node)
	return nil
}
