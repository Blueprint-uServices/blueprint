# 📝 Developing a Blueprint Plugin

This doc describes how to develop a blueprint plugin. This covers the following topics in detail:
  * Breakdown of the core components of a plugin
  * Tutorial on implementing a plugin for instrumenting existing methods of services in an application workflow
  * Tutorial on adding new methods to a service in an application workflow
  * Tutorial on changing the signatures (adding parameters, return values) of existing methods
  * Advanced Concepts when developing Plugins

All the code listed in this tutorial is available at [tutorial](github.com/blueprint-uservices/tutorial) unless otherwise stated.

## Plugin Components

A plugin can be composed of many components. Here are the components:
* IR Nodes: These are the IR nodes that the plugin provides. Usually placed in files called `ir_*.go`. 
* Wiring funcs: These are the functions made available by the plugin which can be called by wiring specifications to add new IR nodes. Usually placed in a file called `wiring.go`.
* (Optional) Code generation: Some plugins might also optionally generate code. The generation code could be placed alongside the IR Node definitions or could be placed in a separate `codegen` sub-package.
* (Optional) Runtime component: Some plugins might introduce a runtime component. The runtime component must be placed in a separate runtime module.

## Tutorial: Instrumenting Services

This section provides a tutorial for implementing a plugin that instruments the server and client side of a service with logging statements.

### Wiring Specification

The following implementation provides a function that can be called from the wiring spec to instrument the server and client side of the desired service.

```go
func Instrument(spec wiring.WiringSpec, serviceName string) {
	// Define the names for the wrapper nodes we are adding to the Blueprint IR
	wrapper_name := serviceName + ".hello.instrument.server"
	client_wrapper_name := serviceName + ".hello.instrument.client"

	// Get the pointer for the serviceName to ensure that the newly defined wrapper IR node will be attached to the node chain of the desired service
	ptr := pointer.GetPointer(spec, serviceName)
	if ptr == nil {
		slog.Error("Unable to add instrument " + serviceName + " as it is not a pointer. Did you forget to define " + serviceName + "? You can define serviceName using `workflow.Service`")
		return
	}

	// Attach the Hello wrapper node to the server-side node chain of the desired service
	serverNext := ptr.AddDstModifier(spec, wrapper_name)

	// Define the IRNode for the wrapper node and add it to the wiring specification
	spec.Define(wrapper_name, &HelloInstrumentServerWrapper{}, func(ns wiring.Namespace) (ir.IRNode, error) {
		// Get the IRNode that will be wrapped by HelloWrapper
		var server golang.Service
		if err := ns.Get(serverNext, &server); err != nil {
			return nil, blueprint.Errorf("Tutorial Plugin %s expected %s to be a golang.Service, but encountered %s", wrapper_name, serverNext, err)
		}

		// Instantiate the IRNode
		return newHelloInstrumentServerWrapper(wrapper_name, server)
	})

	// Attach the Hello wrapper node to the client-side node chain of the desired service
	clientNext := ptr.AddSrcModifier(spec, client_wrapper_name)

	// Define the IRNode for the wrapper node and add it to the wiring specification
	spec.Define(client_wrapper_name, &HelloInstrumentClientWrapper{}, func(ns wiring.Namespace) (ir.IRNode, error) {
		// Get the IRNode that will be wrapped by HelloWrapper
		var client golang.Service
		if err := ns.Get(clientNext, &client); err != nil {
			return nil, blueprint.Errorf("Tutorial Plugin %s expected %s to be a golang.Service, but encountered %s", wrapper_name, serverNext, err)
		}

		return newHelloInstrumentClientWrapper(client_wrapper_name, client)
	})
}
```

### IR Nodes

We need to define two different IR Nodes: (i) an IR Node that wraps and instruments the server-side of the desired service; and (ii) an IR Node that wraps and instruments the client-side of the desired service.

Following is the implementation of the server-side IR Node.

