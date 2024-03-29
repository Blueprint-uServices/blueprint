<!-- Code generated by gomarkdoc. DO NOT EDIT -->

# thriftcodegen

```go
import "github.com/blueprint-uservices/blueprint/plugins/thrift/thriftcodegen"
```

## Index

- [func CompileThriftFile\(thriftFileName string\) error](<#CompileThriftFile>)
- [func GenerateClient\(builder golang.ModuleBuilder, service \*gocode.ServiceInterface, outputPackage string\) error](<#GenerateClient>)
- [func GenerateServerHandler\(builder golang.ModuleBuilder, service \*gocode.ServiceInterface, outputPackage string\) error](<#GenerateServerHandler>)
- [func GenerateThrift\(builder golang.ModuleBuilder, service \*gocode.ServiceInterface, outputPackage string\) error](<#GenerateThrift>)
- [type ThriftBuilder](<#ThriftBuilder>)
  - [func NewThriftBuilder\(code \*goparser.ParsedModuleSet\) \*ThriftBuilder](<#NewThriftBuilder>)
  - [func \(b \*ThriftBuilder\) AddService\(iface \*gocode.ServiceInterface\) error](<#ThriftBuilder.AddService>)
  - [func \(b \*ThriftBuilder\) GenerateMarshallingCode\(outputFilePath string\) error](<#ThriftBuilder.GenerateMarshallingCode>)
  - [func \(b \*ThriftBuilder\) GetOrAddMessage\(t \*gocode.UserType\) \(\*ThriftStructDecl, error\)](<#ThriftBuilder.GetOrAddMessage>)
  - [func \(b \*ThriftBuilder\) WriteThriftFile\(outputFilePath string\) error](<#ThriftBuilder.WriteThriftFile>)
- [type ThriftField](<#ThriftField>)
  - [func \(f \*ThriftField\) Marshall\(imports \*gogen.Imports, obj string, pkg string\) \(string, error\)](<#ThriftField.Marshall>)
  - [func \(f \*ThriftField\) Unmarshall\(imports \*gogen.Imports, obj string, pkg string\) \(string, error\)](<#ThriftField.Unmarshall>)
- [type ThriftMethodDecl](<#ThriftMethodDecl>)
  - [func \(m \*ThriftMethodDecl\) MarshallRequest\(imports \*gogen.Imports, pkg string\) \(string, error\)](<#ThriftMethodDecl.MarshallRequest>)
  - [func \(m \*ThriftMethodDecl\) MarshallResponse\(imports \*gogen.Imports, pkg string\) \(string, error\)](<#ThriftMethodDecl.MarshallResponse>)
- [type ThriftServiceDecl](<#ThriftServiceDecl>)
- [type ThriftStructDecl](<#ThriftStructDecl>)


<a name="CompileThriftFile"></a>
## func [CompileThriftFile](<https://github.com/blueprint-uservices/blueprint/blob/main/plugins/thrift/thriftcodegen/thriftgen.go#L79>)

```go
func CompileThriftFile(thriftFileName string) error
```



<a name="GenerateClient"></a>
## func [GenerateClient](<https://github.com/blueprint-uservices/blueprint/blob/main/plugins/thrift/thriftcodegen/clientgen.go#L17>)

```go
func GenerateClient(builder golang.ModuleBuilder, service *gocode.ServiceInterface, outputPackage string) error
```

This function is used by the Thrift plugin to generate the client\-side caller of the Thrift service.

It is assumed that outputPackage is the same as the one where the .thrift is generated to

<a name="GenerateServerHandler"></a>
## func [GenerateServerHandler](<https://github.com/blueprint-uservices/blueprint/blob/main/plugins/thrift/thriftcodegen/servergen.go#L17>)

```go
func GenerateServerHandler(builder golang.ModuleBuilder, service *gocode.ServiceInterface, outputPackage string) error
```

This function is used by the Thrift plugin to generate the server\-side Thrift service.

It is assumed that outputPackage is the same as the one where the .thrift is generated to

<a name="GenerateThrift"></a>
## func [GenerateThrift](<https://github.com/blueprint-uservices/blueprint/blob/main/plugins/thrift/thriftcodegen/thriftgen.go#L21>)

```go
func GenerateThrift(builder golang.ModuleBuilder, service *gocode.ServiceInterface, outputPackage string) error
```

Generates the .thrift file for the provided service interface, then compiles it using \`thrift\`. See the plugin README for the required thrift package dependencies.

<a name="ThriftBuilder"></a>
## type [ThriftBuilder](<https://github.com/blueprint-uservices/blueprint/blob/main/plugins/thrift/thriftcodegen/thriftgen.go#L125-L135>)



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

<a name="NewThriftBuilder"></a>
### func [NewThriftBuilder](<https://github.com/blueprint-uservices/blueprint/blob/main/plugins/thrift/thriftcodegen/thriftgen.go#L137>)

```go
func NewThriftBuilder(code *goparser.ParsedModuleSet) *ThriftBuilder
```



<a name="ThriftBuilder.AddService"></a>
### func \(\*ThriftBuilder\) [AddService](<https://github.com/blueprint-uservices/blueprint/blob/main/plugins/thrift/thriftcodegen/thriftgen.go#L221>)

```go
func (b *ThriftBuilder) AddService(iface *gocode.ServiceInterface) error
```

Adds a service declaration for the provided golang service interface.

<a name="ThriftBuilder.GenerateMarshallingCode"></a>
### func \(\*ThriftBuilder\) [GenerateMarshallingCode](<https://github.com/blueprint-uservices/blueprint/blob/main/plugins/thrift/thriftcodegen/marshallgen.go#L20>)

```go
func (b *ThriftBuilder) GenerateMarshallingCode(outputFilePath string) error
```

Generates marshalling functions that convert between Go objects and Thrift struct objects

This extends the code in thriftgen.go and is called from thriftgen.go

<a name="ThriftBuilder.GetOrAddMessage"></a>
### func \(\*ThriftBuilder\) [GetOrAddMessage](<https://github.com/blueprint-uservices/blueprint/blob/main/plugins/thrift/thriftcodegen/thriftgen.go#L241>)

```go
func (b *ThriftBuilder) GetOrAddMessage(t *gocode.UserType) (*ThriftStructDecl, error)
```



<a name="ThriftBuilder.WriteThriftFile"></a>
### func \(\*ThriftBuilder\) [WriteThriftFile](<https://github.com/blueprint-uservices/blueprint/blob/main/plugins/thrift/thriftcodegen/thriftgen.go#L164>)

```go
func (b *ThriftBuilder) WriteThriftFile(outputFilePath string) error
```



<a name="ThriftField"></a>
## type [ThriftField](<https://github.com/blueprint-uservices/blueprint/blob/main/plugins/thrift/thriftcodegen/thriftgen.go#L97-L103>)

A basic structural representation of the Thrift messages and services

```go
type ThriftField struct {
    SrcType      gocode.TypeName
    ThriftType   string
    ThriftGoType gocode.TypeName
    Name         string
    Position     int
}
```

<a name="ThriftField.Marshall"></a>
### func \(\*ThriftField\) [Marshall](<https://github.com/blueprint-uservices/blueprint/blob/main/plugins/thrift/thriftcodegen/marshallgen.go#L127>)

```go
func (f *ThriftField) Marshall(imports *gogen.Imports, obj string, pkg string) (string, error)
```



<a name="ThriftField.Unmarshall"></a>
### func \(\*ThriftField\) [Unmarshall](<https://github.com/blueprint-uservices/blueprint/blob/main/plugins/thrift/thriftcodegen/marshallgen.go#L182>)

```go
func (f *ThriftField) Unmarshall(imports *gogen.Imports, obj string, pkg string) (string, error)
```



<a name="ThriftMethodDecl"></a>
## type [ThriftMethodDecl](<https://github.com/blueprint-uservices/blueprint/blob/main/plugins/thrift/thriftcodegen/thriftgen.go#L112-L117>)



```go
type ThriftMethodDecl struct {
    Service  *ThriftServiceDecl
    Name     string
    Request  *ThriftStructDecl
    Response *ThriftStructDecl
}
```

<a name="ThriftMethodDecl.MarshallRequest"></a>
### func \(\*ThriftMethodDecl\) [MarshallRequest](<https://github.com/blueprint-uservices/blueprint/blob/main/plugins/thrift/thriftcodegen/marshallgen.go#L103>)

```go
func (m *ThriftMethodDecl) MarshallRequest(imports *gogen.Imports, pkg string) (string, error)
```



<a name="ThriftMethodDecl.MarshallResponse"></a>
### func \(\*ThriftMethodDecl\) [MarshallResponse](<https://github.com/blueprint-uservices/blueprint/blob/main/plugins/thrift/thriftcodegen/marshallgen.go#L115>)

```go
func (m *ThriftMethodDecl) MarshallResponse(imports *gogen.Imports, pkg string) (string, error)
```



<a name="ThriftServiceDecl"></a>
## type [ThriftServiceDecl](<https://github.com/blueprint-uservices/blueprint/blob/main/plugins/thrift/thriftcodegen/thriftgen.go#L119-L123>)



```go
type ThriftServiceDecl struct {
    Builder *ThriftBuilder
    Name    string
    Methods map[string]*ThriftMethodDecl
}
```

<a name="ThriftStructDecl"></a>
## type [ThriftStructDecl](<https://github.com/blueprint-uservices/blueprint/blob/main/plugins/thrift/thriftcodegen/thriftgen.go#L105-L110>)



```go
type ThriftStructDecl struct {
    Builder    *ThriftBuilder
    Name       string
    ThriftType *gocode.UserType
    FieldList  []*ThriftField
}
```

Generated by [gomarkdoc](<https://github.com/princjef/gomarkdoc>)
