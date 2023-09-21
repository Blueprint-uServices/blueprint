package gocode

import (
	"fmt"
	"strings"

	"golang.org/x/tools/go/packages"
)

// A type name is the fully qualified name of a type that you use when
// declaring a variable, including possible imports and go.mod requires
type TypeName interface {
	String() string
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

func IsBuiltinPackage(packageName string) bool {
	initBuiltins()
	_, ok := builtins[packageName]
	return ok
}

func IsBasicType(name string) bool {
	initBasics()
	_, ok := basics[name]
	return ok
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

func (t *BasicType) IsTypeName()     {}
func (t *UserType) IsTypeName()      {}
func (t *Slice) IsTypeName()         {}
func (t *Ellipsis) IsTypeName()      {}
func (t *Pointer) IsTypeName()       {}
func (t *Map) IsTypeName()           {}
func (t *Chan) IsTypeName()          {}
func (t *ReceiveChan) IsTypeName()   {}
func (t *SendChan) IsTypeName()      {}
func (t *AnyType) IsTypeName()       {}
func (t *InterfaceType) IsTypeName() {}
func (t *FuncType) IsTypeName()      {}
func (t *StructType) IsTypeName()    {}
