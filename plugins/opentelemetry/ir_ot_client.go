package opentelemetry

import (
	"fmt"
	"path/filepath"

	"github.com/blueprint-uservices/blueprint/blueprint/pkg/blueprint"
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/coreplugins/service"
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/ir"
	"github.com/blueprint-uservices/blueprint/plugins/golang"
	"github.com/blueprint-uservices/blueprint/plugins/golang/gocode"
	"github.com/blueprint-uservices/blueprint/plugins/golang/gogen"
	"golang.org/x/exp/slog"
)

// Blueprint IR Node that wraps the client-side of a service to generate ot compatible logs
type OpenTelemetryClientWrapper struct {
	golang.Service
	golang.GeneratesFuncs

	WrapperName   string
	outputPackage string
	Wrapped       golang.Service
	Collector     OpenTelemetryCollectorInterface
}

func newOpenTelemetryClientWrapper(name string, server golang.Service, collector OpenTelemetryCollectorInterface) (*OpenTelemetryClientWrapper, error) {
	node := &OpenTelemetryClientWrapper{}
	node.WrapperName = name
	node.Wrapped = server
	node.Collector = collector
	node.outputPackage = "ot"
	return node, nil
}

func (node *OpenTelemetryClientWrapper) Name() string {
	return node.WrapperName
}

func (node *OpenTelemetryClientWrapper) String() string {
	return node.Name() + " = OTClientWrapper(" + node.Wrapped.Name() + ", " + node.Collector.Name() + ")"
}

func (node *OpenTelemetryClientWrapper) genInterface(ctx ir.BuildContext) (*gocode.ServiceInterface, error) {
	iface, err := golang.GetGoInterface(ctx, node.Wrapped)
	if err != nil {
		return nil, err
	}
	module_ctx, valid := ctx.(golang.ModuleBuilder)
	if !valid {
		return nil, blueprint.Errorf("OTClientWrapper expected build context to be a ModuleBuiler, got %v", ctx)
	}
	i := gocode.CopyServiceInterface(fmt.Sprintf("%v_OTClientWrapperInterface", iface.BaseName), module_ctx.Info().Name+"/"+node.outputPackage, iface)
	for name, method := range i.Methods {
		method.Arguments = method.Arguments[:len(method.Arguments)-1]
		i.Methods[name] = method
	}
	return i, nil
}

func (node *OpenTelemetryClientWrapper) GetInterface(ctx ir.BuildContext) (service.ServiceInterface, error) {
	return node.genInterface(ctx)
}

// Part of code generation compilation pass; creates the interface definition code for the wrapper,
// and any new generated structs that are exposed and can be used by other IRNodes
func (node *OpenTelemetryClientWrapper) AddInterfaces(builder golang.ModuleBuilder) error {
	return node.Wrapped.AddInterfaces(builder)
}

// Part of code generation compilation pass; provides implementation of interfaces from GenerateInterfaces
func (node *OpenTelemetryClientWrapper) GenerateFuncs(builder golang.ModuleBuilder) error {
	builder.Require("go.opentelemetry.io/otel/trace", "v1.26.0")
	wrapped_iface, err := golang.GetGoInterface(builder, node.Wrapped)
	if err != nil {
		return err
	}

	coll_iface, err := golang.GetGoInterface(builder, node.Collector)
	if err != nil {
		return err
	}

	impl_iface, err := node.genInterface(builder)
	if err != nil {
		return err
	}

	// Only generate code once
	if builder.Visited(impl_iface.Name + ".ot_client_impl") {
		return nil
	}

	return generateClientHandler(builder, wrapped_iface, impl_iface, coll_iface, node.outputPackage)
}

