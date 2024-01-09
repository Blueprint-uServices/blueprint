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

// Blueprint IR Node that wraps the server-side of a service to generate xtrace compatible logs
type XtraceServerWrapper struct {
	golang.Service
	golang.Instantiable
	golang.GeneratesFuncs

	InstanceName  string
	outputPackage string
	Wrapped       golang.Service
	XTClient      *XTraceClient
}

func (node *XtraceServerWrapper) Name() string {
	return node.InstanceName
}

func (node *XtraceServerWrapper) String() string {
	return node.Name() + " = XtraceServerWrapper(" + node.Wrapped.Name() + ", " + node.XTClient.Name() + ")"
}

func (node *XtraceServerWrapper) ImplementsGolangNode()    {}
func (node *XtraceServerWrapper) ImplementsGolangService() {}

func newXtraceServerWrapper(name string, wrapped ir.IRNode, xtraceClient ir.IRNode) (*XtraceServerWrapper, error) {
	serverNode, is_callable := wrapped.(golang.Service)
	if !is_callable {
		return nil, blueprint.Errorf("xtrace server wrapper requires %s to be a golang service but got %s", wrapped.Name(), reflect.TypeOf(wrapped).String())
	}

	xtClient, is_client := xtraceClient.(*XTraceClient)
	if !is_client {
		return nil, blueprint.Errorf("xtrace server wrapper requires %s to be a xtrace client", xtraceClient.Name())
	}

	node := &XtraceServerWrapper{}
	node.InstanceName = name
	node.outputPackage = "xtrace"
	node.Wrapped = serverNode
	node.XTClient = xtClient
	return node, nil
}

func (node *XtraceServerWrapper) genInterface(ctx ir.BuildContext) (*gocode.ServiceInterface, error) {
	iface, err := golang.GetGoInterface(ctx, node.Wrapped)
	if err != nil {
		return nil, err
	}
	module_ctx, valid := ctx.(golang.ModuleBuilder)
	if !valid {
		return nil, blueprint.Errorf("XtraceServerWrapper expected build context to be a ModuleBuilder, got %v", ctx)
	}
	i := gocode.CopyServiceInterface(fmt.Sprintf("%v_XTraceServerWrapperInterface", iface.BaseName), module_ctx.Info().Name+"/"+node.outputPackage, iface)
	for name, method := range i.Methods {
		method.AddArgument(gocode.Variable{Name: "baggage", Type: &gocode.BasicType{Name: "string"}})
		method.AddRetVar(gocode.Variable{Name: "", Type: &gocode.BasicType{Name: "string"}})
		i.Methods[name] = method
	}
	return i, nil
}

func (node *XtraceServerWrapper) AddInstantiation(builder golang.NamespaceBuilder) error {
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
			Name: fmt.Sprintf("New_%v_XTraceServerWrapper", iface.BaseName),
			Arguments: []gocode.Variable{
				{Name: "ctx", Type: &gocode.UserType{Package: "context", Name: "Context"}},
				{Name: "service", Type: iface},
				{Name: "xtraceClient", Type: xtrace_iface},
			},
		},
	}

	return builder.DeclareConstructor(node.InstanceName, constructor, []ir.IRNode{node.Wrapped, node.XTClient})
}

func (node *XtraceServerWrapper) GenerateFuncs(builder golang.ModuleBuilder) error {
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

	return generateServerHandler(builder, wrapped_iface, impl_iface, xtrace_iface, node.outputPackage)
}

func (node *XtraceServerWrapper) GetInterface(ctx ir.BuildContext) (service.ServiceInterface, error) {
	return node.genInterface(ctx)
}

func (node *XtraceServerWrapper) AddInterfaces(builder golang.ModuleBuilder) error {
	iface, err := node.genInterface(builder)
	if err != nil {
		return err
	}

	err = generateClientSideInterfaces(builder, iface, node.outputPackage)
	if err != nil {
		return err
	}

	return node.Wrapped.AddInterfaces(builder)
}

