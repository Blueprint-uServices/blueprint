---
title: plugins/golang/gogen
---
# plugins/golang/gogen
```go
package gogen // import "gitlab.mpi-sws.org/cld/blueprint/plugins/golang/gogen"
```

## FUNCTIONS

## func ExecuteTemplate
```go
func ExecuteTemplate(name string, body string, args any) (string, error)
```
## func ExecuteTemplateToFile
```go
func ExecuteTemplateToFile(name string, body string, args any, filename string) error
```

## TYPES

```go
type GraphBuilderImpl struct {
	ir.VisitTrackerImpl
```
Implements the GraphBuilder interface defined in gogen/ir.go
```go
	Package      golang.PackageInfo
	FileName     string            // The short name of the file
	FilePath     string            // The fully qualified path to the file
	FuncName     string            // The name of the function to generate
	Imports      *Imports          // Import declarations in the file; map of shortname to full package import name
	Declarations map[string]string // The DI declarations
	// Has unexported fields.
}
```
GraphBuilder is useful for any golang namespace node, such as the goproc
plugin, that wants to be able to instantiate Golang IR nodes.

The GraphBuilder accumulates code from child nodes and stubs for
instantiating the code.

The GraphBuilder can then be used to generate a source file that constructs
the graph.

The typical usage of a `GraphBuilder` is to:

 1. Create a new `GraphBuilder` with the `NewGraphBuilder` method

 2. Collect code definitions from child nodes by calling
    `Instantiable.AddInstantiation` on those nodes

 3. Generate the final output by calling `GraphBuilder.Finish`

## func NewGraphBuilder
```go
func NewGraphBuilder(module golang.ModuleBuilder, fileName, packagePath, funcName string) (*GraphBuilderImpl, error)
```
Create a new GraphBuilder

## func 
```go
func (code *GraphBuilderImpl) Build() error
```
Generates the file within its module

## func 
```go
func (graph *GraphBuilderImpl) Declare(name, buildFuncSrc string) error
```

## func 
```go
func (graph *GraphBuilderImpl) DeclareConstructor(name string, constructor *gocode.Constructor, args []ir.IRNode) error
```

## func 
```go
func (code *GraphBuilderImpl) ImplementsBuildContext()
```

## func 
```go
func (graph *GraphBuilderImpl) Import(packageName string) string
```

## func 
```go
func (graph *GraphBuilderImpl) ImportType(typeName gocode.TypeName) string
```

## func 
```go
func (graph *GraphBuilderImpl) Info() golang.GraphInfo
```

## func 
```go
func (code *GraphBuilderImpl) Module() golang.ModuleBuilder
```

A helper struct for managing imports in generated golang files.
```go
type Imports struct {
	// Has unexported fields.
}
```
Used by plugins like the GRPC plugin.

The string representation of the Imports struct is the import declaration.

The NameOf method provides the correctly qualified name for the specified
userType

## func NewImports
```go
func NewImports(packageName string) *Imports
```
Creates a new ImportedPackages struct, treating the provided fully-qualified
packageName as the "current" package

## func 
```go
func (imports *Imports) AddPackage(pkg string) string
```

## func 
```go
func (imports *Imports) AddPackages(pkgs ...string)
```

## func 
```go
func (imports *Imports) AddType(typeName gocode.TypeName)
```

## func 
```go
func (imports *Imports) NameOf(typeName gocode.TypeName) string
```

## func 
```go
func (imports *Imports) Qualify(pkg string, name string) string
```

## func 
```go
func (imports *Imports) String() string
```

```go
type ModuleBuilderImpl struct {
	ir.VisitTrackerImpl
	Name string // The FQ name of this module
```
Implements the ModuleBuilder interface defined in golang/ir.go
```go
	ModuleDir string // The directory containing this module
	// Has unexported fields.
}
```
The ModuleBuilder is used by plugins that generate golang source files,
and need a module to put that code into.

## func NewModuleBuilder
```go
func NewModuleBuilder(workspace *WorkspaceBuilderImpl, moduleName string) (*ModuleBuilderImpl, error)
```
This method is used by plugins if they want to generate their own Go module
from scratch.

After dependencies and code have been added to the module, plugins must call
Generate on the Generated Module to finish building it.

## func 
```go
func (module *ModuleBuilderImpl) Build(nodes []ir.IRNode) error
```

## func 
```go
func (module *ModuleBuilderImpl) CreatePackage(packageName string) (golang.PackageInfo, error)
```

## func 
```go
func (module *ModuleBuilderImpl) ImplementsBuildContext()
```

## func 
```go
func (module *ModuleBuilderImpl) Info() golang.ModuleInfo
```

## func 
```go
func (module *ModuleBuilderImpl) Workspace() golang.WorkspaceBuilder
```

Implements the WorkspaceBuilder interface defined in golang/ir.go
```go
type WorkspaceBuilderImpl struct {
	ir.VisitTrackerImpl
	WorkspaceDir     string            // The directory containing this workspace
	ModuleDirs       map[string]string // map from FQ module name to directory name within WorkspaceDir
	Modules          map[string]string // map from directory name to FQ module name within WorkspaceDir
	GeneratedModules map[string]string // map from directory name to FQ module name within WorkspaceDir
}
```
The WorkspaceBuilder is used for accumulating local module directories into
a golang workspace.

## func NewWorkspaceBuilder
```go
func NewWorkspaceBuilder(workspaceDir string) (*WorkspaceBuilderImpl, error)
```
Creates a new WorkspaceBuilder at the specified output dir.

Will return an error if the workspacedir already exists

## func 
```go
func (workspace *WorkspaceBuilderImpl) AddLocalModule(shortName string, moduleSrcPath string) error
```

## func 
```go
func (workspace *WorkspaceBuilderImpl) AddLocalModuleRelative(shortName string, relativeModuleSrcPath string) error
```
This method is used by plugins if they want to copy a locally-defined module
into the generated workspace.

The specified relativeModuleSrcPath must point to a valid Go module with a
go.mod file, relative to the calling file's location.

## func 
```go
func (workspace *WorkspaceBuilderImpl) CreateModule(moduleName string, moduleVersion string) (string, error)
```

## func 
```go
func (workspace *WorkspaceBuilderImpl) Finish() error
```
This method should be used by plugins after all modules in a workspace have
been combined.

The method will do the following:
  - creates a go.work file in the root of the workspace that points to all
    of the modules contained therein
  - updates the go.mod files of all contained modules with 'replace'
    directives for any required modules that exist in the workspace

## func 
```go
func (workspace *WorkspaceBuilderImpl) GetLocalModule(modulePath string) (string, bool)
```

## func 
```go
func (workspace *WorkspaceBuilderImpl) ImplementsBuildContext()
```

## func 
```go
func (workspace *WorkspaceBuilderImpl) Info() golang.WorkspaceInfo
```


