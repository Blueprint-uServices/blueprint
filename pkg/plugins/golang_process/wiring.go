package golang_process

import (
	"gitlab.mpi-sws.org/cld/blueprint/pkg/blueprint"
)

// Adds a service of type serviceType to the wiring spec, giving it the name specified.
// Services can have arguments which are other named nodes
func Add(wiring *blueprint.WiringSpec, name string, args ...string) {
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
		service, err := newGolangProcessNode(name, arg_nodes)
		return "golang process", service, err
	})
}
