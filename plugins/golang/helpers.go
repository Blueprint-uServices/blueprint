package golang

import (
	"reflect"

	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/blueprint"
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/coreplugins/service"
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/ir"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/golang/gocode"
	"golang.org/x/exp/slog"
)

/*
Helper method that does typecasting on builder and service.

Assumes builder is a golang module builder, and service is a golang module; if so, gets the golang
service interface for the service.  If not, returns an error.
*/
func GetGoInterface(ctx ir.BuildContext, node ir.IRNode) (*gocode.ServiceInterface, error) {
	service, isService := node.(service.ServiceNode)
	if !isService {
		return nil, blueprint.Errorf("cannot get a service interface from non-service node %v", node)
	}
	if graph, isGraphBuilder := ctx.(NamespaceBuilder); isGraphBuilder {
		ctx = graph.Module()
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

func AddRuntimeModule(workspace WorkspaceBuilder) error {
	if !workspace.Visited("runtime") {
		slog.Info("Copying local module runtime to workspace")
		return workspace.AddLocalModuleRelative("runtime", "../../runtime")
	}
	return nil
}
