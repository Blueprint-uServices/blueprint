package golang_process

import (
	"fmt"

	"gitlab.mpi-sws.org/cld/blueprint/pkg/blueprint"
)

// Adds a child node to an existing process
func Instantiate(wiring *blueprint.WiringSpec, procName, childName string) {
	wiring.AddProperty(procName, "children", childName)
}

// Adds a process that explicitly instantiates all of the children provided.
// The process will also implicitly instantiate any of the dependencies of the children
func Define(wiring *blueprint.WiringSpec, procName string, children ...string) {
	// If any children were provided in this call, add them to the process via a property
	for _, childName := range children {
		Instantiate(wiring, procName, childName)
	}

	wiring.Define(procName, &GolangProcessNode{}, func(scope blueprint.Scope) (any, error) {
		process := NewGolangProcessScope(scope, wiring, procName)

		childNames, err := scope.GetProperty(procName, "children")
		if err != nil {
			return nil, fmt.Errorf("unable to build Golang process as the \"children\" property is not defined: %s", err.Error())
		}

		// Get all of the child nodes.  If the child node hasn't actually been defined, then this will error out
		for _, childName := range childNames {
			_, err := process.Get(childName.(string))
			if err != nil {
				return nil, err
			}
		}

		// Instantiate and return the service
		return process.Build()
	})
}
