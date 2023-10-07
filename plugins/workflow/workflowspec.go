package workflow

import (
	"fmt"
	"reflect"

	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/blueprint"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/golang/gocode"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/golang/goparser"
	"golang.org/x/exp/slog"
)

/*
Representation of a workflow spec.

This code makes heavy use of the Golang code parser defined in the Golang plugin.  That
code parser extracts structs, interfaces, and function definitions from a set of golang
modules.

This code adds functionality that:
  - Identifies valid service interfaces
  - Matches structs to interfaces that they implement
  - Finds constructors of structs
*/
type WorkflowSpec struct {
	Parsed *goparser.ParsedModuleSet
}

type WorkflowSpecService struct {
	Iface       *goparser.ParsedInterface
	Constructor *goparser.ParsedFunc
}

/*
Parses the specified module directories and loads workflow specs from there.

This will return an error if *any* of the provided srcModuleDirs are not valid Go modules
*/
func NewWorkflowSpec(srcModuleDirs ...string) (*WorkflowSpec, error) {
	parsed, err := goparser.ParseModules(srcModuleDirs...)
	if err != nil {
		return nil, err
	}
	spec := &WorkflowSpec{}
	spec.Parsed = parsed
	return spec, nil
}

/*
Looks up the named service in the workflow spec.  When a wiring spec instantiates a workflow spec
service, this method will ultimately get called.

Returns the service and a constructor
*/
func (spec *WorkflowSpec) Get(name string) (*WorkflowSpecService, error) {
	// Allowed to name an interface or a struct; different logic depending on which was named.

	for _, mod := range spec.Parsed.Modules {
		for _, pkg := range mod.Packages {
			if iface, isIface := pkg.Interfaces[name]; isIface {
				return spec.makeServiceFromInterface(iface)
			}
			if struc, isStruct := pkg.Structs[name]; isStruct {
				return spec.makeServiceFromStruct(struc)
			}
		}
	}

	return nil, blueprint.Errorf("unable to find service %v in workflow spec", name)
}

func (spec *WorkflowSpec) makeServiceFromStruct(struc *goparser.ParsedStruct) (*WorkflowSpecService, error) {
	ifaces := spec.findInterfacesFor(struc)
	if len(ifaces) == 0 {
		return nil, blueprint.Errorf("unable to find service interfaces for %v", struc.Name)
	}

	// The struct might implement many interfaces, but some of them might not be services
	var validIfaces []*goparser.ParsedInterface
	var errors []string
	for _, iface := range ifaces {
		if valid, err := isInterfaceAValidService(iface); !valid {
			errors = append(errors, err.Error())
		} else {
			validIfaces = append(validIfaces, iface)
		}
	}

	// Make sure we still have a valid interface
	if len(validIfaces) == 0 {
		if len(errors) == 1 {
			return nil, blueprint.Errorf("%v implements %v interfaces but none were valid, errors %v", struc.Name, len(ifaces), errors)
		}
		return nil, blueprint.Errorf("%v implements %v but it is not a valid service due to %v", struc.Name, ifaces[0].Name, errors[0])
	}
	if len(validIfaces) > 1 {
		slog.Warn(fmt.Sprintf("Warning: struct %v implements more than one service interface; using %v", struc.Name, validIfaces[0]))
	}

	// Find constructors
	constructors := spec.findConstructorsOfStruct(struc)
	if len(constructors) == 0 {
		return nil, blueprint.Errorf("no constructors for %v could be found, ie. funcs returning (*%v, error)", struc.Name, struc.Type().String())
	}
	if len(constructors) > 1 {
		slog.Warn(fmt.Sprintf("multiple constructors of struct %v found; using %v", struc.Name, constructors[0].Name))
	}

	service := &WorkflowSpecService{
		Iface:       validIfaces[0],
		Constructor: constructors[0],
	}
	slog.Info(fmt.Sprintf("Located workflow spec service %v with constructor %v in package %v\n", struc.Name, constructors[0].Name, validIfaces[0].File.Package.Name))
	return service, nil
}

func (spec *WorkflowSpec) makeServiceFromInterface(iface *goparser.ParsedInterface) (*WorkflowSpecService, error) {
	valid, err := isInterfaceAValidService(iface)
	if !valid {
		return nil, blueprint.Errorf("interface %v is not a valid service because %v", iface.Name, err.Error())
	}
	constructors := spec.findConstructorsOfIface(iface)
	if len(constructors) == 0 {
		return nil, blueprint.Errorf("found interface %v in %v but could not find any constructor methods", iface.Name, iface.File.Package.Name)
	}
	if len(constructors) > 1 {
		slog.Warn(fmt.Sprintf("multiple constructors of interface %v found; using %v", iface.Name, constructors[0].Name))
	}
	service := &WorkflowSpecService{
		Iface:       iface,
		Constructor: constructors[0],
	}
	slog.Info(fmt.Sprintf("Located workflow spec service %v with constructor %v in package %v\n", iface.Name, constructors[0].Name, iface.File.Package.Name))
	return service, nil

}

