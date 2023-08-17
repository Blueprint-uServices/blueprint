package pointer

import (
	"fmt"
	"reflect"

	"gitlab.mpi-sws.org/cld/blueprint/pkg/blueprint"
)

type pointerInfo struct {
	name         string
	pointsTo     string
	visibility   any
	reachability any
}

// Metadata used to check for reachability of addresses
type VisibilityMetadata struct {
	blueprint.IRMetadata

	info *pointerInfo
	addr *Address
}

// Represents an address to a node that might reside in a different scope
type Address struct {
	blueprint.IRNode

	info        *pointerInfo
	scope       blueprint.Scope // the scope within which the address is reachable
	callers     []*PointerCallsite
	destination blueprint.IRNode
}

// Metadata used to figure out where to begin instantiating the server-side of a pointer
type PointerCallsite struct {
	blueprint.IRMetadata

	info  *pointerInfo
	scope blueprint.Scope // the scope of the callsite
}

/*
Defining a pointer will actually produce four definitions:
  - a definition for the name itself
  - an address definition that will exist in the 'reachability' scope
  - a visibility metadata definition that will exist in the 'visibility' scope
*/
func DefinePointer(wiring blueprint.WiringSpec, name string, pointsTo string, visibility any, reachability any) error {
	addressName := name + ".addr"
	metadataName := name + ".visibility"

	info := &pointerInfo{}
	info.name = name
	info.pointsTo = pointsTo
	info.visibility = visibility
	info.reachability = reachability

	def := wiring.GetDef(pointsTo)
	if def == nil {
		return fmt.Errorf("trying to define pointer %s that points to node %s but the node is not defined", name, pointsTo)
	}

	wiring.Define(name, def.NodeType, func(scope blueprint.Scope) (blueprint.IRNode, error) {
		addrNode, err := scope.Get(addressName)
		if err != nil {
			return nil, err
		}

		addr, is_valid := addrNode.(*Address)
		if !is_valid {
			return nil, fmt.Errorf("pointer %s -> %s should have an address node under the name %s, but it is of unexpected type %s", name, pointsTo, addressName, reflect.TypeOf(addrNode).Elem().Name())
		}

		callsite := &PointerCallsite{}
		callsite.info = info
		callsite.scope = scope

		// Save the callsite info on the address for when we come to instantiate the destination-side of the pointer
		addr.callers = append(addr.callers, callsite)

		// If we are the first caller, schedule a deferred function that will instantiate
		// the destination of the pointer (if it hasn't already been instantiated by that point)
		if len(addr.callers) == 1 {
			scope.Info("Deferring instantiation of %v", pointsTo)
			scope.Defer(func() error {
				if addr.destination != nil {
					// it has been instantiated already
					return nil
				}
				scope.Info("Instantiating %v", pointsTo)

				// Create the destination node using the scope of the callsite
				_, err := scope.Get(pointsTo)
				return err
			})
		}

		return addrNode, nil
	})

	wiring.Define(addressName, reachability, func(scope blueprint.Scope) (blueprint.IRNode, error) {
		visibilityNode, err := scope.Get(metadataName)
		if err != nil {
			return nil, err
		}

		metadata, is_valid := visibilityNode.(*VisibilityMetadata)
		if !is_valid {
			return nil, fmt.Errorf("pointer %s -> %s should have reachability metadata under the name %s, but it is of unexpected type %s", name, pointsTo, metadataName, reflect.TypeOf(visibilityNode).Elem().Name())
		}

		addr := &Address{}
		addr.info = info
		addr.scope = scope
		addr.callers = nil
		addr.destination = nil

		// Save the addr on the visibility metadata for performing visibility checks
		if metadata.addr != nil {
			return nil, fmt.Errorf("reachability error while building %s for pointer %s (-> %s); %s cannot be simultaneously reached from scopes %s and %s", addressName, name, pointsTo, pointsTo, scope.Name(), metadata.addr.scope.Name())
		}
		metadata.addr = addr

		return addr, nil
	})

	wiring.Define(metadataName, visibility, func(scope blueprint.Scope) (blueprint.IRNode, error) {
		metadata := &VisibilityMetadata{}
		metadata.info = info
		metadata.addr = nil
		return metadata, nil
	})

	return nil
}

func (md *VisibilityMetadata) Name() string {
	return md.info.name + ".visibility"
}

func (md *VisibilityMetadata) String() string {
	return md.Name()
}

func (md *Address) Name() string {
	return md.info.name + ".addr"
}

func (md *Address) String() string {
	return md.Name()
}

func (md *PointerCallsite) Name() string {
	return md.info.name + ".callsite"
}

func (md *PointerCallsite) String() string {
	return md.Name()
}
