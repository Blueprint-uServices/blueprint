package http

import (
	"errors"
	"fmt"
	"reflect"

	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/blueprint"
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/core/irutil"
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/core/service"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/golang"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/golang/gocode"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/http/httpcodegen"
)

type GolangHttpClient struct {
	golang.Node
	golang.Service
	golang.GeneratesFuncs
	golang.Instantiable

	InstanceName string
	ServerAddr   *GolangHttpServerAddress

	outputPackage string
}

func newGolangHttpClient(name string, serverAddr blueprint.IRNode) (*GolangHttpClient, error) {
	addr, is_addr := serverAddr.(*GolangHttpServerAddress)
	if !is_addr {
		return nil, errors.New(fmt.Sprintf("HTTP client %s expected %s to be an address, but got %s", name, serverAddr.Name(), reflect.TypeOf(serverAddr).String()))
	}

	node := &GolangHttpClient{}
	node.InstanceName = name
	node.ServerAddr = addr
	node.outputPackage = "http"

	return node, nil
}

func (n *GolangHttpClient) String() string {
	return n.InstanceName + " = HTTPClient(" + n.ServerAddr.Name() + ")"
}

func (n *GolangHttpClient) Name() string {
	return n.InstanceName
}

func (n *GolangHttpClient) GetInterface(visitor irutil.BuildContext) service.ServiceInterface {
	return n.GetGoInterface(visitor)
}

func (n *GolangHttpClient) GetGoInterface(visitor irutil.BuildContext) *gocode.ServiceInterface {
	http, isHttp := n.ServerAddr.GetInterface(visitor).(*HttpInterface)
	if !isHttp {
		return nil
	}
	wrapped, isValid := http.Wrapped.(*gocode.ServiceInterface)
	if !isValid {
		return nil
	}
	return wrapped
}

func (node *GolangHttpClient) GenerateFuncs(builder golang.ModuleBuilder) error {
	if builder.Visited(node.InstanceName + ".generateFuncs") {
		return nil
	}

	service := node.GetGoInterface(builder)
	if service == nil {
		return errors.New(fmt.Sprintf("expected %v to have a gocode.ServiceInterface but got %v", node.Name(), node.ServerAddr.GetInterface(builder)))
	}

	return httpcodegen.GenerateClient(builder, service, node.outputPackage)
}

func (node *GolangHttpClient) AddInstantiation(builder golang.GraphBuilder) error {
	// Only generate instantiation code for this instance once
	if builder.Visited(node.InstanceName) {
		return nil
	}

	constructor := &gocode.Constructor{
		Package: builder.Module().Info().Name + "/" + node.outputPackage,
		Func: gocode.Func{
			Name: fmt.Sprintf("New_%v_HTTPClient", node.GetGoInterface(builder).Name),
			Arguments: []gocode.Variable{
				{Name: "addr", Type: &gocode.BasicType{Name: "string"}},
			},
		},
	}

	return builder.DeclareConstructor(node.InstanceName, constructor, []blueprint.IRNode{node.ServerAddr})
}

func (node *GolangHttpClient) ImplementsGolangNode()    {}
func (node *GolangHttpClient) ImplementsGolangService() {}
