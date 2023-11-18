---
title: plugins/grpc/grpccodegen
---
# plugins/grpc/grpccodegen
```go
package grpccodegen // import "gitlab.mpi-sws.org/cld/blueprint/plugins/grpc/grpccodegen"
```

## FUNCTIONS

## func CompileProtoFile
```go
func CompileProtoFile(protoFileName string) error
```
## func GenerateClient
```go
func GenerateClient(builder golang.ModuleBuilder, service *gocode.ServiceInterface, outputPackage string) error
```
This function is used by the GRPC plugin to generate the client-side caller
of the GRPC service.

It is assumed that outputPackage is the same as the one where the .proto is
generated to

## func GenerateGRPCProto
```go
func GenerateGRPCProto(builder golang.ModuleBuilder, service *gocode.ServiceInterface, outputPackage string) error
```
Generates the GRPC .proto file for the provided service interface, then
compiles it using `protoc`.

See the plugin README for the required GRPC and protocol buffers package
dependencies.

## func GenerateServerHandler
```go
func GenerateServerHandler(builder golang.ModuleBuilder, service *gocode.ServiceInterface, outputPackage string) error
```
This function is used by the GRPC plugin to generate the server-side GRPC
service.

It is assumed that outputPackage is the same as the one where the .proto is
generated to


## TYPES

A basic structural representation of the GRPC messages and services
```go
type GRPCField struct {
	SrcType   gocode.TypeName // The source type
	ProtoType string          // The GRPC type in proto
	GRPCType  gocode.TypeName // The GRPC type in golang
	Name      string
	Position  int
}
```
## func 
```go
func (f *GRPCField) Marshall(imports *gogen.Imports, obj string) (string, error)
```

## func 
```go
func (f *GRPCField) Unmarshall(imports *gogen.Imports, obj string) (string, error)
```

A basic structural representation of the GRPC messages and services
```go
type GRPCMessageDecl struct {
	Builder   *GRPCProtoBuilder
	Name      string
	GRPCType  *gocode.UserType // The GRPC-generated type for this message
	FieldList []*GRPCField
}
```
A basic structural representation of the GRPC messages and services
```go
type GRPCMethodDecl struct {
	Service  *GRPCServiceDecl
	Name     string
	Request  *GRPCMessageDecl
	Response *GRPCMessageDecl
}
```
A basic structural representation of the GRPC messages and services
```go
type GRPCProtoBuilder struct {
	Code        *goparser.ParsedModuleSet
	Package     string // Package shortname
	Module      golang.ModuleInfo
	PackageName string // Fully qualified package
	Services    map[string]*GRPCServiceDecl
	Messages    map[string]*GRPCMessageDecl
	Structs     map[gocode.UserType]*GRPCMessageDecl // Mapping from golang struct to the corresponding message
}
```
## func NewProtoBuilder
```go
func NewProtoBuilder(code *goparser.ParsedModuleSet) *GRPCProtoBuilder
```

## func 
```go
func (b *GRPCProtoBuilder) AddService(iface *gocode.ServiceInterface) error
```
Adds a service declaration for the provided golang service interface.

# This will create message and service definitions within the grpc proto

For arguments and return values on methods in the interface, corresponding
GRPC message objects are needed. The ProtoBuilder will consult the parsed
code to find the definitions of arguments and return values.

## func 
```go
func (b *GRPCProtoBuilder) GenerateMarshallingCode(outputFilePath string) error
```

## func 
```go
func (b *GRPCProtoBuilder) GetOrAddMessage(t *gocode.UserType) (*GRPCMessageDecl, error)
```

## func 
```go
func (b *GRPCProtoBuilder) WriteProtoFile(outputFilePath string) error
```

A basic structural representation of the GRPC messages and services
```go
type GRPCServiceDecl struct {
	Builder *GRPCProtoBuilder
	Name    string
	Methods map[string]*GRPCMethodDecl
}
```

