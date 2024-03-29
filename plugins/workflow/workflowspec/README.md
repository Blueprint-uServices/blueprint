<!-- Code generated by gomarkdoc. DO NOT EDIT -->

# workflowspec

```go
import "github.com/blueprint-uservices/blueprint/plugins/workflow/workflowspec"
```

## Index

- [func Add\[T any\]\(\) error](<#Add>)
- [func AddModule\(moduleName string\) error](<#AddModule>)
- [type Service](<#Service>)
  - [func GetService\[T any\]\(\) \(\*Service, error\)](<#GetService>)
  - [func GetServiceByName\(pkg, name string\) \(\*Service, error\)](<#GetServiceByName>)
  - [func \(s \*Service\) AddToModule\(builder golang.ModuleBuilder\) error](<#Service.AddToModule>)
  - [func \(s \*Service\) AddToWorkspace\(builder golang.WorkspaceBuilder\) error](<#Service.AddToWorkspace>)
  - [func \(s \*Service\) Modules\(\) \[\]\*goparser.ParsedModule](<#Service.Modules>)
- [type WorkflowSpec](<#WorkflowSpec>)
  - [func Get\(\) \*WorkflowSpec](<#Get>)
  - [func New\(\) \*WorkflowSpec](<#New>)
  - [func \(spec \*WorkflowSpec\) Derive\(\) \*WorkflowSpec](<#WorkflowSpec.Derive>)
  - [func \(spec \*WorkflowSpec\) Parse\(modInfo \*goparser.ModuleInfo\) error](<#WorkflowSpec.Parse>)
  - [func \(spec \*WorkflowSpec\) ParseModule\(moduleName string\) error](<#WorkflowSpec.ParseModule>)


<a name="Add"></a>
## func [Add](<https://github.com/blueprint-uservices/blueprint/blob/main/plugins/workflow/workflowspec/cache.go#L19>)

```go
func Add[T any]() error
```

Parses & adds the module containing T to the cached workflow spec.

<a name="AddModule"></a>
## func [AddModule](<https://github.com/blueprint-uservices/blueprint/blob/main/plugins/workflow/workflowspec/cache.go#L14>)

```go
func AddModule(moduleName string) error
```

Parses & adds a module to the workflow spec search path

<a name="Service"></a>
## type [Service](<https://github.com/blueprint-uservices/blueprint/blob/main/plugins/workflow/workflowspec/service.go#L15-L21>)

A service in the workflow spec

```go
type Service struct {
    // The interface that the service implements
    Iface *goparser.ParsedInterface

    // The constructor func of the service
    Constructor *goparser.ParsedFunc
}
```

<a name="GetService"></a>
### func [GetService](<https://github.com/blueprint-uservices/blueprint/blob/main/plugins/workflow/workflowspec/cache.go#L46>)

```go
func GetService[T any]() (*Service, error)
```

Gets a \[WorkflowSpecService\] for the specified type. Type parameter T should be a service defined in an application's workflow spec or a plugin's runtime directory.

The definition of the service T will be acquired by parsing the module where T is defined. Thus to utilize a specific version of T, set that version in the go.mod file of the wiring spec when requiring T's module.

#### Example Usage

```
leaf := workflowspec.GetService[leaf.LeafService]()
```

#### Internals

By using type parameter T, it ensures that wherever T is defined, its module and version will be on the go path / within the go.mod. By contrast, using [GetServiceByName](<#GetServiceByName>) might fail if it names a package that doesn't exist in the local go cache / on the go path.

<a name="GetServiceByName"></a>
### func [GetServiceByName](<https://github.com/blueprint-uservices/blueprint/blob/main/plugins/workflow/workflowspec/cache.go#L70>)

```go
func GetServiceByName(pkg, name string) (*Service, error)
```

Gets a \[WorkflowSpecService\] for the specified type. pkg and name should be the package and name of a service defined in an application's workflow spec or a plugin's runtime directory.

#### Example Usage

```
leaf := workflowspec.GetServiceByName("github.com/blueprint-uservices/blueprint/examples/leaf", "LeafService")
```

#### Internals

This method is not as robust as [GetService](<#GetService>) and it might fail if pkg isn't a local package or isn't a go.mod dependency. Ensure the named package is in the go.mod file of the application. Anonymously importing a package can help ensure it is not erased from your go.mod file, e.g.

```
import _ "github.com/blueprint-uservices/blueprint/examples/sockshop/tests"
```

<a name="Service.AddToModule"></a>
### func \(\*Service\) [AddToModule](<https://github.com/blueprint-uservices/blueprint/blob/main/plugins/workflow/workflowspec/service.go#L29>)

```go
func (s *Service) AddToModule(builder golang.ModuleBuilder) error
```



<a name="Service.AddToWorkspace"></a>
### func \(\*Service\) [AddToWorkspace](<https://github.com/blueprint-uservices/blueprint/blob/main/plugins/workflow/workflowspec/service.go#L42>)

```go
func (s *Service) AddToWorkspace(builder golang.WorkspaceBuilder) error
```



<a name="Service.Modules"></a>
### func \(\*Service\) [Modules](<https://github.com/blueprint-uservices/blueprint/blob/main/plugins/workflow/workflowspec/service.go#L25>)

```go
func (s *Service) Modules() []*goparser.ParsedModule
```

Get all modules containing definitions for this service. Could be more than one if the interface and implementation are defined in separate modules.

<a name="WorkflowSpec"></a>
## type [WorkflowSpec](<https://github.com/blueprint-uservices/blueprint/blob/main/plugins/workflow/workflowspec/spec.go#L18-L20>)

Representation of a parsed workflow spec.

This code makes heavy use of the Golang code parser defined in the Golang plugin. That code parser extracts structs, interfaces, and function definitions from a set of golang modules.

This code adds functionality that:

- Identifies valid service interfaces
- Matches structs to interfaces that they implement
- Finds constructors of structs

```go
type WorkflowSpec struct {
    Modules *goparser.ParsedModuleSet
}
```

<a name="Get"></a>
### func [Get](<https://github.com/blueprint-uservices/blueprint/blob/main/plugins/workflow/workflowspec/cache.go#L9>)

```go
func Get() *WorkflowSpec
```

Returns a shared / cached WorkflowSpec

<a name="New"></a>
### func [New](<https://github.com/blueprint-uservices/blueprint/blob/main/plugins/workflow/workflowspec/spec.go#L30>)

```go
func New() *WorkflowSpec
```

Returns a new [WorkflowSpec](<#WorkflowSpec>) for parsing workflow modules from scratch.

Most plugins typically don't need to reference the workflow spec directly and can just call [GetService](<#GetService>) or [GetServiceByName](<#GetServiceByName>).

Similarly most plugins shouldn't need to construct their own workflow spec from scratch, and instead should be able to make use of the existing cached one \(through calling Get\(\)\)

<a name="WorkflowSpec.Derive"></a>
### func \(\*WorkflowSpec\) [Derive](<https://github.com/blueprint-uservices/blueprint/blob/main/plugins/workflow/workflowspec/spec.go#L36>)

```go
func (spec *WorkflowSpec) Derive() *WorkflowSpec
```

Derives a [WorkflowSpec](<#WorkflowSpec>) from an existing one, making use of all of the modules already loaded in this workflow spec.

<a name="WorkflowSpec.Parse"></a>
### func \(\*WorkflowSpec\) [Parse](<https://github.com/blueprint-uservices/blueprint/blob/main/plugins/workflow/workflowspec/spec.go#L58>)

```go
func (spec *WorkflowSpec) Parse(modInfo *goparser.ModuleInfo) error
```

Parses the specified module info and adds it to the workflow spec.

<a name="WorkflowSpec.ParseModule"></a>
### func \(\*WorkflowSpec\) [ParseModule](<https://github.com/blueprint-uservices/blueprint/blob/main/plugins/workflow/workflowspec/spec.go#L48>)

```go
func (spec *WorkflowSpec) ParseModule(moduleName string) error
```

Looks up the specified module, parses it, and adds it to the workflow spec.

Generated by [gomarkdoc](<https://github.com/princjef/gomarkdoc>)
