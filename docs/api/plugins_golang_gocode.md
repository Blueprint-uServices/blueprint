---
title: plugins/golang/gocode
---
# plugins/golang/gocode
```go
package gocode // import "gitlab.mpi-sws.org/cld/blueprint/plugins/golang/gocode"
```

## FUNCTIONS

## func IsBasicType
```go
func IsBasicType(name string) bool
```
## func IsBuiltinPackage
```go
func IsBuiltinPackage(packageName string) bool
```

## TYPES

The 'any' type which is just interface{}
```go
type AnyType struct {
	TypeName
}
```
## func 
```go
func (t *AnyType) IsTypeName()
```

## func 
```go
func (t *AnyType) String() string
```

Primitive types that don't need import statements
```go
type BasicType struct {
	TypeName
	Name string
}
```
## func 
```go
func (t *BasicType) IsTypeName()
```

## func 
```go
func (t *BasicType) String() string
```

Bidirectional Channel, e.g. chan string, chan *MyType
```go
type Chan struct {
	TypeName
	ChanOf TypeName
}
```
## func 
```go
func (t *Chan) IsTypeName()
```

## func 
```go
func (t *Chan) String() string
```

```go
type Constructor struct {
	Func
	Package string
}
```
Ellipsis type used in function arguments, e.g. ...string
```go
type Ellipsis struct {
	TypeName
	EllipsisOf TypeName // Elipsis of TypeName
}
```
## func 
```go
func (t *Ellipsis) IsTypeName()
```

## func 
```go
func (t *Ellipsis) String() string
```

```go
type Func struct {
	service.Method
	Name      string
	Arguments []Variable
	Returns   []Variable
}
```
## func 
```go
func (f *Func) AddArgument(variable Variable)
```

## func 
```go
func (f *Func) AddRetVar(variable Variable)
```

## func 
```go
func (f Func) Equals(g Func) bool
```

## func 
```go
func (f *Func) GetArguments() []service.Variable
```

## func 
```go
func (f *Func) GetName() string
```

## func 
```go
func (f *Func) GetReturns() []service.Variable
```

## func 
```go
func (f Func) String() string
```

A function signature. For now Blueprint doesn't support functions in service
method declarations, so we don't bother unravelling and representing the
function declaration here
```go
type FuncType struct {
	TypeName
}
```
## func 
```go
func (t *FuncType) IsTypeName()
```

## func 
```go
func (t *FuncType) String() string
```

A struct with generics. For now blueprint doesn't support generics in
service declarations
```go
type GenericType struct {
	TypeName
	BaseType TypeName
}
```
An interface of any kind. For now Blueprint doesn't support interfaces in
service method declarations, so we don't bother unravelling and representing
the interface declaration here
```go
type InterfaceType struct {
	TypeName
}
```
## func 
```go
func (t *InterfaceType) IsTypeName()
```

## func 
```go
func (t *InterfaceType) String() string
```

Map type, e.g. map[string]context.Context
```go
type Map struct {
	TypeName
	KeyType   TypeName
	ValueType TypeName
}
```
## func 
```go
func (t *Map) IsTypeName()
```

## func 
```go
func (m *Map) String() string
```

Pointer to a type, e.g. *string, *MyType, *context.Context
```go
type Pointer struct {
	TypeName
	PointerTo TypeName // Pointer to TypeName
}
```
## func 
```go
func (t *Pointer) IsTypeName()
```

## func 
```go
func (t *Pointer) String() string
```

Receive-only Channel, e.g. <-chan string, <-chan *MyType
```go
type ReceiveChan struct {
	TypeName
	ReceiveType TypeName
}
```
## func 
```go
func (t *ReceiveChan) IsTypeName()
```

## func 
```go
func (t *ReceiveChan) String() string
```

Send-only Channel, e.g. chan<- string, chan<- *MyType
```go
type SendChan struct {
	TypeName
	SendType TypeName
}
```
## func 
```go
func (t *SendChan) IsTypeName()
```

## func 
```go
func (t *SendChan) String() string
```

Implements service.ServiceInterface
```go
type ServiceInterface struct {
	UserType // Has a Name and a Source location
	BaseName string
	Methods  map[string]Func
}
```
## func CopyServiceInterface
```go
func CopyServiceInterface(name string, pkg string, s *ServiceInterface) *ServiceInterface
```

## func 
```go
func (s *ServiceInterface) AddMethod(f Func)
```

## func 
```go
func (s *ServiceInterface) GetMethods() []service.Method
```

## func 
```go
func (s *ServiceInterface) GetName() string
```

A slice or fixed-size array, e.g. []byte
```go
type Slice struct {
	TypeName
	SliceOf TypeName // Slice of TypeName
}
```
## func 
```go
func (t *Slice) IsTypeName()
```

## func 
```go
func (t *Slice) String() string
```

An inline struct of any kind. For now Blueprint doesn't support inline
structs in service method declarations, so we don't bother unravelling and
representing the struct here
```go
type StructType struct {
	TypeName
}
```
## func 
```go
func (t *StructType) IsTypeName()
```

## func 
```go
func (t *StructType) String() string
```

A type name is the fully qualified name of a type that you use when
declaring a variable, including possible imports and go.mod requires
```go
type TypeName interface {
	String() string
	IsTypeName()
}
```
A type that is declared in a module, thus requiring an import statement and
a go.mod requires statement
```go
type UserType struct {
	TypeName
	Package string
	Name    string // Name of the type within the package
}
```
## func 
```go
func (t *UserType) IsTypeName()
```

## func 
```go
func (t *UserType) String() string
```

```go
type Variable struct {
	service.Variable
	Name string
	Type TypeName
}
```
## func 
```go
func (v *Variable) GetName() string
```

## func 
```go
func (v *Variable) GetType() string
```

## func 
```go
func (v *Variable) String() string
```


