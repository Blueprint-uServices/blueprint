package gocode

import (
	"fmt"
	"reflect"
	"strings"

	"golang.org/x/tools/go/packages"
)

// A type name is the fully qualified name of a type that you use when
// declaring a variable, including possible imports and go.mod requires
type TypeName interface {
	String() string
	Equals(other TypeName) bool
	IsTypeName()
}

var builtins = make(map[string]struct{})
var basics = make(map[string]struct{})

func initBuiltins() {
	if len(builtins) > 0 {
		return
	}
	pkgs, err := packages.Load(nil, "std")
	if err != nil {
		panic(err)
	}

	for _, p := range pkgs {
		builtins[p.PkgPath] = struct{}{}
	}
}

func initBasics() {
	if len(basics) > 0 {
		return
	}
	names := []string{
		"bool",
		"string",
		"int", "int8", "int16", "int32", "int64",
		"uint", "uint8", "uint16", "uint32", "uint64",
		"byte",
		"rune",
		"float32", "float64",
		"complex64", "complex128",
		"error",
	}
	for _, name := range names {
		basics[name] = struct{}{}
	}
}

// Reports whether packageName is a builtin (e.g. "os", "context")
func IsBuiltinPackage(packageName string) bool {
	initBuiltins()
	_, ok := builtins[packageName]
	return ok
}

// Reports whether name is a basic type (e.g. "bool", "string", "int32", "float32", "rune", etc.)
func IsBasicType(name string) bool {
	initBasics()
	_, ok := basics[name]
	return ok
}

// Returns a [UserType] for type T,
func TypeOf[T any]() TypeName {
	return typeof(reflect.TypeOf(new(T)).Elem())
}

// Returns the unqualified shortname for type T
func NameOf[T any]() string {
	return reflect.TypeOf(new(T)).Elem().Name()
}

// converts a reflect.type into the janky typename stuff we have here
func typeof(t reflect.Type) TypeName {
	switch t.Kind() {
	case reflect.Array:
		return &Slice{SliceOf: typeof(t.Elem())}
	case reflect.Chan:
		{
			switch t.ChanDir() {
			case reflect.RecvDir:
				return &ReceiveChan{ReceiveType: typeof(t.Elem())}
			case reflect.SendDir:
				return &SendChan{SendType: typeof(t.Elem())}
			case reflect.BothDir:
				return &Chan{ChanOf: typeof(t.Elem())}
			}
		}
	case reflect.Func:
		return &FuncType{}
	case reflect.Interface:
		if t.Name() == "" {
			return &InterfaceType{}
		} else if t.PkgPath() == "" {
			return &BasicType{Name: t.Name()}
		} else {
			return &UserType{Package: t.PkgPath(), Name: t.Name()}
		}
	case reflect.Map:
		return &Map{KeyType: typeof(t.Key()), ValueType: typeof(t.Elem())}
	case reflect.Ptr:
		return &Pointer{PointerTo: typeof(t.Elem())}
	case reflect.Slice:
		return &Slice{SliceOf: typeof(t.Elem())}
	case reflect.Struct:
		if t.Name() == "" {
			return &StructType{}
		} else if t.PkgPath() == "" {
			return &BasicType{Name: t.Name()}
		} else {
			return &UserType{Package: t.PkgPath(), Name: t.Name()}
		}
	}

	if t.PkgPath() == "" {
		return &BasicType{Name: t.Name()}
	} else {
		return &UserType{Package: t.PkgPath(), Name: t.Name()}
	}
}

