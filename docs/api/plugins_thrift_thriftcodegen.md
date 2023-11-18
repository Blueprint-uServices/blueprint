---
title: plugins/thrift/thriftcodegen
---
# plugins/thrift/thriftcodegen
```go
package thriftcodegen // import "gitlab.mpi-sws.org/cld/blueprint/plugins/thrift/thriftcodegen"
```

## FUNCTIONS

## func CompileThriftFile
```go
func CompileThriftFile(thriftFileName string) error
```
## func GenerateClient
```go
func GenerateClient(builder golang.ModuleBuilder, service *gocode.ServiceInterface, outputPackage string) error
```
## func GenerateServerHandler
```go
func GenerateServerHandler(builder golang.ModuleBuilder, service *gocode.ServiceInterface, outputPackage string) error
```
## func GenerateThrift
```go
func GenerateThrift(builder golang.ModuleBuilder, service *gocode.ServiceInterface, outputPackage string) error
```
Generates the .thrift file for the provided service interface, then compiles
it using `thrift`. See the plugin README for the required thrift package
dependencies.


## TYPES

```go
type ThriftBuilder struct {
	Code        *goparser.ParsedModuleSet
	Package     string
	Module      golang.ModuleInfo
	PackageName string
	ImportName  string
	InternalPkg string
	Services    map[string]*ThriftServiceDecl
	Structs     map[string]*ThriftStructDecl
	GoStructs   map[gocode.UserType]*ThriftStructDecl
}
```
## func NewThriftBuilder
```go
func NewThriftBuilder(code *goparser.ParsedModuleSet) *ThriftBuilder
```

## func 
```go
func (b *ThriftBuilder) AddService(iface *gocode.ServiceInterface) error
```
Adds a service declaration for the provided golang service interface.

## func 
```go
func (b *ThriftBuilder) GenerateMarshallingCode(outputFilePath string) error
```

## func 
```go
func (b *ThriftBuilder) GetOrAddMessage(t *gocode.UserType) (*ThriftStructDecl, error)
```

## func 
```go
func (b *ThriftBuilder) WriteThriftFile(outputFilePath string) error
```

```go
type ThriftField struct {
	SrcType      gocode.TypeName
	ThriftType   string
	ThriftGoType gocode.TypeName
	Name         string
	Position     int
}
```
## func 
```go
func (f *ThriftField) Marshall(imports *gogen.Imports, obj string, pkg string) (string, error)
```

## func 
```go
func (f *ThriftField) Unmarshall(imports *gogen.Imports, obj string, pkg string) (string, error)
```

```go
type ThriftMethodDecl struct {
	Service  *ThriftServiceDecl
	Name     string
	Request  *ThriftStructDecl
	Response *ThriftStructDecl
}
```
## func 
```go
func (m *ThriftMethodDecl) MarshallRequest(imports *gogen.Imports, pkg string) (string, error)
```

## func 
```go
func (m *ThriftMethodDecl) MarshallResponse(imports *gogen.Imports, pkg string) (string, error)
```

```go
type ThriftServiceDecl struct {
	Builder *ThriftBuilder
	Name    string
	Methods map[string]*ThriftMethodDecl
}
```
```go
type ThriftStructDecl struct {
	Builder    *ThriftBuilder
	Name       string
	ThriftType *gocode.UserType
	FieldList  []*ThriftField
}
```

