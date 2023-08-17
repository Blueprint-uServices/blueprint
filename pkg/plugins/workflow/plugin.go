package workflow

import "gitlab.mpi-sws.org/cld/blueprint/pkg/plugins/golang"

var workflowSpecPaths []string

// Golang workflow must be initialized with a path to the workflow code
func Init(path string) {
	workflowSpecPaths = append(workflowSpecPaths, path)
}

// Finds the service with the specified type in the workflow spec.
// This method searches the WorkflowSpecPath and returns an error if not found.
func findService(serviceType string) (*golang.GolangServiceDetails, error) {
	// TODO: this searches the WorkflowSpecPath for the service of the requested type,
	//       and either returns its details or an error
	// return nil, fmt.Errorf("could not find service \"%s\" in the workflow spec", serviceType)

	mockup := golang.GolangServiceDetails{}
	mockup.Interface.Name = serviceType
	mockup.Package = "my.workflow.package"
	mockup.Files = []string{workflowSpecPaths[0] + "path/to/my/service"}

	return &mockup, nil
}