// Structs representing the different kinds of TypeName that you can have in Go
// User types include the module and package name for use with go.mod and import
// statements
type (

	/*
		Primitive types that don't need import statements
	*/
	BasicType struct {
		TypeName
		Name string
	}

	/*
		A type that is declared in a module, thus requiring an import statement and a
		go.mod requires statement
	*/
	UserType struct {
		TypeName
		Package string
		Name    string // Name of the type within the package
	}

	/*
		A slice or fixed-size array, e.g. []byte
	*/
	Slice struct {
		TypeName
		SliceOf TypeName // Slice of TypeName
	}

	/*
		Ellipsis type used in function arguments, e.g. ...string
	*/
	Ellipsis struct {
		TypeName
		EllipsisOf TypeName // Elipsis of TypeName
	}

	/*
		Pointer to a type, e.g. *string, *MyType, *context.Context
	*/
	Pointer struct {
		TypeName
		PointerTo TypeName // Pointer to TypeName
	}

	/*
		Map type, e.g. map[string]context.Context
	*/
	Map struct {
		TypeName
		KeyType   TypeName
		ValueType TypeName
	}

	/*
		Bidirectional Channel, e.g. chan string, chan *MyType
	*/
	Chan struct {
		TypeName
		ChanOf TypeName
	}

	/*
		Receive-only Channel, e.g. <-chan string, <-chan *MyType
	*/
	ReceiveChan struct {
		TypeName
		ReceiveType TypeName
	}

	/*
		Send-only Channel, e.g. chan<- string, chan<- *MyType
	*/
	SendChan struct {
		TypeName
		SendType TypeName
	}

	/*
		An interface of any kind.  For now Blueprint doesn't support
		interfaces in service method declarations, so we don't
		bother unravelling and representing the interface
		declaration here
	*/
	InterfaceType struct {
		TypeName
	}

	/*
		The 'any' type which is just interface{}
	*/
	AnyType struct {
		TypeName
	}

	/*
		A function signature.  For now Blueprint doesn't support
		functions in service method declarations, so we don't
		bother unravelling and representing the function
		declaration here
	*/
	FuncType struct {
		TypeName
	}

	/*
		An inline struct of any kind.  For now Blueprint doesn't
		support inline structs in service method declarations, so
		we don't bother unravelling and representing the struct here
	*/
	StructType struct {
		TypeName
	}

	/*
		A struct with generics. For now blueprint doesn't support generics in service declarations
	*/
	GenericType struct {
		TypeName
		BaseType  TypeName
		TypeParam TypeName
	}

	// The type parameter of a generic struct or func
	GenericTypeParam struct {
		TypeName
		ParamName string
	}
)

func shortName(Package string) string {
	splits := strings.Split(Package, "/")
	return splits[len(splits)-1]
}

func (t *BasicType) String() string {
	return t.Name
}

func (t *UserType) String() string {
	return fmt.Sprintf("%s.%s", shortName(t.Package), t.Name)
}

func (t *Slice) String() string {
	return fmt.Sprintf("[]%s", t.SliceOf)
}

func (t *Ellipsis) String() string {
	return fmt.Sprintf("...%s", t.EllipsisOf)
}

func (t *Pointer) String() string {
	return fmt.Sprintf("*%s", t.PointerTo)
}

func (m *Map) String() string {
	return fmt.Sprintf("map[%s]%s", m.KeyType, m.ValueType)
}

func (t *Chan) String() string {
	return fmt.Sprintf("chan %s", t.ChanOf)
}

func (t *ReceiveChan) String() string {
	return fmt.Sprintf("<-chan %s", t.ReceiveType)
}

func (t *SendChan) String() string {
	return fmt.Sprintf("chan<- %s", t.SendType)
}

func (t *InterfaceType) String() string {
	return "interface{.}"
}

func (t *AnyType) String() string {
	return "any"
}

func (t *FuncType) String() string {
	return "func(.){.}"
}

func (t *StructType) String() string {
	return "struct{...}"
}

func (t *GenericType) String() string {
	return fmt.Sprintf("%v[%v]", t.BaseType, t.TypeParam)
}

func (t *GenericTypeParam) String() string {
	return t.ParamName
}

func (t *BasicType) IsTypeName()        {}
func (t *UserType) IsTypeName()         {}
func (t *Slice) IsTypeName()            {}
func (t *Ellipsis) IsTypeName()         {}
func (t *Pointer) IsTypeName()          {}
func (t *Map) IsTypeName()              {}
func (t *Chan) IsTypeName()             {}
func (t *ReceiveChan) IsTypeName()      {}
func (t *SendChan) IsTypeName()         {}
func (t *AnyType) IsTypeName()          {}
func (t *InterfaceType) IsTypeName()    {}
func (t *FuncType) IsTypeName()         {}
func (t *StructType) IsTypeName()       {}
func (t *GenericType) IsTypeName()      {}
func (t *GenericTypeParam) IsTypeName() {}

