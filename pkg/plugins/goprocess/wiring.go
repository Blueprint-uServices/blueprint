package workflow

import (
	"os"

	"gitlab.mpi-sws.org/cld/blueprint/pkg/blueprint"
	"golang.org/x/exp/slog"
)

var WorkflowSpecPath string

// Set the path to inspect when looking for golang workflow spec services
func SetWorkflowSpecPath(path string) {
	WorkflowSpecPath = path
}

// Finds the service with the specified type in the workflow spec.
// This method searches the WorkflowSpecPath and returns an error if not found.
func findService(serviceType string) (*GolangServiceDetails, error) {
	// TODO: this searches the WorkflowSpecPath for the service of the requested type,
	//       and either returns its details or an error
	// return nil, fmt.Errorf("could not find service \"%s\" in the workflow spec", serviceType)

	mockup := GolangServiceDetails{}
	mockup.Interface.Name = serviceType
	mockup.Package = "my.workflow.package"
	mockup.Files = []string{WorkflowSpecPath + "path/to/my/service"}

	return &mockup, nil
}

// Adds a service of type serviceType to the wiring spec, giving it the name specified.
// Services can have arguments which are other named nodes
func Add(wiring *blueprint.WiringSpec, name, serviceType string, args ...string) {
	// Eagerly look up the service in the workflow spec to make sure it exists
	details, err := findService(serviceType)
	if err != nil {
		slog.Error("Unable to resolve workflow spec services used by the wiring spec, exiting", "error", err)
		os.Exit(1)
	}

	wiring.Add(name, func(bp *blueprint.Blueprint) (string, interface{}, error) {
		// Get all of the argument nodes; can error out if the arguments weren't actually defined
		var arg_nodes []blueprint.IRNode
		for _, arg_name := range args {
			node, err := bp.Get(arg_name)
			if err != nil {
				return "", nil, err
			}
			arg_nodes = append(arg_nodes, node)
		}

		// Instantiate and return the service
		service := newGolangWorkflowSpecServiceNode(name, details, arg_nodes)
		return "golang instance", service, err
	})
}