```go
// Blueprint IRNode for representing the wrapper node that instruments every server-side method in the node that gets wrapped
type HelloInstrumentServerWrapper struct {
	golang.Service
	golang.GeneratesFuncs
	golang.Instantiable

	InstanceName string
	Wrapped      golang.Service

	outputPackage string
}

// Implements ir.IRNode
func (node *HelloInstrumentServerWrapper) ImplementsGolangNode() {}

// Implements ir.IRNode
func (node *HelloInstrumentServerWrapper) Name() string {
	return node.InstanceName
}

// Implements ir.IRNode
func (node *HelloInstrumentServerWrapper) String() string {
	return node.Name() + " = HelloInstrumentServerWrapper(" + node.Wrapped.Name() + ")"
}

// Implements golang.ProvidesInterface
func (node *HelloInstrumentServerWrapper) AddInterfaces(builder golang.ModuleBuilder) error {
	return node.Wrapped.AddInterfaces(builder)
}

func newHelloInstrumentServerWrapper(name string, server ir.IRNode) (*HelloInstrumentServerWrapper, error) {
	serverNode, ok := server.(golang.Service)
	if !ok {
		return nil, blueprint.Errorf("tutorial server wrapper requires %s to be a golang service but got %s", server.Name(), reflect.TypeOf(server).String())
	}

	node := &HelloInstrumentServerWrapper{}
	node.InstanceName = name
	node.Wrapped = serverNode
	node.outputPackage = "tutorial"

	return node, nil
}

// Implements service.ServiceNode
func (node *HelloInstrumentServerWrapper) GetInterface(ctx ir.BuildContext) (service.ServiceInterface, error) {
	return node.Wrapped.GetInterface(ctx)
}

// Implements golang.GeneratesFuncs
func (node *HelloInstrumentServerWrapper) GenerateFuncs(builder golang.ModuleBuilder) error {
	iface, err := golang.GetGoInterface(builder, node)
	if err != nil {
		return err
	}
	return generateServerInstrumentHandler(builder, iface, node.outputPackage)
}

// Implements golang.Instantiable
func (node *HelloInstrumentServerWrapper) AddInstantiation(builder golang.NamespaceBuilder) error {
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
			Name: fmt.Sprintf("New_%v_TutorialInstrumentServerWrapper", iface.BaseName),
			Arguments: []gocode.Variable{
				{Name: "ctx", Type: &gocode.UserType{Package: "context", Name: "Context"}},
				{Name: "service", Type: iface},
			},
		},
	}

	return builder.DeclareConstructor(node.InstanceName, constructor, []ir.IRNode{node.Wrapped})
}
```

Following is the implementation of the client-side IR Node.

```go
// Blueprint IRNode for representing the wrapper node that instruments every client-side method in the node that gets wrapped
type HelloInstrumentClientWrapper struct {
	golang.Service
	golang.GeneratesFuncs
	golang.Instantiable

	InstanceName string
	Wrapped      golang.Service

	outputPackage string
}

// Implements ir.IRNode
func (node *HelloInstrumentClientWrapper) ImplementsGolangNode() {}

// Implements ir.IRNode
func (node *HelloInstrumentClientWrapper) Name() string {
	return node.InstanceName
}

// Implements ir.IRNode
func (node *HelloInstrumentClientWrapper) String() string {
	return node.Name() + " = HelloInstrumentClientWrapper(" + node.Wrapped.Name() + ")"
}

// Implements golang.ProvidesInterface
func (node *HelloInstrumentClientWrapper) AddInterfaces(builder golang.ModuleBuilder) error {
	return node.Wrapped.AddInterfaces(builder)
}

func newHelloInstrumentClientWrapper(name string, wrapped ir.IRNode) (*HelloInstrumentClientWrapper, error) {
	serverNode, ok := wrapped.(golang.Service)
	if !ok {
		return nil, blueprint.Errorf("tutorial server wrapper requires %s to be a golang service but got %s", wrapped.Name(), reflect.TypeOf(wrapped).String())
	}

	node := &HelloInstrumentClientWrapper{}
	node.InstanceName = name
	node.Wrapped = serverNode
	node.outputPackage = "tutorial"

	return node, nil
}

// Implements service.ServiceNode
func (node *HelloInstrumentClientWrapper) GetInterface(ctx ir.BuildContext) (service.ServiceInterface, error) {
	return node.Wrapped.GetInterface(ctx)
}

// Implements golang.GeneratesFuncs
func (node *HelloInstrumentClientWrapper) GenerateFuncs(builder golang.ModuleBuilder) error {
	iface, err := golang.GetGoInterface(builder, node)
	if err != nil {
		return err
	}
	return generateClientInstrumentHandler(builder, iface, node.outputPackage)
}

// Implements golang.Instantiable
func (node *HelloInstrumentClientWrapper) AddInstantiation(builder golang.NamespaceBuilder) error {
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
			Name: fmt.Sprintf("New_%v_TutorialInstrumentClientWrapper", iface.BaseName),
			Arguments: []gocode.Variable{
				{Name: "ctx", Type: &gocode.UserType{Package: "context", Name: "Context"}},
				{Name: "service", Type: iface},
			},
		},
	}

	return builder.DeclareConstructor(node.InstanceName, constructor, []ir.IRNode{node.Wrapped})
}
```

### Code Generation

The plugin requires code-generation on both the server-side and client-side of the service.

