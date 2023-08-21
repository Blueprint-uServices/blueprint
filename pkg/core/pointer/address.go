package pointer

import (
	"fmt"

	"gitlab.mpi-sws.org/cld/blueprint/pkg/blueprint"
)

// Represents an address to a node that might reside in a different scope
type Address struct {
	blueprint.IRNode

	name        string
	Destination string
	DstNode     blueprint.IRNode
}

func DefineAddress(wiring blueprint.WiringSpec, addressName string, pointsTo string, reachability any) error {
	def := wiring.GetDef(pointsTo)
	if def == nil {
		return fmt.Errorf("trying to define address %s that points to %s but %s is not defined", addressName, pointsTo, pointsTo)
	}

	wiring.Define(addressName, reachability, func(scope blueprint.Scope) (blueprint.IRNode, error) {
		addr := &Address{}
		addr.name = addressName
		addr.Destination = pointsTo
		addr.DstNode = nil
		return addr, nil
	})

	return nil
}

func (md *Address) Name() string {
	return md.name
}

func (md *Address) String() string {
	return md.name + " = Address(-> " + md.Destination + ")"
}

func (ad *Address) ImplementsAddress() {}
