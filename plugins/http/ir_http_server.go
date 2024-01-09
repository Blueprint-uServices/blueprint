package http

import (
	"fmt"
	"reflect"

	"github.com/blueprint-uservices/blueprint/blueprint/pkg/blueprint"
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/coreplugins/address"
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/coreplugins/service"
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/ir"
	"github.com/blueprint-uservices/blueprint/plugins/golang"
	"github.com/blueprint-uservices/blueprint/plugins/golang/gocode"
	"github.com/blueprint-uservices/blueprint/plugins/http/httpcodegen"
)

// IRNode representing a Golang HTTP server.
// This node does not introduce any new runtime interfaces or types that can be used by other IRNodes.
type golangHttpServer struct {
	service.ServiceNode
	golang.GeneratesFuncs
	golang.Instantiable

	InstanceName string
	Bind         *address.BindConfig
	Wrapped      golang.Service

	outputPackage string
}

// Represents a service that is exposed over HTTP
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

func newGolangHttpServer(name string, wrapped ir.IRNode) (*golangHttpServer, error) {
	service, is_service := wrapped.(golang.Service)
	if !is_service {
		return nil, blueprint.Errorf("HTTP server %s expected %s to be a golang service, but got %s", name, wrapped.Name(), reflect.TypeOf(wrapped).String())
	}

	node := &golangHttpServer{}
	node.InstanceName = name
	node.Wrapped = service
	node.outputPackage = "http"
	return node, nil
}

func (n *golangHttpServer) String() string {
	return n.InstanceName + " = HTTPServer(" + n.Wrapped.Name() + ", " + n.Bind.Name() + ")"
}

func (n *golangHttpServer) Name() string {
	return n.InstanceName
}

// Generates the HTTP Server handler
func (node *golangHttpServer) GenerateFuncs(builder golang.ModuleBuilder) error {
	iface, err := golang.GetGoInterface(builder, node.Wrapped)
	if err != nil {
		return err
	}

	err = httpcodegen.GenerateServerHandler(builder, iface, node.outputPackage)
	if err != nil {
		return err
	}
	return nil
}

func (node *golangHttpServer) AddInstantiation(builder golang.NamespaceBuilder) error {
	// Only generate instantiation code for this instance once
	if builder.Visited(node.InstanceName) {
		return nil
	}

	iface, err := golang.GetGoInterface(builder, node.Wrapped)
	if err != nil {
		return err
	}

	constructor := &gocode.Constructor{
		Package: builder.Module().Info().Name + "/" + node.outputPackage,
		Func: gocode.Func{
			Name: fmt.Sprintf("New_%v_HTTPServerHandler", iface.BaseName),
			Arguments: []gocode.Variable{
				{Name: "ctx", Type: &gocode.UserType{Package: "context", Name: "Context"}},
				{Name: "service", Type: iface},
				{Name: "serverAddr", Type: &gocode.BasicType{Name: "string"}},
			},
		},
	}
	return builder.DeclareConstructor(node.InstanceName, constructor, []ir.IRNode{node.Wrapped, node.Bind})
}

func (node *golangHttpServer) GetInterface(ctx ir.BuildContext) (service.ServiceInterface, error) {
	iface, err := node.Wrapped.GetInterface(ctx)
	return &HttpInterface{Wrapped: iface}, err
}

func (node *golangHttpServer) ImplementsGolangNode() {}
