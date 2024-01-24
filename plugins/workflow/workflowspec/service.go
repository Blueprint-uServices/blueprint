package workflowspec

import (
	"fmt"
	"reflect"

	"github.com/blueprint-uservices/blueprint/blueprint/pkg/blueprint"
	"github.com/blueprint-uservices/blueprint/plugins/golang"
	"github.com/blueprint-uservices/blueprint/plugins/golang/gocode"
	"github.com/blueprint-uservices/blueprint/plugins/golang/goparser"
	"golang.org/x/exp/slog"
)

// A service in the workflow spec
type Service struct {
	// The interface that the service implements
	Iface *goparser.ParsedInterface

	// The constructor func of the service
	Constructor *goparser.ParsedFunc
}

// Get all modules containing definitions for this service.
// Could be more than one if the interface and implementation are defined in separate modules.
func (s *Service) Modules() []*goparser.ParsedModule {
	return []*goparser.ParsedModule{s.Iface.File.Package.Module, s.Constructor.File.Package.Module}
}

func (s *Service) AddToModule(builder golang.ModuleBuilder) error {
	for _, mod := range s.Modules() {
		if !mod.IsLocal {
			if !builder.Visited(mod.Name) {
				if err := builder.Require(mod.Name, mod.Version); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func (s *Service) AddToWorkspace(builder golang.WorkspaceBuilder) error {
	for _, mod := range s.Modules() {
		if mod.IsLocal {
			if !builder.Visited(mod.Name) {
				_, err := builder.AddLocalModule(mod.ShortName, mod.SrcDir)
				return err
			}
		}
	}
	return nil
}

func (spec *WorkflowSpec) makeServiceFromStruct(struc *goparser.ParsedStruct) (*Service, error) {
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

	service := &Service{
		Iface:       validIfaces[0],
		Constructor: constructors[0],
	}
	slog.Info(fmt.Sprintf("Located %v (%v) in package %v", struc.Name, constructors[0].Name, validIfaces[0].File.Package.Name))
	return service, nil
}

func (spec *WorkflowSpec) makeServiceFromInterface(iface *goparser.ParsedInterface) (*Service, error) {
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
	service := &Service{
		Iface:       iface,
		Constructor: constructors[0],
	}
	slog.Info(fmt.Sprintf("Located %v (%v) in package %v", iface.Name, constructors[0].Name, iface.File.Package.Name))
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
	for _, mod := range spec.Modules.Modules {
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
	for _, mod := range spec.Modules.Modules {
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
	for _, mod := range spec.Modules.Modules {
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
