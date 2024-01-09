package pointer

import (
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/blueprint"
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/ir"
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/wiring"
)

// Metadata used to enforce reachability constraints for nodes (primarily services)
type visibilityMetadata struct {
	ir.IRMetadata

	name      string
	node      ir.IRNode
	namespace wiring.Namespace
}

func (md *visibilityMetadata) Name() string {
	return md.name
}

func (md *visibilityMetadata) String() string {
	return md.name
}

/*
A uniqueness check can be applied to any aliased node.

It requires that the specified node must be unique up to a certain granularity.

This is independent of whether it can be addressed by any node within that granularity.

The name argument should be an alias that this call will redefine.
*/
func RequireUniqueness(spec wiring.WiringSpec, alias string, visibility any) {
	name, is_alias := spec.GetAlias(alias)
	if !is_alias {
		spec.AddError(blueprint.Errorf("cannot configure the uniqueness of %s because it points directly to a node; uniqueness can only be set for aliases", alias))
		return
	}

	def := spec.GetDef(name)
	if def == nil {
		spec.AddError(blueprint.Errorf("cannot configure the uniqueness of %s because it does not exist", name))
		return
	}

	mdName := name + ".visibility"
	spec.Define(mdName, visibility, func(namespace wiring.Namespace) (ir.IRNode, error) {
		md := &visibilityMetadata{}
		md.name = mdName
		md.node = nil
		md.namespace = nil
		return md, nil
	})

	checkName := name + ".uniqueness_check"
	spec.Define(checkName, def.NodeType, func(namespace wiring.Namespace) (ir.IRNode, error) {
		var md *visibilityMetadata
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

	spec.Alias(alias, checkName)
}
