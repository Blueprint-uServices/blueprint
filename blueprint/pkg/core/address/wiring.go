package address

import (
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/blueprint"
)

/*
Defines an address called `addressName` that points to the definition `pointsto`.

The provided buildFunc should build an IRNode that implements the address.Address interface
*/
func Define[ServerType blueprint.IRNode](wiring blueprint.WiringSpec, addressName string, pointsTo string, reachability any) {
	def := wiring.GetDef(pointsTo)
	if def == nil {
		wiring.AddError(blueprint.Errorf("trying to define address %s that points to %s but %s is not defined", addressName, pointsTo, pointsTo))
	}

	// Define the metadata for the address, used during the build process
	wiring.Define(addressName, reachability, func(namespace blueprint.Namespace) (blueprint.IRNode, error) {
		addr := &Address[ServerType]{}
		addr.AddrName = addressName
		return addr, nil
	})
	wiring.SetProperty(addressName, "pointsTo", pointsTo)

	// Add Config nodes for the server bind address and client address
	wiring.Define(bind(addressName), reachability, func(namespace blueprint.Namespace) (blueprint.IRNode, error) {
		return &BindConfig{Key: bind(addressName)}, nil
	})
	wiring.Define(dial(addressName), reachability, func(namespace blueprint.Namespace) (blueprint.IRNode, error) {
		return &DialConfig{Key: dial(addressName)}, nil
	})
}

func DestinationOf(namespace blueprint.Namespace, addressName string) (string, error) {
	var pointsTo string
	if err := namespace.GetProperty(addressName, "pointsTo", &pointsTo); err != nil {
		return "", blueprint.Errorf("expected pointsTo property of %v to be a string; %v", addressName, err.Error())
	}
	return pointsTo, nil
}

/*
The client side of an address should call this method to get the address to dial for a server

Under the hood this will ensure the configuration values for the dialling address get added to the namespace
*/
func Dial[ServerType blueprint.IRNode](namespace blueprint.Namespace, addressName string) (*Address[ServerType], error) {
	var addr *Address[ServerType]
	if err := namespace.Get(addressName, &addr); err != nil {
		return nil, err
	}

	// By getting the dial config value here, it gets implicitly added as an argument node to all namespaces
	var dialConf *DialConfig
	if err := namespace.Get(dial(addr.AddrName), &dialConf); err != nil {
		return nil, err
	}
	addr.Dial = dialConf

	return addr, nil
}

/*
The server side of an address should call this method to get the address to bind for a server

Under the hood this will ensure the configuration values for the binding address get added to the namespace
*/
func Bind[ServerType blueprint.IRNode](namespace blueprint.Namespace, addressName string) (*Address[ServerType], error) {
	var addr *Address[ServerType]
	if err := namespace.Get(addressName, &addr); err != nil {
		return nil, err
	}

	// By getting the bind config value here, it gets implicitly added as an argument node to all namespaces
	var bindConf *BindConfig
	if err := namespace.Get(bind(addr.AddrName), &bindConf); err != nil {
		return nil, err
	}
	addr.Bind = bindConf

	return addr, nil
}

func bind(addressName string) string {
	return blueprint.ReplaceSuffix(addressName, "addr", "bind_addr")
}

func dial(addressName string) string {
	return blueprint.ReplaceSuffix(addressName, "addr", "dial_addr")
}
