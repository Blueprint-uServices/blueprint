package gocode

import (
	"fmt"
	"reflect"
	"strings"

	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/core/service"
)

/*
Basic structs used by IR nodes to describe Golang service interfaces

These structs implement the generic interfaces described in the core 'service' package

TypeName is defined separately in typename.go
*/

type (
	Variable struct {
		service.Variable
		Name string
		Type TypeName
	}

	Func struct {
		service.Method
		Name      string
		Arguments []Variable
		Returns   []Variable
	}

	Constructor struct {
		Func
		Package string
	}

	// Implements service.ServiceInterface
	ServiceInterface struct {
		UserType // Has a Name and a Source location
		Methods  map[string]Func
	}
)

func (s *ServiceInterface) GetName() string {
	return s.UserType.Name
}

func (s *ServiceInterface) GetMethods() []service.Method {
	var methods []service.Method
	for _, method := range s.Methods {
		methods = append(methods, &method)
	}
	return methods
}

func (f *Func) GetName() string {
	return f.Name
}

func (f *Func) GetArguments() []service.Variable {
	var variables []service.Variable
	for _, variable := range f.Arguments {
		variables = append(variables, &variable)
	}
	return variables
}

func (f *Func) GetReturns() []service.Variable {
	var variables []service.Variable
	for _, variable := range f.Returns {
		variables = append(variables, &variable)
	}
	return variables
}

func (v *Variable) GetName() string {
	return v.Name
}

func (v *Variable) GetType() string {
	return v.Type.String()
}

func sameTypes(a []Variable, b []Variable) bool {
	if len(a) != len(b) {
		return false
	}
	for i, va := range a {
		if !reflect.DeepEqual(va.Type, b[i].Type) {
			return false
		}
	}
	return true
}

func (f Func) Equals(g Func) bool {
	return f.Name == g.Name && sameTypes(f.Arguments, g.Arguments) && sameTypes(f.Returns, g.Returns)
}

func (v *Variable) String() string {
	if v.Name == "" {
		return v.Type.String()
	} else {
		return fmt.Sprintf("%v %v", v.Name, v.Type)
	}
}

func (f Func) String() string {
	var arglist []string
	for _, arg := range f.Arguments {
		arglist = append(arglist, arg.String())
	}
	args := strings.Join(arglist, ", ")
	var retlist []string
	for _, ret := range f.Returns {
		retlist = append(retlist, ret.String())
	}
	rets := strings.Join(retlist, ", ")
	if len(f.Returns) > 1 {
		return fmt.Sprintf("func %v(%v) (%v)", f.Name, args, rets)
	} else if len(f.Returns) == 1 {
		return fmt.Sprintf("func %v(%v) %v", f.Name, args, rets)
	} else {
		return fmt.Sprintf("func %v(%v)", f.Name, args)
	}
}
