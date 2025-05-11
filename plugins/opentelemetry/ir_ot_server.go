package opentelemetry

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

// Blueprint IR Node that wraps the server-side of a service to generate ot compatible logs
type OpenTelemetryServerWrapper struct {
	golang.Service
	golang.Instantiable
	golang.GeneratesFuncs

	WrapperName   string
	outputPackage string
	Wrapped       golang.Service
	Collector     OpenTelemetryCollectorInterface
}

func newOpenTelemetryServerWrapper(name string, server ir.IRNode, collector ir.IRNode) (*OpenTelemetryServerWrapper, error) {
	serverNode, is_callable := server.(golang.Service)
	if !is_callable {
		return nil, blueprint.Errorf("opentelemetry server wrapper requires %s to be a golang service but got %s", server.Name(), reflect.TypeOf(server).String())
	}

	collectorClient, is_collector_client := collector.(OpenTelemetryCollectorInterface)
	if !is_collector_client {
		return nil, blueprint.Errorf("opentelemetry server wrapper requires %s to be an opentelemetry collector client", collector.Name())
	}

	node := &OpenTelemetryServerWrapper{}
	node.WrapperName = name
	node.Wrapped = serverNode
	node.Collector = collectorClient
	node.outputPackage = "ot"
	return node, nil
}

func (node *OpenTelemetryServerWrapper) Name() string {
	return node.WrapperName
}

func (node *OpenTelemetryServerWrapper) String() string {
	return node.Name() + " = OTServerWrapper(" + node.Wrapped.Name() + ", " + node.Collector.Name() + ")"
}

func (node *OpenTelemetryServerWrapper) genInterface(ctx ir.BuildContext) (*gocode.ServiceInterface, error) {
	iface, err := golang.GetGoInterface(ctx, node.Wrapped)
	if err != nil {
		return nil, err
	}
	module_ctx, valid := ctx.(golang.ModuleBuilder)
	if !valid {
		return nil, blueprint.Errorf("OTServerWrapper expected build context to be a ModuleBuilder, got %v", ctx)
	}
	i := gocode.CopyServiceInterface(fmt.Sprintf("%v_OTServerWrapperInterface", iface.BaseName), module_ctx.Info().Name+"/"+node.outputPackage, iface)
	for name, method := range i.Methods {
		method.AddArgument(gocode.Variable{Name: "traceCtx", Type: &gocode.BasicType{Name: "string"}})
		i.Methods[name] = method
	}
	return i, nil
}

func (node *OpenTelemetryServerWrapper) GetInterface(ctx ir.BuildContext) (service.ServiceInterface, error) {
	return node.genInterface(ctx)
}

func (node *OpenTelemetryServerWrapper) ImplementsGolangNode()    {}
func (node *OpenTelemetryServerWrapper) ImplementsGolangService() {}

// Part of code generation compilation pass; creates the interface definition code for the wrapper,
// and any new generated structs that are exposed and can be used by other IRNodes
func (node *OpenTelemetryServerWrapper) AddInterfaces(builder golang.ModuleBuilder) error {
	iface, err := node.genInterface(builder)
	if err != nil {
		return err
	}

	// Only generate code once
	if builder.Visited(iface.Name + ".ot_server_iface") {
		return nil
	}

	err = generateClientSideInterfaces(builder, iface, node.outputPackage)
	if err != nil {
		return err
	}

	return node.Wrapped.AddInterfaces(builder)
}

// Part of code generation compilation pass; provides implementation of interfaces from GenerateInterfaces
func (node *OpenTelemetryServerWrapper) GenerateFuncs(builder golang.ModuleBuilder) error {
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
	if builder.Visited(impl_iface.Name + ".ot_server_impl") {
		return nil
	}

	return generateServerHandler(builder, wrapped_iface, impl_iface, coll_iface, node.outputPackage)
}