To generate the code, we first define a code generation struct that can be used by the `gogen` plugin to generate source code to specific files.
For more information on templated code-generation in Blueprint, refer to the [gogen](https://github.com/Blueprint-uServices/blueprint/tree/main/plugins/golang/gogen) plugin.

Following is the implementation of the code generation struct.

```go
type serverArgs struct {
	Package   golang.PackageInfo
	Service   *gocode.ServiceInterface
	Iface     *gocode.ServiceInterface
	Name      string
	IfaceName string
	Imports   *gogen.Imports
  ServerIfaceName string
}
```

We then define a method that generates the server-side source code using the `gogen` plugin and the previously defined `serverArgs` code generation struct.

```go
func generateServerInstrumentHandler(builder golang.ModuleBuilder, wrapped *gocode.ServiceInterface, outputPackage string) error {
	pkg, err := builder.CreatePackage(outputPackage)
	if err != nil {
		return err
	}

	server := &serverArgs{
		Package:   pkg,
		Service:   wrapped,
		Iface:     wrapped,
		Name:      wrapped.BaseName + "_TutorialInstrumentServerWrapper",
		IfaceName: wrapped.Name,
		Imports:   gogen.NewImports(pkg.Name),
	}

	server.Imports.AddPackages("context", "log")

	slog.Info(fmt.Sprintf("Generating %v/%v", server.Package.PackageName, wrapped.BaseName+"_TutorialInstrumentServerWrapper"))
	outputFile := filepath.Join(server.Package.Path, wrapped.BaseName+"_TutorialInstrumentServerWrapper.go")
	return gogen.ExecuteTemplateToFile("Tutorial", serverInstrumentTemplate, server, outputFile)
}

var serverInstrumentTemplate = `// Blueprint: Auto-generated by Tutorial Plugin
package {{.Package.ShortName}}

{{.Imports}}

type {{.Name}} struct {
	Service {{.Imports.NameOf .Service.UserType}}
}

func New_{{.Name}}(ctx context.Context, service {{.Imports.NameOf .Service.UserType}}) (*{{.Name}}, error) {
	handler := &{{.Name}}{}
	handler.Service = service
	return handler, nil
}

{{$service := .Service.Name -}}
{{$receiver := .Name -}}
{{ range $_, $f := .Service.Methods }}
func (handler *{{$receiver}}) {{$f.Name -}} ({{ArgVarsAndTypes $f "ctx context.Context"}}) ({{RetTypes $f "error"}}) {
	log.Println("Processing {{$f.Name}}")
	return handler.Service.{{$f.Name}}({{ArgVars $f "ctx"}})
}
{{end}}
`
```

We then define a method that generates the client-side source code using the `gogen` plugin and the previously defined `serverArgs` code generation struct.

```go
func generateClientInstrumentHandler(builder golang.ModuleBuilder, wrapped *gocode.ServiceInterface, outputPackage string) error {
	pkg, err := builder.CreatePackage(outputPackage)
	if err != nil {
		return err
	}

	server := &serverArgs{
		Package:   pkg,
		Service:   wrapped,
		Iface:     wrapped,
		Name:      wrapped.BaseName + "_TutorialInstrumentClientWrapper",method

type {{.Name}} struct {
	Client {{.Imports.NameOf .Service.UserType}}
}

func New_{{.Name}}(ctx context.Context, client {{.Imports.NameOf .Service.UserType}}) (*{{.Name}}, error) {
	handler := &{{.Name}}{}
	handler.Client = client
	return handler, nil
}

{{$service := .Service.Name -}}
{{$receiver := .Name -}}
{{ range $_, $f := .Service.Methods }}
func (handler *{{$receiver}}) {{$f.Name -}} ({{ArgVarsAndTypes $f "ctx context.Context"}}) ({{RetTypes $f "error"}}) {
	log.Println("Processing {{$f.Name}}")
	return handler.Client.{{$f.Name}}({{ArgVars $f "ctx"}})
}
{{end}}
`
```

## Tutorial: Adding new Methods

This section provides a tutorial for implementing a plugin that adds a new method to a service.

### Wiring Specification

The following implementation provides a function that can be called from the wiring spec to add a new method to the service.

```go
func AddHelloMethod(spec wiring.WiringSpec, serviceName string) {
	// Define the name for the wrapper node we are adding to the Blueprint IR
	wrapper_name := serviceName + ".hello.method"

	// Get the pointer for the serviceName to ensure that the newly defined wrapper IR node will be attached to the node chain of the desired service
	ptr := pointer.GetPointer(spec, serviceName)
	if ptr == nil {
		slog.Error("Unable to add hello method to " + serviceName + " as it is not a pointer. Did you forget to define " + serviceName + "? You can define serviceName using `workflow.Service`")
		return
	}

	// Attach the Hello wrapper node to the server-side node chain of the desired service
	serverNext := ptr.AddDstModifier(spec, wrapper_name)

	// Define the IRNode for the wrapper node and add it to the wiring specification
	spec.Define(wrapper_name, &HelloMethodWrapper{}, func(ns wiring.Namespace) (ir.IRNode, error) {
		// Get the IRNode that will be wrapped by HelloWrapper
		var server golang.Service
		if err := ns.Get(serverNext, &server); err != nil {
			return nil, blueprint.Errorf("Tutorial Plugin %s expected %s to be a golang.Service, but encountered %s", wrapper_name, serverNext, err)
		}

		// Instantiate the IRNode
		return newHelloMethodWrapper(wrapper_name, server)
	})
}
```

### IR Node

We need to define one IR Node: an IR Node that wraps and adds a new method to the server-side of the desired service.

The IRNode must now also generate a new interface as it has extended the service's existing interface with a new method.

```go
// Blueprint IRNode for representing the wrapper node that adds a `Hello` method to the wrapped IRNode.
type HelloMethodWrapper struct {
	golang.Service
	golang.GeneratesFuncs
	golang.Instantiable

	InstanceName string
	Wrapped      golang.Service

	outputPackage string
}

// Implements ir.IRNode
func (node *HelloMethodWrapper) ImplementsGolangNode() {}

// Implements ir.IRNode
func (node *HelloMethodWrapper) Name() string {
	return node.InstanceName
}

// Implements ir.IRNode
func (node *HelloMethodWrapper) String() string {
	return node.Name() + " = HelloMethodWrapper(" + node.Wrapped.Name() + ")"
}

// IMplements golang.ProvidesInterface
func (node *HelloMethodWrapper) AddInterfaces(builder golang.ModuleBuilder) error {
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

func newHelloMethodWrapper(name string, server ir.IRNode) (*HelloMethodWrapper, error) {
	serverNode, is_callable := server.(golang.Service)
	if !is_callable {
		return nil, blueprint.Errorf("tutorial server wrapper requires %s to be a golang service but got %s", server.Name(), reflect.TypeOf(server).String())
	}

	node := &HelloMethodWrapper{}
	node.InstanceName = name
	node.Wrapped = serverNode
	node.outputPackage = "tutorial_method"

	return node, nil
}

func (node *HelloMethodWrapper) genInterface(ctx ir.BuildContext) (*gocode.ServiceInterface, error) {
	iface, err := golang.GetGoInterface(ctx, node.Wrapped)
	if err != nil {
		return nil, err
	}
	module_ctx, valid := ctx.(golang.ModuleBuilder)
	if !valid {
		return nil, blueprint.Errorf("Tutorial expected build context to be a ModuleBuilder, got %v", ctx)
	}
	i := gocode.CopyServiceInterface(fmt.Sprintf("%v_TutorialMethod", iface.BaseName), module_ctx.Info().Name+"/"+node.outputPackage, iface)
	health_check_method := &gocode.Func{}
	health_check_method.Name = "HelloNew"
	health_check_method.Returns = append(health_check_method.Returns, gocode.Variable{Type: &gocode.BasicType{Name: "string"}})
	i.AddMethod(*health_check_method)
	return i, nil
}

// Implements service.ServiceNode
func (node *HelloMethodWrapper) GetInterface(ctx ir.BuildContext) (service.ServiceInterface, error) {
	return node.genInterface(ctx)
}

// Implements golang.GeneratesFuncs
func (node *HelloMethodWrapper) GenerateFuncs(builder golang.ModuleBuilder) error {
	service, err := golang.GetGoInterface(builder, node.Wrapped)
	if err != nil {
		return err
	}
	iface, err := golang.GetGoInterface(builder, node)
	if err != nil {
		return err
	}
	err = generateServerHandler(builder, iface, service, node.outputPackage)
	if err != nil {
		return err
	}
	return nil
}

// Implements golang.Instantiable
func (node *HelloMethodWrapper) AddInstantiation(builder golang.NamespaceBuilder) error {
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
			Name: fmt.Sprintf("New_%v_TutorialMethodImpl", iface.BaseName),
			Arguments: []gocode.Variable{
				{Name: "ctx", Type: &gocode.UserType{Package: "context", Name: "Context"}},
				{Name: "service", Type: iface},
			},
		},
	}

	return builder.DeclareConstructor(node.InstanceName, constructor, []ir.IRNode{node.Wrapped})
}
```

### Code Generation

The plugin requires code-generation on both the server-side and client-side of the service.

On the server-side, a wrapper is generated that adds a new method and its implementation.

```go
func generateServerHandler(builder golang.ModuleBuilder, iface *gocode.ServiceInterface, wrapped_service *gocode.ServiceInterface, outputPackage string) error {
	pkg, err := builder.CreatePackage(outputPackage)
	if err != nil {
		return err
	}

	server := &serverArgs{
		Package:   pkg,
		Service:   wrapped_service,
		Iface:     iface,
		Name:      wrapped_service.BaseName + "_TutorialMethodImpl",
		IfaceName: iface.Name,
		Imports:   gogen.NewImports(pkg.Name),
	}

	server.Imports.AddPackages("context")

	slog.Info(fmt.Sprintf("Generating %v/%v", server.Package.PackageName, iface.Name))
	outputFile := filepath.Join(server.Package.Path, iface.Name+".go")
	return gogen.ExecuteTemplateToFile("Tutorial", serverTemplate, server, outputFile)
}

