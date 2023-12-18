package address

import (
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/blueprint"
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/blueprint/stringutil"
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/ir"
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/wiring"
)

// Defines an address called addressName whose server-side node has name pointsTo.
//
// This method is primarily intended for use by other Blueprint plugins.
//
// The type parameter ServerType should correspond to the node type of pointsTo.
//
// Reachability of an address defines how far up the parent namespaces the address should
// exist and be reachable.  By default most addresses will want to use ir.ApplicationNode as
// the reachability to indicate that the address can be reached by any node anywhere in
// the application.
func Define[ServerType ir.IRNode](spec wiring.WiringSpec, addressName string, pointsTo string, reachability any) {
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

// Gets the [DialConfig] configuration node of addressName from the namespace.
//
// This method is intended for use by other Blueprint plugins within their own BuildFuncs.
//
// This is a convenience method for use when only the dial address is needed.  It is equivalent to getting
// addressName directly from namespace and then reading then [Address.Dial] field.
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

// Gets the [BindConfig] configuration node of addressName from the namespace.
//
// This method is intended for use by other Blueprint plugins within their own BuildFuncs.
//
// This is a convenience method for use when only the dial address is needed.  It is equivalent to getting
// addressName directly from namespace and then reading then [Address.Bind] field.
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

// Returns the value of pointsTo that was provided when addressName was defined.
//
// Used by the pointer plugin.
func PointsTo(namespace wiring.Namespace, addressName string) (string, error) {
	var pointsTo string
	if err := namespace.GetProperty(addressName, "pointsTo", &pointsTo); err != nil {
		return "", blueprint.Errorf("expected pointsTo property of %v to be a string; %v", addressName, err.Error())
	}
	if pointsTo == "" {
		return "", blueprint.Errorf("pointsTo is not set for %v", addressName)
	}
	return pointsTo, nil
}
