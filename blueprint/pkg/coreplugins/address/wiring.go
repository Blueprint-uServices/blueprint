package address

import (
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/blueprint/stringutil"
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/ir"
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/wiring"
)

// Additional optional options for use when defining an address
type AddressOpts struct {
	// Defines the nodes that can reach this address.  If left unspecified, an address
	// will be reachable application-wide (ie, Reachability uses the value &ir.ApplicationNode{}).
	// Plugins can restrict an address's reachability by specifying a more restrictive node type,
	// e.g. to restrict an address to only being reachable by nodes within the same container or machine.
	Reachability any
}

var defaultOpts = AddressOpts{
	Reachability: &ir.ApplicationNode{},
}

// Defines an address called addressName whose server-side node has name pointsTo.
//
// This method is primarily intended for use by other Blueprint plugins.
//
// The type parameter ServerType should correspond to the node type of pointsTo.
//
// By default the address is reachable application wide.  [AddressOpts] can be optionally provided
// to further configure the address.
func Define[ServerType ir.IRNode](spec wiring.WiringSpec, addressName string, pointsTo string, opts ...AddressOpts) {
	// Default address options
	options := defaultOpts
	if len(opts) > 0 {
		options = opts[0]
		// Don't bother merging multiple provided opts yet...
	}

	// Configure the address metadata in the wiring spec
	setAddressDef[ServerType](spec, addressName, pointsTo)

	// Define the IRMetadata node for the address, used during the build process
	spec.Define(addressName, options.Reachability, func(wiring.Namespace) (ir.IRNode, error) {
		addr := &Address[ServerType]{}
		addr.AddrName = addressName
		return addr, nil
	})

	// Add IRConfig nodes for the server bind address and client address
	spec.Define(bind(addressName), options.Reachability, func(wiring.Namespace) (ir.IRNode, error) {
		conf := &BindConfig{}
		conf.AddressName = addressName
		conf.Key = bind(addressName)
		return conf, nil
	})
	spec.Define(dial(addressName), options.Reachability, func(wiring.Namespace) (ir.IRNode, error) {
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

// Gets the [BindConfig] configuration node of addressName from the namespace and places it in dst
//
// This method is intended for use by other Blueprint plugins within their own BuildFuncs.
//
// This is a convenience method for use when only the bind address is needed.  It is equivalent to getting
// addressName directly from namespace and then reading then [Address.Bind] field.
//
// In addition to setting the [BindConfig] node in dst, this call sets the destination of the address to be serverNode
func Bind[ServerType ir.IRNode](namespace wiring.Namespace, addressName string, serverNode ServerType, dst **BindConfig) error {
	var addr *Address[ServerType]
	if err := namespace.Get(addressName, &addr); err != nil {
		return err
	}

	// By getting the bind config value here, it gets implicitly added as an argument node to all namespaces
	if err := namespace.Get(bind(addr.AddrName), dst); err != nil {
		return err
	}
	addr.Bind = *dst
	addr.Server = serverNode

	return nil
}

func bind(addressName string) string {
	return stringutil.ReplaceSuffix(addressName, "addr", "bind_addr")
}

func dial(addressName string) string {
	return stringutil.ReplaceSuffix(addressName, "addr", "dial_addr")
}

type AddressDef struct {
	Name       string
	PointsTo   string
	ServerType any
}

func setAddressDef[ServerType ir.IRNode](spec wiring.WiringSpec, addrName string, pointsTo string) {
	var serverType ServerType
	def := &AddressDef{
		Name:       addrName,
		PointsTo:   pointsTo,
		ServerType: serverType,
	}
	spec.SetProperty(addrName, "addr", def)
}

// Gets the AddressDef metadata for an address that was defined using [Define]
func GetAddress(spec wiring.WiringSpec, name string) *AddressDef {
	var def *AddressDef
	spec.GetProperty(name, "addr", &def)
	return def
}
