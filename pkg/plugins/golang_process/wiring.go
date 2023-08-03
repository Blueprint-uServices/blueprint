package golang_process

import (
	"gitlab.mpi-sws.org/cld/blueprint/pkg/blueprint"
)

// Adds a process that explicitly instantiates all of the children provided.
// The process will also implicitly instantiate any of the dependencies of the children
func Add(wiring *blueprint.WiringSpec, name string, children ...string) {
	wiring.Add(name, &GolangProcessNode{}, func(scope blueprint.Scope) (any, error) {
		process := NewGolangProcessScope(scope, wiring, name)

		// Get all of the child nodes.  If the child node hasn't actually been defined, then this will error out
		for _, child_name := range children {
			_, err := process.Get(child_name)
			if err != nil {
				return nil, err
			}
		}

		// Instantiate and return the service
		return process.Build()
	})
}