var serverTemplate = `// Blueprint: Auto-generated by Tutorial Plugin
package {{.Package.ShortName}}

{{.Imports}}

type {{.IfaceName}} interface {
	{{range $_, $f := .Iface.Methods -}}
	{{Signature $f}}
	{{end}}
}

type {{.Name}} struct {
	Service {{.Imports.NameOf .Service.UserType}}
}

func New_{{.Name}}(ctx context.Context, service {{.Imports.NameOf .Service.UserType}}) (*{{.Name}}, error) {
	handler := &{{.Name}}{}
	handler.Service = service
	return handler, nil
}

{{$service := .Service.Name -}}
{{$receiver := .Name -}}
{{ range $_, $f := .Service.Methods }}
func (handler *{{$receiver}}) {{$f.Name -}} ({{ArgVarsAndTypes $f "ctx context.Context"}}) ({{RetTypes $f "error"}}) {
	return handler.Service.{{$f.Name}}({{ArgVars $f "ctx"}})
}
{{end}}
func (handler *{{$receiver}}) HelloNew(ctx context.Context) (string, error) {
	return "Hello!", nil
}
`
```

On the client-side, an interface is generated that can be used by the clients to correctly connect to the service.

```go
func generateClientSideInterfaces(builder golang.ModuleBuilder, iface *gocode.ServiceInterface, outputPackage string) error {
	pkg, err := builder.CreatePackage(outputPackage)
	if err != nil {
		return err
	}

	server := &serverArgs{
		Package:   pkg,
		Iface:     iface,
		IfaceName: iface.Name,
		Imports:   gogen.NewImports(pkg.Name),
	}

	server.Imports.AddPackages("context")

	slog.Info(fmt.Sprintf("Generating %v/%v", server.Package.PackageName, iface.Name))
	outputFile := filepath.Join(server.Package.Path, iface.Name+".go")
	return gogen.ExecuteTemplateToFile("Tutorial", clientTemplate, server, outputFile)
}

