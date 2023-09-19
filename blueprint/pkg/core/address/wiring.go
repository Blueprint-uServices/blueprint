package address

import (
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/blueprint"
)

/*
Defines an address called `addressName` that points to the definition `pointsto`.

The provided buildFunc should build an IRNode that implements the address.Address interface
*/
func Define(wiring blueprint.WiringSpec, addressName string, pointsTo string, reachability any, build func(scope blueprint.Scope) (Address, error)) error {
	def := wiring.GetDef(pointsTo)
	if def == nil {
		return blueprint.Errorf("trying to define address %s that points to %s but %s is not defined", addressName, pointsTo, pointsTo)
	}

	wiring.Define(addressName, reachability, func(scope blueprint.Scope) (blueprint.IRNode, error) {
		return build(scope)
	})
	wiring.SetProperty(addressName, "pointsTo", pointsTo)

	return nil
}

func DestinationOf(scope blueprint.Scope, addressName string) (string, error) {
	prop, err := scope.GetProperty(addressName, "pointsTo")
	if err != nil {
		return "", err
	}
	pointsTo, isString := prop.(string)
	if !isString {
		return "", blueprint.Errorf("expected the pointsTo property of %v to be a string but got %v", addressName, prop)
	}
	return pointsTo, nil
}