// Part of code generation compilation pass; provides instantiation snippet
func (node *OpenTelemetryServerWrapper) AddInstantiation(builder golang.NamespaceBuilder) error {
	// Only generate instantiation code for this instance once
	if builder.Visited(node.WrapperName) {
		return nil
	}

	iface, err := golang.GetGoInterface(builder, node.Wrapped)
	if err != nil {
		return err
	}

	collector_iface, err := golang.GetGoInterface(builder, node.Collector)
	if err != nil {
		return err
	}

	constructor := &gocode.Constructor{
		Package: builder.Module().Info().Name + "/" + node.outputPackage,
		Func: gocode.Func{
			Name: fmt.Sprintf("New_%v_OTServerWrapper", iface.BaseName),
			Arguments: []gocode.Variable{
				{Name: "ctx", Type: &gocode.UserType{Package: "context", Name: "Context"}},
				{Name: "service", Type: iface},
				{Name: "otCollectorClient", Type: collector_iface},
			},
		},
	}

	return builder.DeclareConstructor(node.WrapperName, constructor, []ir.IRNode{node.Wrapped, node.Collector})
}

type serverArgs struct {
	Package   golang.PackageInfo
	Service   *gocode.ServiceInterface
	Impl      *gocode.ServiceInterface
	CollIface *gocode.ServiceInterface
	Name      string
	IfaceName string
	Imports   *gogen.Imports
}

func generateServerHandler(builder golang.ModuleBuilder, wrapped *gocode.ServiceInterface, impl *gocode.ServiceInterface, coll_iface *gocode.ServiceInterface, outputPackage string) error {
	pkg, err := builder.CreatePackage(outputPackage)
	if err != nil {
		return err
	}

	server := &serverArgs{
		Package:   pkg,
		Service:   wrapped,
		Impl:      impl,
		CollIface: coll_iface,
		Name:      wrapped.BaseName + "_OTServerWrapper",
		IfaceName: impl.Name,
		Imports:   gogen.NewImports(pkg.Name),
	}

	env := &envArgs{
		Name: wrapped.BaseName,
	}

	server.Imports.AddPackages("context", "go.opentelemetry.io/otel/trace", "github.com/blueprint-uservices/blueprint/runtime/core/backend")

	slog.Info(fmt.Sprintf("Generating %v/%v", server.Package.PackageName, "env.sh"))
	outputFile := filepath.Join(server.Package.Path, "env.sh")
	err = gogen.ExecuteTemplateToFile("OTServerWrapper", envTemplate, env, outputFile)
	if err != nil {
		return err
	}

	slog.Info(fmt.Sprintf("Generating %v/%v", server.Package.PackageName, impl.Name))
	outputFile = filepath.Join(server.Package.Path, impl.Name+".go")
	return gogen.ExecuteTemplateToFile("OTServerWrapper", serverTemplate, server, outputFile)
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
	return gogen.ExecuteTemplateToFile("OTServerWrapper", clientTemplate, server, outputFile)
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
	CollClient {{.Imports.NameOf .CollIface.UserType}}
}

func New_{{.Name}}(ctx context.Context, service {{.Imports.NameOf .Service.UserType}}, coll_client {{.Imports.NameOf .CollIface.UserType}}) (*{{.Name}}, error) {
	handler := &{{.Name}}{}
	handler.Service = service
	handler.CollClient = coll_client
	return handler, nil
}

{{$service := .Service.Name -}}
{{$receiver := .Name -}}
{{range $_, $f := .Service.Methods}}
func (handler *{{$receiver}}) {{$f.Name -}} ({{ArgVarsAndTypes $f "ctx context.Context"}}, traceCtx string) ({{RetVarsAndTypes $f "err error"}}) {
	if traceCtx != "" {
		span_ctx_config, _ := backend.GetSpanContext(traceCtx)
		span_ctx := trace.NewSpanContext(span_ctx_config)
		ctx = trace.ContextWithRemoteSpanContext(ctx, span_ctx)
	}

	tp, _ := handler.CollClient.GetTracerProvider(ctx)
	tr := tp.Tracer("{{$service}}")
	ctx, span := tr.Start(ctx, "{{$service}}Server_{{$f.Name}}")
	defer span.End()
	{{RetVars $f "err"}} = handler.Service.{{$f.Name}}({{ArgVars $f "ctx"}})
	if err != nil {
		span.RecordError(err)
	}
	return
}
{{end}}
`

var clientTemplate = `// Blueprint: Auto-generated by OT plugin
package {{.Package.ShortName}}

{{.Imports}}

type {{.IfaceName}} interface {
	{{range $_, $f := .Impl.Methods -}}
	{{Signature $f}}
	{{end}}
}
`

type envArgs struct {
	Name string
}

var envTemplate = `#!/bin/bash 
# Auto-generated by OT plugin

export OTEL_SERVICE_NAME="{{.Name}}"
`