var clientTemplate = `// Blueprint: Auto-generated by Tutorial plugin
package {{.Package.ShortName}}

{{.Imports}}

type {{.IfaceName}} interface {
	{{range $_, $f := .Iface.Methods -}}
	{{Signature $f}}
	{{end}}
}
`
```

## Tutorial: Modifying Function Signatures

This section provides a tutorial for implementing a plugin that adds an extra call parameter and an extra return parameter to every method in the service.

### Wiring Specification

The following implementation provides a function that can be called from the wiring spec to extend the function signatures of every method exported by the service.

```go
func AddHelloParam(spec wiring.WiringSpec, serviceName string) {
	// Define the names for the wrapper nodes we are adding to the Blueprint IR
	wrapper_name := serviceName + ".hello.param.server"
	client_wrapper_name := serviceName + ".hello.param.client"

	// Get the pointer for the serviceName to ensure that the newly defined wrapper IR node will be attached to the node chain of the desired service
	ptr := pointer.GetPointer(spec, serviceName)
	if ptr == nil {
		slog.Error("Unable to add hello param to " + serviceName + " as it is not a pointer. Did you forget to define " + serviceName + "? You can define serviceName using `workflow.Service`")
		return
	}

	// Attach the Hello wrapper node to the server-side node chain of the desired service
	serverNext := ptr.AddDstModifier(spec, wrapper_name)

	// Define the IRNode for the wrapper node and add it to the wiring specification
	spec.Define(wrapper_name, &HelloParamServerWrapper{}, func(ns wiring.Namespace) (ir.IRNode, error) {
		// Get the IRNode that will be wrapped by HelloWrapper
		var server golang.Service
		if err := ns.Get(serverNext, &server); err != nil {
			return nil, blueprint.Errorf("Tutorial Plugin %s expected %s to be a golang.Service, but encountered %s", wrapper_name, serverNext, err)
		}

		// Instantiate the IRNode
		return newHelloParamServerWrapper(wrapper_name, server)
	})

	// Attach the Hello wrapper node to the client-side node chain of the desired service
	clientNext := ptr.AddSrcModifier(spec, client_wrapper_name)

	// Define the IRNode for the wrapper node and add it to the wiring specification
	spec.Define(client_wrapper_name, &HelloParamClientWrapper{}, func(ns wiring.Namespace) (ir.IRNode, error) {
		// Get the IRNode that will be wrapped by HelloWrapper
		var client golang.Service
		if err := ns.Get(clientNext, &client); err != nil {
			return nil, blueprint.Errorf("Tutorial Plugin %s expected %s to be a golang.Service, but encountered %s", wrapper_name, serverNext, err)
		}

		return newHelloParamClientWrapper(client_wrapper_name, client)
	})
}
```

### IR Nodes

We need to define two different IR Nodes: (i) an IR Node that wraps and extends the server-side of the desired service; and (ii) an IR Node that wraps and extends the client-side of the desired service.

The IR Nodes must now also generate a new interface as they have modified the service's existing interface by extending the function signatures with new parameters.

Following is the implementation of the server-side IR Node.

```go
// Blueprint IRNode for representing the wrapper node that adds an additional call parameter and an additional return parameter to every server-side method in the node that gets wrapped
type HelloParamServerWrapper struct {
	golang.Service
	golang.GeneratesFuncs
	golang.Instantiable

	InstanceName string
	Wrapped      golang.Service

	outputPackage string
}

