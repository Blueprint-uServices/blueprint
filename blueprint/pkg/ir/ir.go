// Package ir provides the basic interfaces for Blueprint's Internal Representation (IR)
// and for subsequently generating application artifacts such as code and container images.
//
// An application's IR representation is produced by constructing and then building a wiring
// spec using methods from the wiring package and from wiring extensions provided by plugins.
package ir

// All nodes implement the IRNode interface
type IRNode interface {
	Name() string
	String() string
}

// Metadata is an IR node that exists in the IR of an application but does not build
// any artifacts or provide configuration or anything like that.
type IRMetadata interface {
	IRNode
	ImplementsIRMetadata()
}

// IRConfig is an IR node that represents a configured or configurable variable.
// In a generated application, IRConfig nodes typically map down to things like
// environment variables or command line arguments, and can be passed all the way
// into specific application-level instances.  IRConfig is also used for addressing.
type IRConfig interface {
	IRNode
	Optional() bool
	// At various points during the build process, an IRConfig node might have a concrete value
	// set, or it might be left unbound.
	HasValue() bool

	// Returns the current value of the config node if it has been set.  Config values
	// are always strings.
	Value() string
	ImplementsIRConfig()
}

// A hard-coded value
type IRValue struct {
	Value string
}

func (v *IRValue) Name() string {
	return v.String()
}

func (v *IRValue) String() string {
	return "\"" + v.Value + "\""
}

// Most IRNodes can generate code artifacts but they do so in the context of some
// [BuildContext].  A few IRNodes, however, can generate artifacts independent of
// any external context.  Those IRNodes implement the ArtifactGenerator interface.
// Typically these are namespace nodes such as golang processes, linux containers,
// or docker deployments.
type ArtifactGenerator interface {

	// Generate all artifacts for this node to the specified dir on the local filesystem.
	GenerateArtifacts(dir string) error
}

// The IR Node that represents the whole application.  Building a wiring spec
// will return an ApplicationNode.  An ApplicationNode can be built
// with the GenerateArtifacts method.
type ApplicationNode struct {
	IRNode
	ArtifactGenerator

	ApplicationName string
	Children        []IRNode
}

func (node *ApplicationNode) Name() string {
	return node.ApplicationName
}

// Print the IR graph
func (node *ApplicationNode) String() string {
	return PrettyPrintNamespace(node.ApplicationName, "BlueprintApplication", nil, node.Children)
}

func (app *ApplicationNode) GenerateArtifacts(dir string) error {
	return defaultBuilders.buildAll(dir, app.Children)
}
