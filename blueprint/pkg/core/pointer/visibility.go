package pointer

import (
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/blueprint"
)

// Metadata used to enforce reachability constraints for nodes (primarily services)
type VisibilityMetadata struct {
	blueprint.IRMetadata

	name      string
	node      blueprint.IRNode
	namespace blueprint.Namespace
}

func (md *VisibilityMetadata) Name() string {
	return md.name
}

func (md *VisibilityMetadata) String() string {
	return md.name
}

/*
A uniqueness check can be applied to any aliased node.

It requires that the specified node must be unique up to a certain granularity.

This is independent of whether it can be addressed by any node within that granularity.

The name argument should be an alias that this call will redefine.
*/
func RequireUniqueness(wiring blueprint.WiringSpec, alias string, visibility any) {
	name, is_alias := wiring.GetAlias(alias)
	if !is_alias {
		wiring.AddError(blueprint.Errorf("cannot configure the uniqueness of %s because it points directly to a node; uniqueness can only be set for aliases", alias))
		return
	}

	def := wiring.GetDef(name)
	if def == nil {
		wiring.AddError(blueprint.Errorf("cannot configure the uniqueness of %s because it does not exist", name))
		return
	}

	mdName := name + ".visibility"
	wiring.Define(mdName, visibility, func(namespace blueprint.Namespace) (blueprint.IRNode, error) {
		md := &VisibilityMetadata{}
		md.name = mdName
		md.node = nil
		md.namespace = nil
		return md, nil
	})

	checkName := name + ".uniqueness_check"
	wiring.Define(checkName, def.NodeType, func(namespace blueprint.Namespace) (blueprint.IRNode, error) {
		var md *VisibilityMetadata
		if err := namespace.Get(mdName, &md); err != nil {
			return nil, blueprint.Errorf("expected %v to be uniqueness metadata but got %v", mdName, err)
		}

		if md.node != nil {
			return nil, blueprint.Errorf("reachability error detected for %s; %s is configured to be unique but cannot be simultaneously reached from namespaces %s and %s; fix by disabling uniqueness for %s or exposing %s over RPC", name, name, namespace.Name(), md.namespace.Name(), name, name)
		}

		md.namespace = namespace
		err := namespace.Get(name, &md.node)
		return md.node, err
	})

	wiring.Alias(alias, checkName)
}
