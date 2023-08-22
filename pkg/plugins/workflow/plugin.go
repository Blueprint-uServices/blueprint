package workflow

import (
	"fmt"
	"strings"

	"gitlab.mpi-sws.org/cld/blueprint/pkg/plugins/golang"
	"gitlab.mpi-sws.org/cld/blueprint/pkg/plugins/workflow/parser"
)

var workflowSpecPaths []string
var spec *parser.SpecParser

// Golang workflow must be initialized with a path to the workflow code
func Init(path string) {
	workflowSpecPaths = append(workflowSpecPaths, path)
	spec = nil
}

func getSpec() (*parser.SpecParser, error) {
	if spec == nil {
		spec = parser.NewSpecParser(workflowSpecPaths...)
		err := spec.ParseSpec()
		if err != nil {
			spec = nil
			return nil, err
		}
	}
	return spec, nil
}

// Finds the service with the specified type in the workflow spec.
// This method searches the WorkflowSpecPath and returns an error if not found.
func findService(serviceType string) (*golang.GolangServiceDetails, error) {
	spec, err := getSpec()
	if err != nil {
		return nil, err
	}

	if impl, exists := spec.Implementations[serviceType]; exists {
		mockup := golang.GolangServiceDetails{}
		mockup.Interface.Name = serviceType
		mockup.Package = impl.PkgPath
		mockup.Files = []string{impl.PkgPath}
		return &mockup, nil
	} else {
		return nil, fmt.Errorf("unable to find workflow spec service %s in any of the following locations: %s", serviceType, strings.Join(workflowSpecPaths, "; "))
	}
}
