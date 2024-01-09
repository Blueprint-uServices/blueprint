package xtrace

import (
	"fmt"
	"path/filepath"
	"reflect"

	"github.com/blueprint-uservices/blueprint/blueprint/pkg/blueprint"
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/coreplugins/service"
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/ir"
	"github.com/blueprint-uservices/blueprint/plugins/golang"
	"github.com/blueprint-uservices/blueprint/plugins/golang/gocode"
	"github.com/blueprint-uservices/blueprint/plugins/golang/gogen"
	"golang.org/x/exp/slog"
)

// Blueprint IR Node that wraps the client-side of a service to generate xtrace compatible logs
type XtraceClientWrapper struct {
	golang.Service
	golang.Instantiable
	golang.GeneratesFuncs

	InstanceName  string
	outputPackage string
	Wrapped       golang.Service
	XTClient      *XTraceClient
}

func (node *XtraceClientWrapper) Name() string {
	return node.InstanceName
}

func (node *XtraceClientWrapper) String() string {
	return node.Name() + " = XtraceServerWrapper(" + node.Wrapped.Name() + ", " + node.XTClient.Name() + ")"
}

func (node *XtraceClientWrapper) ImplementsGolangNode()    {}
func (node *XtraceClientWrapper) ImplementsGolangService() {}

func newXtraceClientWrapper(name string, wrapped ir.IRNode, xtraceClient ir.IRNode) (*XtraceClientWrapper, error) {
	serverNode, is_callable := wrapped.(golang.Service)
	if !is_callable {
		return nil, blueprint.Errorf("xtrace client wrapper requires %s to be a golang service but got %s", wrapped.Name(), reflect.TypeOf(wrapped).String())
	}

	xtClient, is_client := xtraceClient.(*XTraceClient)
	if !is_client {
		return nil, blueprint.Errorf("xtrace client wrapper requires %s to be a xtrace client", xtraceClient.Name())
	}

	node := &XtraceClientWrapper{}
	node.InstanceName = name
	node.outputPackage = "xtrace"
	node.Wrapped = serverNode
	node.XTClient = xtClient
	return node, nil
}

func (node *XtraceClientWrapper) genInterface(ctx ir.BuildContext) (*gocode.ServiceInterface, error) {
	iface, err := golang.GetGoInterface(ctx, node.Wrapped)
	if err != nil {
		return nil, err
	}
	module_ctx, valid := ctx.(golang.ModuleBuilder)
	if !valid {
		return nil, blueprint.Errorf("XtraceClientWrapper expected build context to be a ModuleBuilder, got %v", ctx)
	}
	i := gocode.CopyServiceInterface(fmt.Sprintf("%v_XTraceClientWrapperInterface", iface.BaseName), module_ctx.Info().Name+"/"+node.outputPackage, iface)
	for name, method := range i.Methods {
		method.Arguments = method.Arguments[:len(method.Arguments)-1]
		method.Returns = method.Returns[:len(method.Returns)-1]
		i.Methods[name] = method
	}
	return i, nil
}

func (node *XtraceClientWrapper) AddInstantiation(builder golang.NamespaceBuilder) error {
	if builder.Visited(node.InstanceName) {
		return nil
	}

	iface, err := golang.GetGoInterface(builder, node.Wrapped)
	if err != nil {
		return err
	}

	xtrace_iface, err := golang.GetGoInterface(builder, node.XTClient)
	if err != nil {
		return err
	}

	constructor := &gocode.Constructor{
		Package: builder.Module().Info().Name + "/" + node.outputPackage,
		Func: gocode.Func{
			Name: fmt.Sprintf("New_%v_XTraceClientWrapper", iface.BaseName),
			Arguments: []gocode.Variable{
				{Name: "ctx", Type: &gocode.UserType{Package: "context", Name: "Context"}},
				{Name: "client", Type: iface},
				{Name: "xtraceClient", Type: xtrace_iface},
			},
		},
	}

	return builder.DeclareConstructor(node.InstanceName, constructor, []ir.IRNode{node.Wrapped, node.XTClient})
}

