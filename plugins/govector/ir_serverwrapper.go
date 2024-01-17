package govector

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

// Blueprint IR node that wraps the server-side of a service to generate govec compatible logs
type GovecServerWrapper struct {
	golang.Service
	golang.Instantiable
	golang.GeneratesFuncs

	InstanceName  string
	outputPackage string
	Wrapped       golang.Service
}

// Implements ir.IRNode
func (node *GovecServerWrapper) Name() string {
	return node.InstanceName
}

// Implements ir.IRNode
func (node *GovecServerWrapper) String() string {
	return node.Name() + " = GovecServerWrapper(" + node.Wrapped.Name() + ")"
}

// Implements golang.Node
func (node *GovecServerWrapper) ImplementsGolangNode() {}

// Implements golang.Service
func (node *GovecServerWrapper) ImplementsGolangService() {}

func newGovecServerWrapper(name string, wrapped golang.Service) (*GovecServerWrapper, error) {
	node := &GovecServerWrapper{}
	node.InstanceName = name
	node.outputPackage = "govec"
	node.Wrapped = wrapped
	return node, nil
}

func (node *GovecServerWrapper) genInterface(ctx ir.BuildContext) (*gocode.ServiceInterface, error) {
	iface, err := golang.GetGoInterface(ctx, node.Wrapped)
	if err != nil {
		return nil, err
	}

	module_ctx, valid := ctx.(golang.ModuleBuilder)
	if !valid {
		return nil, blueprint.Errorf("GoVecServerWrapper expected build context to be a ModuleBuilder, got %v", ctx)
	}
	i := gocode.CopyServiceInterface(fmt.Sprintf("%v_GoVecServerWrapperInterface", iface.BaseName), module_ctx.Info().Name+"/"+node.outputPackage, iface)
	for name, method := range i.Methods {
		method.AddArgument(gocode.Variable{Name: "govecctx", Type: &gocode.Slice{SliceOf: &gocode.BasicType{Name: "byte"}}})
		method.AddRetVar(gocode.Variable{Name: "", Type: &gocode.Slice{SliceOf: &gocode.BasicType{Name: "byte"}}})
		i.Methods[name] = method
	}
	return i, nil
}

// Implements golang.Instantiable
func (node *GovecServerWrapper) AddInstantiation(builder golang.NamespaceBuilder) error {
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
			Name: fmt.Sprintf("New_%v_GoVecServerWrapper", iface.BaseName),
			Arguments: []gocode.Variable{
				{Name: "ctx", Type: &gocode.UserType{Package: "context", Name: "Context"}},
				{Name: "service", Type: iface},
			},
		},
	}
	return builder.DeclareConstructor(node.InstanceName, constructor, []ir.IRNode{node.Wrapped})
}

// Implements golang.GeneratesFuncs
func (node *GovecServerWrapper) GenerateFuncs(builder golang.ModuleBuilder) error {
	wrapped_iface, err := golang.GetGoInterface(builder, node.Wrapped)
	if err != nil {
		return err
	}

	impl_iface, err := node.genInterface(builder)
	if err != nil {
		return err
	}

	return generateServerHandler(builder, wrapped_iface, impl_iface, node.outputPackage)
}

// Implements service.ServiceNode
func (node *GovecServerWrapper) GetInterface(ctx ir.BuildContext) (service.ServiceInterface, error) {
	return node.genInterface(ctx)
}

// Implements golang.PRovidesInterface
func (node *GovecServerWrapper) AddInterfaces(builder golang.ModuleBuilder) error {
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

func generateServerHandler(builder golang.ModuleBuilder, wrapped *gocode.ServiceInterface, impl *gocode.ServiceInterface, outputPackage string) error {
	pkg, err := builder.CreatePackage(outputPackage)
	if err != nil {
		return err
	}

	server := &serverArgs{
		Package:   pkg,
		Service:   wrapped,
		Impl:      impl,
		Name:      wrapped.BaseName + "_GoVecServerWrapper",
		IfaceName: impl.Name,
		Imports:   gogen.NewImports(pkg.Name),
	}

	server.Imports.AddPackages("context", "github.com/blueprint-uservices/blueprint/runtime/plugins/govector", "errors")

	slog.Info(fmt.Sprintf("Generating %v/%v", server.Package.PackageName, impl.Name))
	outputFile := filepath.Join(server.Package.Path, impl.Name+".go")
	return gogen.ExecuteTemplateToFile("GoVector", serverTemplate, server, outputFile)
}

type serverArgs struct {
	Package   golang.PackageInfo
	Service   *gocode.ServiceInterface
	Impl      *gocode.ServiceInterface
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
	return gogen.ExecuteTemplateToFile("GoVector", clientInterfaceTemplate, server, outputFile)
}

var serverTemplate = `// Blueprint: Auto-generated by GoVector Plugin
package {{.Package.ShortName}}

{{.Imports}}

type {{.IfaceName}} interface {
	{{range $_, $f := .Impl.Methods -}}
	{{Signature $f}}
	{{end}}
}

type {{.Name}} struct {
	Service {{.Imports.NameOf .Service.UserType}}
	logger *govector.GoVecLogger
}

func New_{{.Name}} (ctx context.Context, service {{.Imports.NameOf .Service.UserType}}) (*{{.Name}}, error) {
	handler := &{{.Name}}{}
	handler.Service = service
	handler.logger = govector.GetLogger()
	if handler.logger == nil {
		return nil, errors.New("GoVector logger is not initialized")
	}
	return handler, nil
}

{{$service := .Service.Name -}}
{{$receiver := .Name -}}
{{range $_, $f := .Service.Methods}}
func (handler *{{$receiver}}) {{$f.Name -}} ({{ArgVarsAndTypes $f "ctx context.Context"}}, govecctx []byte) ({{RetVarsAndTypes $f "govecret []byte" "err error"}}) {
	handler.logger.UnpackReceiveCtx(ctx, "Unpacking go vec context from client request", govecctx)
	{{RetVars $f "err"}} = handler.Service.{{$f.Name}}({{ArgVars $f "ctx"}})
	govecret, _ = handler.logger.GetSendCtx(ctx, "Sending response to the client")
	return
}
{{end}}
`

var clientInterfaceTemplate = `// Blueprint: Auto-generated by GoVector plugin
package {{.Package.ShortName}}

{{.Imports}}

type {{.IfaceName}} interface {
	{{range $_, $f := .Impl.Methods -}}
	{{Signature $f}}
	{{end}}
}
`
