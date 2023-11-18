---
title: plugins/golang/goparser
---
# plugins/golang/goparser
```go
package goparser // import "gitlab.mpi-sws.org/cld/blueprint/plugins/golang/goparser"
```

## TYPES

A set of modules on the local filesystem that contain workflow spec
interfaces and implementations. It is allowed for a workflow spec
implementation in one package to use the interface defined in another
package. However, currently, it is not possible to use workflow spec nodes
whose interface or implementation comes entirely from an external module
(ie. a module that exists only as a 'require' directive of a go.mod)
```go
type ParsedField struct {
	gocode.Variable
	Struct   *ParsedStruct
	Position int
	Ast      *ast.Field
}
```
## func 
```go
func (f *ParsedField) Parse() error
```

## func 
```go
func (f *ParsedField) String() string
```

A set of modules on the local filesystem that contain workflow spec
interfaces and implementations. It is allowed for a workflow spec
implementation in one package to use the interface defined in another
package. However, currently, it is not possible to use workflow spec nodes
whose interface or implementation comes entirely from an external module
(ie. a module that exists only as a 'require' directive of a go.mod)
```go
type ParsedFile struct {
	Package          *ParsedPackage
	Name             string                   // Filename
	Path             string                   // Fully qualified path to the file
	AnonymousImports []*ParsedImport          // Import declarations that were imported with .
	NamedImports     map[string]*ParsedImport // Import declarations - map from shortname to fully qualified package import name
	Ast              *ast.File                // The AST of the file
}
```
## func 
```go
func (f *ParsedFile) LoadFuncs() error
```
Assumes that all structs and interfaces have been loaded for the package
containing the file.

Loads the names of all funcs. If the func has a receiver type, then it is
saved as a method on the appropriate struct; if it does not have a receiver
type, then it is saved as a package func.

This does not parse the arguments or returns of the func

## func 
```go
func (f *ParsedFile) LoadImports() error
```

## func 
```go
func (f *ParsedFile) LoadStructsAndInterfaces() error
```
Looks for:
  - structs defined in the file
  - interfaces defined in the file
  - other user types defined in the file

Does not:
  - look for function declarations

## func 
```go
func (f *ParsedFile) ResolveIdent(name string) gocode.TypeName
```
An ident can be:
  - a basic type, like int64, float32 etc.
  - any
  - a type declared locally within the file or package
  - a type imported with an `import . "package"` decl

## func 
```go
func (f *ParsedFile) ResolveSelector(packageShortName string, name string) gocode.TypeName
```

## func 
```go
func (f *ParsedFile) ResolveType(expr ast.Expr) gocode.TypeName
```

## func 
```go
func (f *ParsedFile) String() string
```

A set of modules on the local filesystem that contain workflow spec
interfaces and implementations. It is allowed for a workflow spec
implementation in one package to use the interface defined in another
package. However, currently, it is not possible to use workflow spec nodes
whose interface or implementation comes entirely from an external module
(ie. a module that exists only as a 'require' directive of a go.mod)
```go
type ParsedFunc struct {
	gocode.Func
	File *ParsedFile
	Ast  *ast.FuncType
}
```
## func 
```go
func (f *ParsedFunc) AsConstructor() *gocode.Constructor
```

## func 
```go
func (f *ParsedFunc) Parse() error
```

## func 
```go
func (f *ParsedFunc) String() string
```

A set of modules on the local filesystem that contain workflow spec
interfaces and implementations. It is allowed for a workflow spec
implementation in one package to use the interface defined in another
package. However, currently, it is not possible to use workflow spec nodes
whose interface or implementation comes entirely from an external module
(ie. a module that exists only as a 'require' directive of a go.mod)
```go
type ParsedImport struct {
	File    *ParsedFile
	Package string
}
```
A set of modules on the local filesystem that contain workflow spec
interfaces and implementations. It is allowed for a workflow spec
implementation in one package to use the interface defined in another
package. However, currently, it is not possible to use workflow spec nodes
whose interface or implementation comes entirely from an external module
(ie. a module that exists only as a 'require' directive of a go.mod)
```go
type ParsedInterface struct {
	File    *ParsedFile
	Ast     *ast.InterfaceType
	Name    string
	Methods map[string]*ParsedFunc
}
```
## func 
```go
func (iface *ParsedInterface) ServiceInterface(ctx ir.BuildContext) *gocode.ServiceInterface
```

