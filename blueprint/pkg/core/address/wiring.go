package address

import (
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/blueprint"
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/ir"
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/stringutil"
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/wiring"
)

/*
Defines an address called `addressName` that points to the definition `pointsto`.
*/
func Define[ServerType ir.IRNode](spec wiring.WiringSpec, addressName string, pointsTo string, reachability any) {
	def := spec.GetDef(pointsTo)
	if def == nil {
		spec.AddError(blueprint.Errorf("trying to define address %s that points to %s but %s is not defined", addressName, pointsTo, pointsTo))
	}

	// Define the metadata for the address, used during the build process
	spec.Define(addressName, reachability, func(namespace wiring.Namespace) (ir.IRNode, error) {
		addr := &Address[ServerType]{}
		addr.AddrName = addressName
		return addr, nil
	})
	spec.SetProperty(addressName, "pointsTo", pointsTo)

	// Add Config nodes for the server bind address and client address
	spec.Define(bind(addressName), reachability, func(namespace wiring.Namespace) (ir.IRNode, error) {
		conf := &BindConfig{}
		conf.AddressName = addressName
		conf.Key = bind(addressName)
		return conf, nil
	})
	spec.Define(dial(addressName), reachability, func(namespace wiring.Namespace) (ir.IRNode, error) {
		conf := &DialConfig{}
		conf.AddressName = addressName
		conf.Key = dial(addressName)
		return conf, nil
	})
}

func DestinationOf(namespace wiring.Namespace, addressName string) (string, error) {
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
func Dial[ServerType ir.IRNode](namespace wiring.Namespace, addressName string) (*Address[ServerType], error) {
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
func Bind[ServerType ir.IRNode](namespace wiring.Namespace, addressName string) (*Address[ServerType], error) {
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
	return stringutil.ReplaceSuffix(addressName, "addr", "bind_addr")
}

func dial(addressName string) string {
	return stringutil.ReplaceSuffix(addressName, "addr", "dial_addr")
}