/*
A service interface is only valid if all methods receive ctx as
first argument and return error as final retval
*/
func isInterfaceAValidService(iface *goparser.ParsedInterface) (bool, error) {
	for _, method := range iface.Methods {
		if len(method.Arguments) == 0 {
			return false, blueprint.Errorf("first argument of %v.%v must be context.Context", iface.Name, method.Name)
		}
		arg0, isUserType := method.Arguments[0].Type.(*gocode.UserType)
		if !isUserType || arg0.Package != "context" || arg0.Name != "Context" {
			return false, blueprint.Errorf("first argument of %v.%v must be context.Context", iface.Name, method.Name)
		}
		if len(method.Returns) == 0 {
			return false, blueprint.Errorf("last retval of %v.%v must be error", iface.Name, method.Name)
		}
		retL, isBasic := method.Returns[len(method.Returns)-1].Type.(*gocode.BasicType)
		if !isBasic || retL.Name != "error" {
			return false, blueprint.Errorf("last retval of %v.%v must be error", iface.Name, method.Name)
		}
		// TODO: could potentially validate the serializability of args here
		// TODO: handling chans for asynchronous calls
	}
	return true, nil
}

/*
For a parsed struct, finds all valid interfaces that the struct implements
*/
func (spec *WorkflowSpec) findInterfacesFor(struc *goparser.ParsedStruct) []*goparser.ParsedInterface {
	var ifaces []*goparser.ParsedInterface
	for _, mod := range spec.Parsed.Modules {
		for _, pkg := range mod.Packages {
			for _, iface := range pkg.Interfaces {
				if valid, _ := implements(struc, iface); valid {
					ifaces = append(ifaces, iface)
				}
			}
		}
	}
	return ifaces
}

/*
Determines if the given struct implements the given interface
*/
func implements(struc *goparser.ParsedStruct, iface *goparser.ParsedInterface) (bool, error) {
	for name, method1 := range iface.Methods {
		method2, exists := struc.Methods[name]
		if !exists {
			return false, blueprint.Errorf("struct %v does not implement %v because it lacks %v", struc.Name, iface.Name, method1.Func.String())
		}
		if !method1.Func.Equals(method2.Func) {
			return false, blueprint.Errorf("struct %v does not implement %v because it has a different method signature for %v", struc.Name, iface.Name, method2.Func.String())
		}
	}
	return true, nil
}

func (spec *WorkflowSpec) findConstructorsOfIface(iface *goparser.ParsedInterface) []*goparser.ParsedFunc {
	var constructors []*goparser.ParsedFunc
	for _, mod := range spec.Parsed.Modules {
		for _, pkg := range mod.Packages {
			for _, f := range pkg.Funcs {
				if isConstructorOfIface(f, iface) {
					constructors = append(constructors, f)
				}
			}
		}
	}
	return constructors
}

func (spec *WorkflowSpec) findConstructorsOfStruct(struc *goparser.ParsedStruct) []*goparser.ParsedFunc {
	var constructors []*goparser.ParsedFunc
	for _, mod := range spec.Parsed.Modules {
		for _, pkg := range mod.Packages {
			for _, f := range pkg.Funcs {
				if isConstructorOfStruct(f, struc) {
					constructors = append(constructors, f)
				}
			}
		}
	}
	return constructors
}

/*
Currently Blueprint is strict about the method signatures of constructors:
  - the first argument must be context.Context
  - the final retval must be error
*/
func validateConstructorSignature(f *goparser.ParsedFunc) error {
	if len(f.Returns) == 0 {
		return blueprint.Errorf("%v has no return values", f.Func)
	}
	if len(f.Returns) == 1 {
		return blueprint.Errorf("%v must return two values but only returns one", f.Func)
	}
	if len(f.Returns) > 2 {
		return blueprint.Errorf("%v has too many return values", f.Func)
	}

	if ret1, isRet1Basic := f.Returns[1].Type.(*gocode.BasicType); !isRet1Basic || ret1.Name != "error" {
		return blueprint.Errorf("second retval of %v must be error", f.Func)
	}

	if len(f.Arguments) == 0 {
		return blueprint.Errorf("%v has no arguments", f.Func)
	}

	if arg0, isArg0User := f.Arguments[0].Type.(*gocode.UserType); !isArg0User || arg0.Name != "Context" || arg0.Package != "context" {
		return blueprint.Errorf("first argument of %v must be context.Context", f.Func)
	}

	return nil
}

func isConstructorOfIface(f *goparser.ParsedFunc, iface *goparser.ParsedInterface) bool {
	err := validateConstructorSignature(f)
	if err != nil {
		return false
	}

	// For an interface, retval can only be a usertype
	retType, valid := f.Returns[0].Type.(*gocode.UserType)
	if !valid {
		return false
	}

	// See if the iface is the return type
	if !reflect.DeepEqual(retType, iface.Type()) {
		return false
	}

	return true
}

func isConstructorOfStruct(f *goparser.ParsedFunc, struc *goparser.ParsedStruct) bool {
	err := validateConstructorSignature(f)
	if err != nil {
		return false
	}

	// For a struct, retval can only be a pointer to usertype
	retPointer, validPointer := f.Returns[0].Type.(*gocode.Pointer)
	if !validPointer {
		return false
	}

	// If the return value is a pointer, then it must be a pointer to an usertype implementation
	retType, isValid := retPointer.PointerTo.(*gocode.UserType)
	if !isValid {
		return false
	}

	// See if the iface is the return type
	if !reflect.DeepEqual(retType, struc.Type()) {
		return false
	}

	return true
}
