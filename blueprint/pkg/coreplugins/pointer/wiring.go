package pointer

import (
	"strings"

	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/blueprint"
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/ir"
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/wiring"
)

// A helper function for use by Blueprint plugins.  Instantiates the server-side
// nodes of the specified pointer(s) within the provided namespace.
//
// Returns a map of the instantiated node(s).
func Instantiate(spec wiring.WiringSpec, namespace wiring.Namespace, names ...string) (nodes map[string]ir.IRNode, err error) {
	nodes = make(map[string]ir.IRNode)
	for _, childName := range names {
		var child ir.IRNode
		ptr := GetPointer(spec, childName)
		if ptr == nil {
			err = namespace.Get(childName, &child)
		} else {
			child, err = ptr.InstantiateDst(namespace)
		}
		if err != nil {
			return
		}
		nodes[childName] = child
	}
	return
}

// Similar to Instantiate, but first consulting the propertyName property of the namespace
// to discover which nodes should be instantiated.
func InstantiateFromProperty(spec wiring.WiringSpec, namespace wiring.Namespace, propertyName string) (map[string]ir.IRNode, error) {
	var nodeNames []string
	if err := namespace.GetProperties(namespace.Name(), propertyName, &nodeNames); err != nil {
		return nil, blueprint.Errorf("%v InstantiateFromProperty %v failed due to %s", namespace.Name(), propertyName, err.Error())
	}
	namespace.Info("%v = %s", propertyName, strings.Join(nodeNames, ", "))
	return Instantiate(spec, namespace, nodeNames...)
}

// This effectively just calls namespace.Get() for the names provided. Included here
// for convenience
func InstantiateClients(spec wiring.WiringSpec, namespace wiring.Namespace, names ...string) (map[string]ir.IRNode, error) {
	nodes := make(map[string]ir.IRNode)
	for _, childName := range names {
		var child ir.IRNode
		if err := namespace.Get(childName, &child); err != nil {
			return nodes, err
		}
		nodes[childName] = child
	}
	return nodes, nil
}

// Similar to InstantiateClients, but first consulting the propertyName property of the namespace
// to discover which nodes should be instantiated.
func InstantiateClientsFromProperty(spec wiring.WiringSpec, namespace wiring.Namespace, propertyName string) (map[string]ir.IRNode, error) {
	var nodeNames []string
	if err := namespace.GetProperties(namespace.Name(), propertyName, &nodeNames); err != nil {
		return nil, blueprint.Errorf("%v InstantiateClientsFromProperty %v failed due to %s", namespace.Name(), propertyName, err.Error())
	}
	namespace.Info("%v = %s", propertyName, strings.Join(nodeNames, ", "))
	return InstantiateClients(spec, namespace, nodeNames...)
}