func (t *BasicType) Equals(other TypeName) bool {
	if t == nil || other == nil {
		return false
	}
	t2, isSameType := other.(*BasicType)
	if !isSameType {
		return false
	}
	return t.Name == t2.Name
}

func (t *UserType) Equals(other TypeName) bool {
	if t == nil || other == nil {
		return false
	}
	t2, isSameType := other.(*UserType)
	if !isSameType {
		return false
	}
	return t.Package == t2.Package && t.Name == t2.Name
}

func (t *Slice) Equals(other TypeName) bool {
	if t == nil || other == nil {
		return false
	}
	t2, isSameType := other.(*Slice)
	if !isSameType {
		return false
	}
	return t.SliceOf.Equals(t2.SliceOf)
}

func (t *Ellipsis) Equals(other TypeName) bool {
	if t == nil || other == nil {
		return false
	}
	t2, isSameType := other.(*Ellipsis)
	if !isSameType {
		return false
	}
	return t.EllipsisOf.Equals(t2.EllipsisOf)
}

func (t *Pointer) Equals(other TypeName) bool {
	if t == nil || other == nil {
		return false
	}
	t2, isSameType := other.(*Pointer)
	if !isSameType {
		return false
	}
	return t.PointerTo.Equals(t2.PointerTo)
}

func (t *Map) Equals(other TypeName) bool {
	if t == nil || other == nil {
		return false
	}
	t2, isSameType := other.(*Map)
	if !isSameType {
		return false
	}
	return t.KeyType.Equals(t2.KeyType) && t.ValueType.Equals(t2.ValueType)
}

func (t *Chan) Equals(other TypeName) bool {
	if t == nil || other == nil {
		return false
	}
	t2, isSameType := other.(*Chan)
	if !isSameType {
		return false
	}
	return t.ChanOf.Equals(t2.ChanOf)
}

func (t *ReceiveChan) Equals(other TypeName) bool {
	if t == nil || other == nil {
		return false
	}
	t2, isSameType := other.(*ReceiveChan)
	if !isSameType {
		return false
	}
	return t.ReceiveType.Equals(t2.ReceiveType)
}

func (t *SendChan) Equals(other TypeName) bool {
	if t == nil || other == nil {
		return false
	}
	t2, isSameType := other.(*SendChan)
	if !isSameType {
		return false
	}
	return t.SendType.Equals(t2.SendType)
}

func (t *InterfaceType) Equals(other TypeName) bool {
	if t == nil || other == nil {
		return false
	}
	_, isSameType := other.(*InterfaceType)
	return isSameType // TODO: interface matching not implemented yet
}

func (t *AnyType) Equals(other TypeName) bool {
	if t == nil || other == nil {
		return false
	}
	_, isSameType := other.(*AnyType)
	return isSameType // TODO: interface matching not implemented yet
}

func (t *FuncType) Equals(other TypeName) bool {
	if t == nil || other == nil {
		return false
	}
	_, isSameType := other.(*FuncType)
	return isSameType // TODO: functype matching not implemented yet
}

func (t *StructType) Equals(other TypeName) bool {
	if t == nil || other == nil {
		return false
	}
	_, isSameType := other.(*StructType)
	return isSameType // TODO: StructType matching not implemented yet
}

func (t *GenericType) Equals(other TypeName) bool {
	if t == nil || other == nil {
		return false
	}
	t2, isSameType := other.(*GenericType)
	if !isSameType {
		return false
	}
	return t.BaseType.Equals(t2.BaseType) && t.TypeParam.Equals(t2.TypeParam)
}

func (t *GenericTypeParam) Equals(other TypeName) bool {
	if t == nil || other == nil {
		return false
	}
	t2, isSameType := other.(*GenericTypeParam)
	if !isSameType {
		return false
	}
	return t.ParamName == t2.ParamName
}
