package http

import (
	"fmt"

	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/blueprint"
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

func newGolangHttpClient(name string, addr *GolangHttpServerAddress) (*GolangHttpClient, error) {
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

func (node *GolangHttpClient) GetInterface(ctx blueprint.BuildContext) (service.ServiceInterface, error) {
	iface, err := node.ServerAddr.GetInterface(ctx)
	if err != nil {
		return nil, err
	}
	http, isHttp := iface.(*HttpInterface)
	if !isHttp {
		return nil, fmt.Errorf("http client expected an HTTP interface from %v but found %v", node.ServerAddr.Name(), iface)
	}
	wrapped, isValid := http.Wrapped.(*gocode.ServiceInterface)
	if !isValid {
		return nil, fmt.Errorf("http client expected the server's HTTP interface to wrap a gocode interface but found %v", http)
	}
	return wrapped, nil
}

// Just makes sure that the interface exposed by the server is included in the built module
func (node *GolangHttpClient) AddInterfaces(builder golang.ModuleBuilder) error {
	return node.ServerAddr.Server.Wrapped.AddInterfaces(builder)
}

func (node *GolangHttpClient) GenerateFuncs(builder golang.ModuleBuilder) error {
	if builder.Visited(node.InstanceName + ".generateFuncs") {
		return nil
	}

	iface, err := golang.GetGoInterface(builder, node)
	if err != nil {
		return err
	}

	return httpcodegen.GenerateClient(builder, iface, node.outputPackage)
}

func (node *GolangHttpClient) AddInstantiation(builder golang.GraphBuilder) error {
	// Only generate instantiation code for this instance once
	if builder.Visited(node.InstanceName) {
		return nil
	}

	iface, err := golang.GetGoInterface(builder, node)
	if err != nil {
		return err
	}

	constructor := &gocode.Constructor{
		Package: builder.Module().Info().Name + "/" + node.outputPackage,
		Func: gocode.Func{
			Name: fmt.Sprintf("New_%v_HTTPClient", iface.Name),
			Arguments: []gocode.Variable{
				{Name: "addr", Type: &gocode.BasicType{Name: "string"}},
			},
		},
	}

	return builder.DeclareConstructor(node.InstanceName, constructor, []blueprint.IRNode{node.ServerAddr})
}

func (node *GolangHttpClient) ImplementsGolangNode()    {}
func (node *GolangHttpClient) ImplementsGolangService() {}