// Implements ir.IRNode
func (node *HelloParamServerWrapper) ImplementsGolangNode() {}

// Implements ir.IRNode
func (node *HelloParamServerWrapper) Name() string {
	return node.InstanceName
}

// Implements ir.IRNode
func (node *HelloParamServerWrapper) String() string {
	return node.Name() + " = HelloParamServerWrapper(" + node.Wrapped.Name() + ")"
}

// Implements golang.ProvidesInterface
func (node *HelloParamServerWrapper) AddInterfaces(builder golang.ModuleBuilder) error {
	iface, err := node.genInterface(builder)
	if err != nil {
		return err
	}
	err = generateClientSideParamInterfaces(builder, iface, node.outputPackage)
	if err != nil {
		return err
	}
	return node.Wrapped.AddInterfaces(builder)
}

func newHelloParamServerWrapper(name string, server ir.IRNode) (*HelloParamServerWrapper, error) {
	serverNode, ok := server.(golang.Service)
	if !ok {
		return nil, blueprint.Errorf("tutorial server wrapper requires %s to be a golang service but got %s", server.Name(), reflect.TypeOf(server).String())
	}

	node := &HelloParamServerWrapper{}
	node.InstanceName = name
	node.Wrapped = serverNode
	node.outputPackage = "tutorial_param"

	return node, nil
}

func (node *HelloParamServerWrapper) genInterface(ctx ir.BuildContext) (*gocode.ServiceInterface, error) {
	iface, err := golang.GetGoInterface(ctx, node.Wrapped)
	if err != nil {
		return nil, err
	}
	module_ctx, valid := ctx.(golang.ModuleBuilder)
	if !valid {
		return nil, blueprint.Errorf("Tutorial expected build context to be a ModuleBuilder, got %v", ctx)
	}
	i := gocode.CopyServiceInterface(fmt.Sprintf("%v_TutorialParam", iface.BaseName), module_ctx.Info().Name+"/"+node.outputPackage, iface)
	for name, method := range i.Methods {
		method.AddArgument(gocode.Variable{Name: "extraparam", Type: &gocode.BasicType{Name: "string"}})
		method.AddRetVar(gocode.Variable{Name: "", Type: &gocode.BasicType{Name: "string"}})
		i.Methods[name] = method
	}
	return i, nil
}

// Implements service.ServiceNode
func (node *HelloParamServerWrapper) GetInterface(ctx ir.BuildContext) (service.ServiceInterface, error) {
	return node.genInterface(ctx)
}

// Implements golang.GeneratesFuncs
func (node *HelloParamServerWrapper) GenerateFuncs(builder golang.ModuleBuilder) error {
	service, err := golang.GetGoInterface(builder, node.Wrapped)
	if err != nil {
		return err
	}
	iface, err := node.genInterface(builder)
	if err != nil {
		return err
	}
	err = generateServerParamHandler(builder, iface, service, node.outputPackage)
	if err != nil {
		return err
	}
	return nil
}

// Implements golang.Instantiable
func (node *HelloParamServerWrapper) AddInstantiation(builder golang.NamespaceBuilder) error {
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
			Name: fmt.Sprintf("New_%v_TutorialParamServerWrapper", iface.BaseName),
			Arguments: []gocode.Variable{
				{Name: "ctx", Type: &gocode.UserType{Package: "context", Name: "context"}},
				{Name: "service", Type: iface},
			},
		},
	}

	return builder.DeclareConstructor(node.InstanceName, constructor, []ir.IRNode{node.Wrapped})
}
```

Following is the implementation of the client-side IR Node.

```go
// Blueprint IRNode for representing the wrapper node that adds an additional call parameter and an additional return parameter to every client-side method in the node that gets wrapped
type HelloParamClientWrapper struct {
	golang.Service
	golang.GeneratesFuncs
	golang.Instantiable

	InstanceName string
	Wrapped      golang.Service

	outputPackage string
}

// Implements ir.IRNode
func (node *HelloParamClientWrapper) ImplementsGolangNode() {}

// Implements ir.IRNode
func (node *HelloParamClientWrapper) Name() string {
	return node.InstanceName
}