// Part of code generation compilation pass; provides instantiation snippet
func (node *OpenTelemetryClientWrapper) AddInstantiation(builder golang.NamespaceBuilder) error {
	// Only generate instantiation code for this instance once
	if builder.Visited(node.WrapperName) {
		return nil
	}

	iface, err := golang.GetGoInterface(builder, node.Wrapped)
	if err != nil {
		return err
	}

	coll_iface, err := golang.GetGoInterface(builder, node.Collector)
	if err != nil {
		return err
	}

	constructor := &gocode.Constructor{
		Package: builder.Module().Info().Name + "/" + node.outputPackage,
		Func: gocode.Func{
			Name: fmt.Sprintf("New_%v_OTClientWrapper", iface.BaseName),
			Arguments: []gocode.Variable{
				{Name: "ctx", Type: &gocode.UserType{Package: "context", Name: "Context"}},
				{Name: "client", Type: iface},
				{Name: "coll_client", Type: coll_iface},
			},
		},
	}

	return builder.DeclareConstructor(node.WrapperName, constructor, []ir.IRNode{node.Wrapped, node.Collector})
}

func (node *OpenTelemetryClientWrapper) ImplementsGolangNode()    {}
func (node *OpenTelemetryClientWrapper) ImplementsGolangService() {}

func generateClientHandler(builder golang.ModuleBuilder, wrapped *gocode.ServiceInterface, impl *gocode.ServiceInterface, coll_iface *gocode.ServiceInterface, outputPackage string) error {
	pkg, err := builder.CreatePackage(outputPackage)
	if err != nil {
		return err
	}

	server := &clientArgs{
		Package:         pkg,
		Service:         wrapped,
		Impl:            impl,
		CollIface:       coll_iface,
		Name:            wrapped.BaseName + "_OTClientWrapper",
		IfaceName:       impl.Name,
		ServerIfaceName: wrapped.BaseName + "_OTServerWrapperInterface",
		Imports:         gogen.NewImports(pkg.Name),
	}

	server.Imports.AddPackages("context")

	slog.Info(fmt.Sprintf("Generating %v/%v", server.Package.PackageName, impl.Name))
	outputFile := filepath.Join(server.Package.Path, impl.Name+".go")
	return gogen.ExecuteTemplateToFile("OTClientWrapper", clientSideTemplate, server, outputFile)
}

type clientArgs struct {
	Package         golang.PackageInfo
	Service         *gocode.ServiceInterface
	Impl            *gocode.ServiceInterface
	CollIface       *gocode.ServiceInterface
	Name            string
	IfaceName       string
	ServerIfaceName string
	Imports         *gogen.Imports
}

var clientSideTemplate = `// Blueprint: Auto-generated by OT Plugin
package {{.Package.ShortName}}

{{.Imports}}

type {{.IfaceName}} interface {
	{{range $_, $f := .Impl.Methods -}}
	{{Signature $f}}
	{{end}}
}

type {{.Name}} struct {
	Client {{.ServerIfaceName}}
	CollClient {{.Imports.NameOf .CollIface.UserType}}
}

func New_{{.Name}}(ctx context.Context, client {{.ServerIfaceName}}, coll_client {{.Imports.NameOf .CollIface.UserType}}) (*{{.Name}}, error) {
	handler := &{{.Name}}{}
	handler.Client = client
	handler.CollClient = coll_client
	return handler, nil
}

{{$service := .Service.Name -}}
{{$basename := .Service.BaseName -}}
{{$receiver := .Name -}}
{{range $_, $f := .Impl.Methods}}
func (handler *{{$receiver}}) {{$f.Name -}} ({{ArgVarsAndTypes $f "ctx context.Context"}}) ({{RetVarsAndTypes $f "err error"}}) {
	tp, _ := handler.CollClient.GetTracerProvider(ctx)
	tr := tp.Tracer("{{$service}}")
	ctx, span := tr.Start(ctx, "{{$basename}}Client_{{$f.Name}}")
	defer span.End()
	trace_ctx, _ := span.SpanContext().MarshalJSON()
	{{RetVars $f "err"}} = handler.Client.{{$f.Name}}({{ArgVars $f "ctx"}}, string(trace_ctx))
	if err != nil {
		span.RecordError(err)
	}
	return
}
{{end}}
`
