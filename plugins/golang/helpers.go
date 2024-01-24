package golang

import (
	"reflect"

	"github.com/blueprint-uservices/blueprint/blueprint/pkg/blueprint"
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/coreplugins/service"
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/ir"
	"github.com/blueprint-uservices/blueprint/plugins/golang/gocode"
	"github.com/blueprint-uservices/blueprint/plugins/golang/goparser"
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
// Looks up the specified moduleName (assuming it is a dependency of the current module),
// with the intention of adding it as a dependency to the provided build context.
//
// If ctx is a [ModuleBuilder], this method adds the module as a dependency to the module,
// ie. as a 'require' to go.mod, but ONLY if the module isn't a local module (ie. with a replace directive).
//
// If ctx is a [WorkspaceBuilder], this method copies the module to the output workspace,
// but ONLY if the module is a local module (ie. with a replace directive).
func AddModule(ctx ir.BuildContext, moduleName string) error {
	modInfo, err := goparser.FindPackageModule(moduleName)
	if err != nil {
		return err
	}

	switch builder := ctx.(type) {
	case ModuleBuilder:
		{
			if !modInfo.IsLocal {
				if !builder.Visited(moduleName) {
					return builder.Require(modInfo.Path, modInfo.Version)
				}
			} else {
				return AddModule(builder.Workspace(), moduleName)
			}
		}
	case WorkspaceBuilder:
		{
			if modInfo.IsLocal {
				if !builder.Visited(moduleName) {
					_, err = builder.AddLocalModule(modInfo.ShortName, modInfo.Dir)
					return err
				}
			}
			return nil
		}
	}
	return nil
}

// A convenience function that can be called by other Blueprint plugins.
// If mod is not a local module, ensures that it is added as a 'require' to go.mod.
func AddToModule(builder ModuleBuilder, mods ...*goparser.ParsedModule) error {
	for _, mod := range mods {
		if !mod.IsLocal {
			if !builder.Visited(mod.Name) {
				if err := builder.Require(mod.Name, mod.Version); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

// A convenience function that can be called by other Blueprint plugins.
// If mod is a local module, ensures that it is copied to the output workspace.
func AddToWorkspace(builder WorkspaceBuilder, mods ...*goparser.ParsedModule) error {
	for _, mod := range mods {
		if mod.IsLocal {
			if !builder.Visited(mod.Name) {
				_, err := builder.AddLocalModule(mod.ShortName, mod.SrcDir)
				return err
			}
		}
	}
	return nil
}
