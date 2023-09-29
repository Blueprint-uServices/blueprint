package http

import (
	"fmt"
	"reflect"

	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/blueprint"
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/core/service"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/golang"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/golang/gocode"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/http/httpcodegen"
)

type GolangHttpServer struct {
	service.ServiceNode
	golang.GeneratesFuncs
	golang.Instantiable

	InstanceName string
	Addr         *GolangHttpServerAddress
	Wrapped      golang.Service

	outputPackage string
}

//Represents a service that is exposed over HTTP
type HttpInterface struct {
	service.ServiceInterface
	Wrapped service.ServiceInterface
}

func (i *HttpInterface) GetName() string {
	return "http(" + i.Wrapped.GetName() + ")"
}

func (i *HttpInterface) GetMethods() []service.Method {
	return i.Wrapped.GetMethods()
}

func newGolangHttpServer(name string, serverAddr blueprint.IRNode, wrapped blueprint.IRNode) (*GolangHttpServer, error) {
	addr, is_addr := serverAddr.(*GolangHttpServerAddress)
	if !is_addr {
		return nil, blueprint.Errorf("HTTP server %s expected %s to be an address, but got %s", name, serverAddr.Name(), reflect.TypeOf(serverAddr).String())
	}

	service, is_service := wrapped.(golang.Service)
	if !is_service {
		return nil, blueprint.Errorf("HTTP server %s expected %s to be a golang service, but got %s", name, wrapped.Name(), reflect.TypeOf(wrapped).String())
	}

	node := &GolangHttpServer{}
	node.InstanceName = name
	node.Addr = addr
	node.Wrapped = service
	node.outputPackage = "http"
	return node, nil
}

func (n *GolangHttpServer) String() string {
	return n.InstanceName + " = HTTPServer(" + n.Wrapped.Name() + ", " + n.Addr.Name() + ")"
}

func (n *GolangHttpServer) Name() string {
	return n.InstanceName
}

// Generates the HTTP Server handler
func (node *GolangHttpServer) GenerateFuncs(builder golang.ModuleBuilder) error {
	service, valid := node.Wrapped.GetInterface().(*gocode.ServiceInterface)
	if !valid {
		return blueprint.Errorf("expected %v to have a gocode.ServiceInterface but got %v", node.Name(), node.Wrapped.GetInterface())
	}

	err := httpcodegen.GenerateServerHandler(builder, service, node.outputPackage)
	if err != nil {
		return err
	}
	return nil
}

func (node *GolangHttpServer) AddInstantiation(builder golang.GraphBuilder) error {
	// Only generate instantiation code for this instance once
	if builder.Visited(node.InstanceName) {
		return nil
	}

	service, valid := node.Wrapped.GetInterface().(*gocode.ServiceInterface)
	if !valid {
		return blueprint.Errorf("expected %v to have a gocode.ServiceInterface but got %v", node.Name(), node.Wrapped.GetInterface())
	}

	constructor := &gocode.Constructor{
		Package: builder.Module().Info().Name + "/" + node.outputPackage,
		Func: gocode.Func{
			Name: fmt.Sprintf("New_%v_HTTPServerHandler", service.Name),
			Arguments: []gocode.Variable{
				{Name: "service", Type: service},
				{Name: "serverAddr", Type: &gocode.BasicType{Name: "string"}},
			},
		},
	}
	return builder.DeclareConstructor(node.InstanceName, constructor, []blueprint.IRNode{node.Wrapped, node.Addr})
}

func (node *GolangHttpServer) GetInterface() service.ServiceInterface {
	return &HttpInterface{Wrapped: node.Wrapped.GetInterface()}
}

func (node *GolangHttpServer) ImplementsGolangNode() {}
