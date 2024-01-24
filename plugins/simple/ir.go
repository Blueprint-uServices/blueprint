package simple

import (
	"fmt"

	"github.com/blueprint-uservices/blueprint/blueprint/pkg/coreplugins/service"
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/ir"
	"github.com/blueprint-uservices/blueprint/plugins/golang"
	"github.com/blueprint-uservices/blueprint/plugins/golang/gocode"
	"github.com/blueprint-uservices/blueprint/plugins/workflow/workflowspec"
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
	golang.ProvidesInterface
	golang.Instantiable

	InstanceName string
	BackendType  string // e.g. "NoSQLDatabase"
	BackendImpl  string // e.g. "SimpleNoSQLDB"

	Spec *workflowspec.Service // The backend's interface and implementation
}

// Creates a [SimpleBackend] IR node.
//   - name should be a name for the instance, e.g. "my_nosql_db"
//   - BackendIface should be the the interface this backend implements, e.g. "NoSQLDatabase"
//   - BackendImpl should be the the implementation, e.g. "SimpleNoSQLDB"
func newSimpleBackend[BackendImpl any](name string) (*SimpleBackend, error) {
	spec, err := workflowspec.GetService[BackendImpl]()
	if err != nil {
		return nil, err
	}
	node := &SimpleBackend{
		InstanceName: name,
		Spec:         spec,
		BackendType:  spec.Iface.Name,
		BackendImpl:  gocode.NameOf[BackendImpl](),
	}

	return node, nil
}

// Implements ir.IRNode
func (node *SimpleBackend) Name() string {
	return node.InstanceName
}

// Implements golang.Service
func (node *SimpleBackend) GetInterface(ctx ir.BuildContext) (service.ServiceInterface, error) {
	return node.Spec.Iface.ServiceInterface(ctx), nil
}

// Implements golang.ProvidesModule
func (node *SimpleBackend) AddToWorkspace(builder golang.WorkspaceBuilder) error {
	return node.Spec.AddToWorkspace(builder)
}

// Implements golang.ProvidesInterface
func (node *SimpleBackend) AddInterfaces(builder golang.ModuleBuilder) error {
	return node.Spec.AddToModule(builder)
}

// Ipmlements golang.Instantiable
func (node *SimpleBackend) AddInstantiation(builder golang.NamespaceBuilder) error {
	// Only generate instantiation code for this instance once
	if builder.Visited(node.InstanceName) {
		return nil
	}

	slog.Info(fmt.Sprintf("Instantiating %v %v in %v/%v", node.BackendImpl, node.InstanceName, builder.Info().Package.PackageName, builder.Info().FileName))
	return builder.DeclareConstructor(node.InstanceName, node.Spec.Constructor.AsConstructor(), nil)
}

// Implements ir.IRNode
func (node *SimpleBackend) String() string {
	return fmt.Sprintf("%v = %v()", node.InstanceName, node.BackendImpl)
}

func (node *SimpleBackend) ImplementsGolangNode()    {}
func (node *SimpleBackend) ImplementsGolangService() {}
