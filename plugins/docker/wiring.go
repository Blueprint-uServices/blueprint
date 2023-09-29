package docker

import "gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/blueprint"

func AddChildToContainer(wiring blueprint.WiringSpec, ctrName string, childName string) {
	wiring.AddProperty(ctrName, "Children", childName)
}

// Adds a container that explicitly instantiates all of the children provided.
// The container will also implicitly instantiate any of the dependencies of the children
func CreateContainer(wiring blueprint.WiringSpec, ctrName string, children ...string) string {
	for _, childName := range children {
		AddChildToContainer(wiring, ctrName, childName)
	}

	// TODO: Implement

	return ctrName
}