func generateServerHandler(builder golang.ModuleBuilder, wrapped *gocode.ServiceInterface, impl *gocode.ServiceInterface, xt_iface *gocode.ServiceInterface, outputPackage string) error {
	pkg, err := builder.CreatePackage(outputPackage)
	if err != nil {
		return err
	}

	server := &serverArgs{
		Package:   pkg,
		Service:   wrapped,
		Impl:      impl,
		XTIface:   xt_iface,
		Name:      wrapped.BaseName + "_XTraceServerWrapper",
		IfaceName: impl.Name,
		Imports:   gogen.NewImports(pkg.Name),
	}

	server.Imports.AddPackages("context", "github.com/tracingplane/tracingplane-go/tracingplane")

	slog.Info(fmt.Sprintf("Generating %v/%v", server.Package.PackageName, impl.Name))
	outputFile := filepath.Join(server.Package.Path, impl.Name+".go")
	return gogen.ExecuteTemplateToFile("XTrace", serverTemplate, server, outputFile)
}

type serverArgs struct {
	Package   golang.PackageInfo
	Service   *gocode.ServiceInterface
	Impl      *gocode.ServiceInterface
	XTIface   *gocode.ServiceInterface
	Name      string
	IfaceName string
	Imports   *gogen.Imports
}

func generateClientSideInterfaces(builder golang.ModuleBuilder, iface *gocode.ServiceInterface, outputPackage string) error {
	pkg, err := builder.CreatePackage(outputPackage)
	if err != nil {
		return err
	}

	server := &serverArgs{
		Package:   pkg,
		Impl:      iface,
		IfaceName: iface.Name,
		Imports:   gogen.NewImports(pkg.Name),
	}

	server.Imports.AddPackages("context")

	slog.Info(fmt.Sprintf("Generating %v/%v", server.Package.PackageName, iface.Name))
	outputFile := filepath.Join(server.Package.Path, iface.Name+".go")
	return gogen.ExecuteTemplateToFile("XTrace", clientTemplate, server, outputFile)
}

var serverTemplate = `// Blueprint: Auto-generated by XTrace Plugin
package {{.Package.ShortName}}

{{.Imports}}

type {{.IfaceName}} interface {
	{{range $_, $f := .Impl.Methods -}}
	{{Signature $f}}
	{{end}}
}

type {{.Name}} struct {
	Service {{.Imports.NameOf .Service.UserType}}
	XTClient {{.Imports.NameOf .XTIface.UserType}}
}

func New_{{.Name}}(ctx context.Context, service {{.Imports.NameOf .Service.UserType}}, xtclient {{.Imports.NameOf .XTIface.UserType}}) (*{{.Name}}, error) {
	handler := &{{.Name}}{}
	handler.Service = service
	handler.XTClient = xtclient
	return handler, nil
}

{{$service := .Service.Name -}}
{{$receiver := .Name -}}
{{range $_, $f := .Service.Methods}}
func (handler *{{$receiver}}) {{$f.Name -}} ({{ArgVarsAndTypes $f "ctx context.Context"}}, baggage string) ({{RetVarsAndTypes $f "ret_baggage string" "err error"}}) {
	if baggage != "" {
		remote_baggage, _ := tracingplane.DecodeBase64(baggage)
		ctx, _ = handler.XTClient.Set(ctx, remote_baggage)
	}

	if res, _ := handler.XTClient.IsTracing(ctx); !res {
		ctx, _ = handler.XTClient.StartTask(ctx, "{{$f.Name}}")
	}
	ctx, _ = handler.XTClient.Log(ctx, "{{$f.Name}} start")

	{{RetVars $f "err"}} = handler.Service.{{$f.Name}}({{ArgVars $f "ctx"}})
	if err != nil {
		ctx, _ = handler.XTClient.LogWithTags(ctx, err.Error(), "Error")
	}
	ctx, _ = handler.XTClient.Log(ctx, "{{$f.Name}} end")
	ret_baggage_raw, _ := handler.XTClient.Get(ctx)
	ret_baggage = tracingplane.EncodeBase64(ret_baggage_raw)
	return
}
{{end}}
`
var clientTemplate = `// Blueprint: Auto-generated by XTrace plugin
package {{.Package.ShortName}}

{{.Imports}}

type {{.IfaceName}} interface {
	{{range $_, $f := .Impl.Methods -}}
	{{Signature $f}}
	{{end}}
}
`
