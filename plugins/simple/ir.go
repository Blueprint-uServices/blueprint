package simple

import (
	"fmt"

	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/coreplugins/service"
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/ir"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/golang"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/golang/gocode"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/golang/goparser"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/workflow"
	"golang.org/x/exp/slog"
)

// The SimpleBackend IR node represents a service or backend implementation that is wholly
// defined in Blueprint's runtime module.  Examples include SimpleCache, SimpleNoSQLDB, etc.
//
// The compiled SimpleBackend node will simply include the runtime module in the compiled output,
// and create instances of the service / backend by calling the appropriate constructors from the
// runtime module.
type SimpleBackend struct {
	golang.Service

	// Interfaces for generating Golang artifacts
	golang.ProvidesModule
	golang.Instantiable

	InstanceName string
	BackendType  string // e.g. "NoSQLDatabase"
	BackendImpl  string // e.g. "SimpleNoSQLDB"

	Iface       *goparser.ParsedInterface // The backend's code interface
	Constructor *gocode.Constructor       // Constructor for this backend implementation
}

// Creates a [SimpleBackend] IR node.
//   - name should be a name for the instance, e.g. "my_nosql_db"
//   - backendType should be the name of the interface this backend implements, e.g. "NoSQLDatabase"
//   - backendImpl should be the name of the implementation, e.g. "SimpleNoSQLDB"
func newSimpleBackend(name, backendType, backendImpl string) (*SimpleBackend, error) {
	node := &SimpleBackend{
		InstanceName: name,
		BackendType:  backendType,
		BackendImpl:  backendImpl,
	}
	return node, node.init()
}

func (node *SimpleBackend) init() error {
	// We use the workflow spec to load the simple backend interface details
	workflow.Init("../../runtime")

	// Look up the backend details; errors out if the backend doesn't exist
	spec, err := workflow.GetSpec()
	if err != nil {
		return err
	}
	details, err := spec.Get(node.BackendImpl)
	if err != nil {
		return err
	}

	node.Iface = details.Iface
	node.Constructor = details.Constructor.AsConstructor()
	return nil
}

func (node *SimpleBackend) Name() string {
	return node.InstanceName
}

func (node *SimpleBackend) GetInterface(ctx ir.BuildContext) (service.ServiceInterface, error) {
	return node.Iface.ServiceInterface(ctx), nil
}

func (node *SimpleBackend) AddToWorkspace(builder golang.WorkspaceBuilder) error {
	// The backend interface and impl implementation exist in the runtime package
	// Add blueprint runtime to the workspace
	return golang.AddRuntimeModule(builder)
}

func (node *SimpleBackend) AddInterfaces(builder golang.ModuleBuilder) error {
	return node.AddToWorkspace(builder.Workspace())
}

func (node *SimpleBackend) AddInstantiation(builder golang.NamespaceBuilder) error {
	// Only generate instantiation code for this instance once
	if builder.Visited(node.InstanceName) {
		return nil
	}

	slog.Info(fmt.Sprintf("Instantiating %v %v in %v/%v", node.BackendImpl, node.InstanceName, builder.Info().Package.PackageName, builder.Info().FileName))
	return builder.DeclareConstructor(node.InstanceName, node.Constructor, nil)
}

func (node *SimpleBackend) String() string {
	return fmt.Sprintf("%v = %v()", node.InstanceName, node.BackendImpl)
}

func (node *SimpleBackend) ImplementsGolangNode()    {}
func (node *SimpleBackend) ImplementsGolangService() {}
