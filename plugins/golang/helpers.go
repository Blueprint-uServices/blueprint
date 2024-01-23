package golang

import (
	"fmt"
	"reflect"

	"github.com/blueprint-uservices/blueprint/blueprint/pkg/blueprint"
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/coreplugins/service"
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/ir"
	"github.com/blueprint-uservices/blueprint/plugins/golang/gocode"
)

// A convenience function that can be called by other Blueprint plugins.
//
// If ctx is a [ModuleBuilder] and node is a [Service], this method returns the [*gocode.ServiceInterface]
// of node.  If not, returns nil and an error.
func GetGoInterface(ctx ir.BuildContext, node ir.IRNode) (*gocode.ServiceInterface, error) {
	service, isService := node.(service.ServiceNode)
	if !isService {
		return nil, blueprint.Errorf("cannot get a service interface from non-service node %v", node)
	}
	if n, isNamespace := ctx.(NamespaceBuilder); isNamespace {
		ctx = n.Module()
	}
	module, isModuleBuilder := ctx.(ModuleBuilder)
	if !isModuleBuilder {
		return nil, blueprint.Errorf("cannot get a golang interface for service %v from non-golang module builder %v", node, ctx)
	}
	iface, err := service.GetInterface(module)
	if err != nil {
		return nil, err
	}
	if goIface, isGoIface := iface.(*gocode.ServiceInterface); isGoIface {
		return goIface, nil
	}
	if goService, isGoService := service.(Service); isGoService {
		return nil, blueprint.Errorf("golang service %v should implement a gocode.ServiceInterface but GetInterface() returned a %v instead", goService.Name(), reflect.TypeOf(iface))
	} else {
		return nil, blueprint.Errorf("getGoInterface on non-golang service %v returned non-gocode.ServiceInterface", service)
	}
}

// A convenience function that can be called by other Blueprint plugins.
// Ensures that Blueprint's [runtime] module is copied to the output workspace.
//
// [runtime]: https://github.com/Blueprint-uServices/blueprint/tree/main/runtime
func AddRuntimeModule(module ModuleBuilder) error {
	// TODO: find runtime module, copy it to workspace if it's local, or add versioned dependency to module if not
	// TODO: push this into namespace impl so that plugins don't need to do it.
	return fmt.Errorf("AddRuntimeModule not implemented")
}