// Implements ir.IRNode
func (node *HelloParamClientWrapper) String() string {
	return node.Name() + " = HelloParamClientWrapper(" + node.Wrapped.Name() + ")"
}

// Implements golang.ProvidesInterface
func (node *HelloParamClientWrapper) AddInterfaces(builder golang.ModuleBuilder) error {
	return node.Wrapped.AddInterfaces(builder)
}

func (node *HelloParamClientWrapper) genInterface(ctx ir.BuildContext) (*gocode.ServiceInterface, error) {
	iface, err := golang.GetGoInterface(ctx, node.Wrapped)
	if err != nil {
		return nil, err
	}
	module_ctx, valid := ctx.(golang.ModuleBuilder)
	if !valid {
		return nil, blueprint.Errorf("TutorialParamClientWrapper expected build context to be a ModuleBuilder, got %v", ctx)
	}
	i := gocode.CopyServiceInterface(fmt.Sprintf("%v_TutorialParamClientWrapperInterface", iface.BaseName), module_ctx.Info().Name+"/"+node.outputPackage, iface)
	for name, method := range i.Methods {
		method.Arguments = method.Arguments[:len(method.Arguments)-1]
		method.Returns = method.Returns[:len(method.Returns)-1]
		i.Methods[name] = method
	}
	return i, nil
}

func newHelloParamClientWrapper(name string, server ir.IRNode) (*HelloParamClientWrapper, error) {
	serverNode, ok := server.(golang.Service)
	if !ok {
		return nil, blueprint.Errorf("tutorial server wrapper requires %s to be a golang service but got %s", server.Name(), reflect.TypeOf(server).String())
	}

	node := &HelloParamClientWrapper{}
	node.InstanceName = name
	node.Wrapped = serverNode
	node.outputPackage = "tutorial_param"

	return node, nil
}

// Implements service.ServiceNode
func (node *HelloParamClientWrapper) GetInterface(ctx ir.BuildContext) (service.ServiceInterface, error) {
	return node.genInterface(ctx)
}

// Implements golang.GeneratesFuncs
func (node *HelloParamClientWrapper) GenerateFuncs(builder golang.ModuleBuilder) error {
	service, err := golang.GetGoInterface(builder, node.Wrapped)
	if err != nil {
		return err
	}
	iface, err := golang.GetGoInterface(builder, node)
	if err != nil {
		return err
	}
	err = generateClientParamHandler(builder, iface, service, node.outputPackage)
	if err != nil {
		return err
	}
	return nil
}

// Implements golang.Instantiable
func (node *HelloParamClientWrapper) AddInstantiation(builder golang.NamespaceBuilder) error {
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
			Name: fmt.Sprintf("New_%v_TutorialParamClientWrapper", iface.BaseName),
			Arguments: []gocode.Variable{
				{Name: "ctx", Type: &gocode.UserType{Package: "context", Name: "context"}},
				{Name: "service", Type: iface},
			},
		},
	}

	return builder.DeclareConstructor(node.InstanceName, constructor, []ir.IRNode{node.Wrapped})
}
```

### Code Generation

The following generates the server-side wrapper.

```go

func generateServerParamHandler(builder golang.ModuleBuilder, iface *gocode.ServiceInterface, wrapped_service *gocode.ServiceInterface, outputPackage string) error {
	pkg, err := builder.CreatePackage(outputPackage)
	if err != nil {
		return err
	}

	server := &serverArgs{
		Package:   pkg,
		Service:   wrapped_service,
		Iface:     iface,
		Name:      wrapped_service.BaseName + "_TutorialParamServerWrapper",
		IfaceName: iface.Name,
		Imports:   gogen.NewImports(pkg.Name),
	}

	server.Imports.AddPackages("context")

	slog.Info(fmt.Sprintf("Generating %v/%v", server.Package.PackageName, wrapped_service.BaseName+"_TutorialParamServerWrapper"))
	outputFile := filepath.Join(server.Package.Path, wrapped_service.BaseName+"_TutorialParamServerWrapper.go")
	return gogen.ExecuteTemplateToFile("Tutorial", serverParamTemplate, server, outputFile)
}