## func 
```go
func (iface *ParsedInterface) Type() *gocode.UserType
```

A set of modules on the local filesystem that contain workflow spec
interfaces and implementations. It is allowed for a workflow spec
implementation in one package to use the interface defined in another
package. However, currently, it is not possible to use workflow spec nodes
whose interface or implementation comes entirely from an external module
(ie. a module that exists only as a 'require' directive of a go.mod)
```go
type ParsedModule struct {
	ModuleSet *ParsedModuleSet
	Name      string                    // Fully qualified name of the module
	Version   string                    // Version of the module
	SrcDir    string                    // Fully qualified location of the module on the filesystem
	Modfile   *modfile.File             // The modfile File struct is sufficiently simple that we just use it directly
	Packages  map[string]*ParsedPackage // Map from fully qualified package name to ParsedPackage
}
```
## func 
```go
func (mod *ParsedModule) Load() error
```

## func 
```go
func (mod *ParsedModule) String() string
```

A set of modules on the local filesystem that contain workflow spec
interfaces and implementations. It is allowed for a workflow spec
implementation in one package to use the interface defined in another
package. However, currently, it is not possible to use workflow spec nodes
whose interface or implementation comes entirely from an external module
(ie. a module that exists only as a 'require' directive of a go.mod)
```go
type ParsedModuleSet struct {
	Modules map[string]*ParsedModule // Map from FQ module name to module object
}
```
## func ParseModules
```go
func ParseModules(srcDirs ...string) (*ParsedModuleSet, error)
```
Parse the specified module directories

## func ParseWorkspace
```go
func ParseWorkspace(workspaceDir string) (*ParsedModuleSet, error)
```
Parse all modules in the specified directory

## func 
```go
func (set *ParsedModuleSet) AddModule(srcDir string) (*ParsedModule, error)
```

## func 
```go
func (set *ParsedModuleSet) GetPackage(name string) *ParsedPackage
```

## func 
```go
func (set *ParsedModuleSet) String() string
```

A set of modules on the local filesystem that contain workflow spec
interfaces and implementations. It is allowed for a workflow spec
implementation in one package to use the interface defined in another
package. However, currently, it is not possible to use workflow spec nodes
whose interface or implementation comes entirely from an external module
(ie. a module that exists only as a 'require' directive of a go.mod)
```go
type ParsedPackage struct {
	Module        *ParsedModule
	Name          string                      // Fully qualified name of the package including module name
	ShortName     string                      // Shortname of the package (ie, the name used in an import statement)
	PackageDir    string                      // Subdirectory within the module containing the package
	SrcDir        string                      // Fully qualified location of the package on the filesystem
	Files         map[string]*ParsedFile      // Map from filename to ParsedFile
	Ast           *ast.Package                // The AST of the package
	DeclaredTypes map[string]gocode.UserType  // Types declared within this package
	Structs       map[string]*ParsedStruct    // Structs parsed from this package
	Interfaces    map[string]*ParsedInterface // Interfaces parsed from this package
	Funcs         map[string]*ParsedFunc      // Functions parsed from this package (does not include funcs with receiver types)
}
```
## func 
```go
func (pkg *ParsedPackage) Load() error
```

## func 
```go
func (pkg *ParsedPackage) Parse() error
```

## func 
```go
func (pkg *ParsedPackage) String() string
```

A set of modules on the local filesystem that contain workflow spec
interfaces and implementations. It is allowed for a workflow spec
implementation in one package to use the interface defined in another
package. However, currently, it is not possible to use workflow spec nodes
whose interface or implementation comes entirely from an external module
(ie. a module that exists only as a 'require' directive of a go.mod)
```go
type ParsedStruct struct {
	File            *ParsedFile
	Ast             *ast.StructType
	Name            string
	Methods         map[string]*ParsedFunc  // Methods declared directly on this struct, does not include promoted methods (not implemented yet)
	FieldsList      []*ParsedField          // All fields in the order that they are declared
	Fields          map[string]*ParsedField // Named fields declared in this struct only, does not include promoted fields (not implemented yet)
	PromotedField   *ParsedField            // If there is a promoted field, stored here
	AnonymousFields []*ParsedField          // Subsequent anonymous fields
}
```
## func 
```go
func (f *ParsedStruct) String() string
```

## func 
```go
func (struc *ParsedStruct) Type() *gocode.UserType
```