func (node *XtraceClientWrapper) GenerateFuncs(builder golang.ModuleBuilder) error {
	wrapped_iface, err := golang.GetGoInterface(builder, node.Wrapped)
	if err != nil {
		return err
	}

	xtrace_iface, err := golang.GetGoInterface(builder, node.XTClient)
	if err != nil {
		return err
	}

	impl_iface, err := node.genInterface(builder)
	if err != nil {
		return err
	}

	return generateClientHandler(builder, wrapped_iface, impl_iface, xtrace_iface, node.outputPackage)
}

func (node *XtraceClientWrapper) GetInterface(ctx ir.BuildContext) (service.ServiceInterface, error) {
	return node.genInterface(ctx)
}

func (node *XtraceClientWrapper) AddInterfaces(builder golang.ModuleBuilder) error {
	return node.Wrapped.AddInterfaces(builder)
}

func generateClientHandler(builder golang.ModuleBuilder, wrapped *gocode.ServiceInterface, impl *gocode.ServiceInterface, xt_iface *gocode.ServiceInterface, outputPackage string) error {
	pkg, err := builder.CreatePackage(outputPackage)
	if err != nil {
		return err
	}

	server := &clientArgs{
		Package:         pkg,
		Service:         wrapped,
		Impl:            impl,
		XTIface:         xt_iface,
		Name:            wrapped.BaseName + "_XTraceClientWrapper",
		IfaceName:       impl.Name,
		ServerIfaceName: wrapped.BaseName + "_XTraceServerWrapperInterface",
		Imports:         gogen.NewImports(pkg.Name),
	}

	server.Imports.AddPackages("context", "github.com/tracingplane/tracingplane-go/tracingplane")

	slog.Info(fmt.Sprintf("Generating %v/%v", server.Package.PackageName, impl.Name))
	outputFile := filepath.Join(server.Package.Path, impl.Name+".go")
	return gogen.ExecuteTemplateToFile("XTrace", clientSideTemplate, server, outputFile)
}

type clientArgs struct {
	Package         golang.PackageInfo
	Service         *gocode.ServiceInterface
	Impl            *gocode.ServiceInterface
	XTIface         *gocode.ServiceInterface
	Name            string
	IfaceName       string
	ServerIfaceName string
	Imports         *gogen.Imports
}

var clientSideTemplate = `// Blueprint: Auto-generated by XTrace Plugin
package {{.Package.ShortName}}

{{.Imports}}

type {{.IfaceName}} interface {
	{{range $_, $f := .Impl.Methods -}}
	{{Signature $f}}
	{{end}}
}

type {{.Name}} struct {
	Client {{.ServerIfaceName}}
	XTClient {{.Imports.NameOf .XTIface.UserType}}
}

func New_{{.Name}}(ctx context.Context, client {{.ServerIfaceName}}, xtclient {{.Imports.NameOf .XTIface.UserType}}) (*{{.Name}}, error) {
	handler := &{{.Name}}{}
	handler.Client = client
	handler.XTClient = xtclient
	return handler, nil
}

{{$service := .Service.Name -}}
{{$receiver := .Name -}}
{{range $_, $f := .Impl.Methods}}
func (handler *{{$receiver}}) {{$f.Name -}} ({{ArgVarsAndTypes $f "ctx context.Context"}}) ({{RetVarsAndTypes $f "err error"}}) {
	ctx, _ = handler.XTClient.Log(ctx, "{{$f.Name}} client call start")
	baggage_raw, _ := handler.XTClient.Get(ctx)
	baggage := tracingplane.EncodeBase64(baggage_raw)

	var ret_baggage string
	{{RetVars $f "ret_baggage" "err"}} = handler.Client.{{$f.Name}}({{ArgVars $f "ctx"}}, baggage)
	ret_baggage_raw, _ := tracingplane.DecodeBase64(ret_baggage)
	ctx, _ = handler.XTClient.Merge(ctx, ret_baggage_raw)
	if err != nil {
		ctx, _ = handler.XTClient.LogWithTags(ctx, err.Error(), "Error")
	}
	ctx, _ = handler.XTClient.Log(ctx, "{{$f.Name}} client call end")
	return
}
{{end}}
`