var serverParamTemplate = `// Blueprint: Auto-generated by Tutorial Plugin
package {{.Package.ShortName}}

{{.Imports}}

type {{.Name}} struct {
	Service {{.Imports.NameOf .Service.UserType}}
}

func New_{{.Name}}(ctx context.Context, service {{.Imports.NameOf .Service.UserType}}) (*{{.Name}}, error) {
	handler := &{{.Name}}{}
	handler.Service = service
	return handler, nil
}

{{$service := .Service.Name -}}
{{$receiver := .Name -}}
{{ range $_, $f := .Service.Methods}}
func (handler *{{$receiver}}) {{$f.Name -}} ({{ArgVarsAndTypes $f "ctx context.Context"}}, extraparam string) ({{RetVarsAndTypes $f "retparam string" "err error"}}) {
	{{RetVars $f "err"}} = handler.Service.{{$f.Name}}({{ArgVars $f "ctx"}})
	retparam = extraparam
	return
}
{{end}}
`
```

On the client-side, an interface is generated that can be used by the clients to correctly connect to the service.

```go
func generateClientSideParamInterfaces(builder golang.ModuleBuilder, iface *gocode.ServiceInterface, outputPackage string) error {
	pkg, err := builder.CreatePackage(outputPackage)
	if err != nil {
		return err
	}

	server := &serverArgs{
		Package:   pkg,
		Iface:     iface,
		IfaceName: iface.Name,
		Imports:   gogen.NewImports(pkg.Name),
	}

	server.Imports.AddPackages("context")
	slog.Info(fmt.Sprintf("Generating %v/%v", server.Package.PackageName, iface.Name))
	outputFile := filepath.Join(server.Package.Path, iface.Name+".go")
	// Re-use the template from ir_method
	return gogen.ExecuteTemplateToFile("Tutorial", clientTemplate, server, outputFile)
}
```

The following generates the client-side wrapper.

```go
func generateClientParamHandler(builder golang.ModuleBuilder, iface *gocode.ServiceInterface, wrapped_service *gocode.ServiceInterface, outputPackage string) error {
	pkg, err := builder.CreatePackage(outputPackage)
	if err != nil {
		return err
	}

	server := &serverArgs{
		Package:         pkg,
		Service:         wrapped_service,
		Iface:           iface,
		Name:            wrapped_service.BaseName + "_TutorialParamClientWrapper",
		IfaceName:       iface.Name,
		ServerIfaceName: wrapped_service.Name,
		Imports:         gogen.NewImports(pkg.Name),
	}

	server.Imports.AddPackages("context", "log")

	slog.Info(fmt.Sprintf("Generating %v/%v", server.Package.PackageName, wrapped_service.BaseName+"_TutorialParamClientWrapper"))
	outputFile := filepath.Join(server.Package.Path, wrapped_service.BaseName+"_TutorialParamClientWrapper.go")
	return gogen.ExecuteTemplateToFile("Tutorial", clientParamTemplate, server, outputFile)
}

var clientParamTemplate = `// Blueprint: Auto-generated by Tutorial plugin
package {{.Package.ShortName}}

{{.Imports}}

type {{.IfaceName}} interface {
	{{range $_, $f := .Iface.Methods -}}
	{{Signature $f}}
	{{end}}
}

type {{.Name}} struct {
	Client {{.ServerIfaceName}}
}

func New_{{.Name}}(ctx context.Context, client {{.ServerIfaceName}}) (*{{.Name}}, error) {
	handler := &{{.Name}}{}
	handler.Client = client
	return handler, nil
}

{{$service := .Service.Name -}}
{{$receiver := .Name -}}
{{ range $_, $f := .Iface.Methods }}
func (handler *{{$receiver}}) {{$f.Name -}} ({{ArgVarsAndTypes $f "ctx context.Context"}}) ({{RetVarsAndTypes $f "err error"}}) {
	var retparam string
	{{RetVars $f "retparam" "err"}} = handler.Client.{{$f.Name}}({{ArgVars $f "ctx"}}, "Extra!")
	log.Println("Ret param was ", retparam)
	return
}
{{end}}
`

```

## Tutorial: Advanced Concepts

### Addresses

A plugin might introduce new components that require addresses to bind to. This is fairly common for plugins that implement backends that run as standalone servers.

To define a new address IR Node in the wiring spec, use the [address.Define](https://github.com/Blueprint-uServices/blueprint/blob/main/blueprint/pkg/coreplugins/address/wiring.go#L30) method. You can then bind a server's IR Node to this address Node by using [address.Bind](https://github.com/Blueprint-uServices/blueprint/blob/main/blueprint/pkg/coreplugins/address/wiring.go#L93) and have the clients to this server correctly establish connections to this standalone server using [address.Dial](https://github.com/Blueprint-uServices/blueprint/blob/main/blueprint/pkg/coreplugins/address/wiring.go#L69).

### Namespaces

A plugin might introduce a new [Namespace](https://github.com/Blueprint-uServices/blueprint/blob/main/blueprint/pkg/wiring/namespace.go#L26) that groups various IR Nodes together. To introduce a new namespace, the plugin must implement the [NamespaceHandler](https://github.com/Blueprint-uServices/blueprint/blob/main/blueprint/pkg/wiring/namespace.go#L114) interface.

Example plugins that introduce a new namespace: [clientpool](https://github.com/Blueprint-uServices/blueprint/tree/main/plugins/clientpool), [dockercompose](https://github.com/Blueprint-uServices/blueprint/tree/main/plugins/dockercompose).