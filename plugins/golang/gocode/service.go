// Package gocode defines basic structs used by IRNodes to describe Golang types, variables,
// funcs, constructors, and service interfaces.
package gocode

import (
	"fmt"
	"strings"

	"github.com/blueprint-uservices/blueprint/blueprint/pkg/coreplugins/service"
)

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
		BaseName string
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

func (s *ServiceInterface) AddMethod(f Func) {
	s.Methods[f.Name] = f
}

func CopyServiceInterface(name string, pkg string, s *ServiceInterface) *ServiceInterface {
	new_s := &ServiceInterface{UserType{Name: name, Package: pkg}, s.BaseName, make(map[string]Func)}

	for method_name, method := range s.Methods {
		new_s.Methods[method_name] = method
	}
	return new_s
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

func (f *Func) AddArgument(variable Variable) {
	f.Arguments = append(f.Arguments, variable)
}

func (f *Func) AddRetVar(variable Variable) {
	f.Returns = append(f.Returns, variable)
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
	for i := range a {
		if !a[i].Type.Equals(b[i].Type) {
			return false
		}
		// if !reflect.DeepEqual(va.Type, b[i].Type) {
		// 	return false
		// }
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

func (i *ServiceInterface) String() string {
	return i.UserType.String()
}

// Reports whether all of the methods in j exist on interface i
func (i *ServiceInterface) Contains(j *ServiceInterface) bool {
	if i == nil || j == nil {
		return false
	}
	for name, jFunc := range j.Methods {
		iFunc, iHasFunc := i.Methods[name]
		if !iHasFunc {
			return false
		}
		if !jFunc.Equals(iFunc) {
			return false
		}
	}
	return true
}
