package parser2

import (
	"fmt"
	"reflect"
)

/*
Workflow spec related extensions to the parsed code
*/

func (modules *ParsedModuleSet) FindService(name string) (*Service, error) {
	for _, mod := range modules.Modules {
		for _, pkg := range mod.Packages {
			if iface, isIface := pkg.Interfaces[name]; isIface {
				valid, err := iface.IsValidService()
				if !valid {
					return nil, fmt.Errorf("found interface %v but it is not a valid service because %v", name, err.Error())
				}
				impls := iface.FindImplementations()
				fmt.Printf("Found %v impls\n", len(impls))
			}
		}
	}
	for _, mod := range modules.Modules {
		for _, pkg := range mod.Packages {
			for _, f := range pkg.Funcs {
				iface, err := f.IsConstructor()
				if err != nil {
					fmt.Printf("Not a constructor: %v is not a constructor due to %v\n", f.Func, err.Error())
				} else {
					fmt.Printf("Found constructor of %v:   %v\n", iface.Name, f.Func)
				}
			}
		}
	}
	return nil, nil
}

/*
A service interface is only valid if all methods receive ctx as
first argument and return error as final retval
*/
func (iface *ParsedInterface) IsValidService() (bool, error) {
	for _, method := range iface.Methods {
		if len(method.Arguments) == 0 {
			return false, fmt.Errorf("first argument of %v.%v must be context.Context", iface.Name, method.Name)
		}
		arg0, isBuiltIn := method.Arguments[0].Type.(*BuiltinType)
		fmt.Println("first arg is " + method.Arguments[0].Type.String())
		if !isBuiltIn || arg0.Package != "context" || arg0.Name != "Context" {
			return false, fmt.Errorf("first argument of %v.%v must be context.Context", iface.Name, method.Name)
		}
		if len(method.Returns) == 0 {
			return false, fmt.Errorf("last retval of %v.%v must be error", iface.Name, method.Name)
		}
		retL, isBasic := method.Returns[len(method.Returns)-1].Type.(*BasicType)
		if !isBasic || retL.Name != "error" {
			return false, fmt.Errorf("last retval of %v.%v must be error", iface.Name, method.Name)
		}
		// TODO: could potentially validate the serializability of args here
		// TODO: handling chans for asynchronous calls
	}
	return true, nil
}

func (iface *ParsedInterface) FindImplementations() []*ParsedStruct {
	var impls []*ParsedStruct
	for _, mod := range iface.File.Package.Module.ModuleSet.Modules {
		for _, pkg := range mod.Packages {
			for _, struc := range pkg.Structs {
				valid, err := struc.Implements(iface)
				if valid {
					fmt.Printf("%v implements %v\n", struc.Name, iface.Name)
					impls = append(impls, struc)
				} else {
					fmt.Println(err.Error())
				}
			}
		}
	}
	return impls
}

func (struc *ParsedStruct) FindInterfaces() []*ParsedInterface {
	var ifaces []*ParsedInterface
	for _, mod := range struc.File.Package.Module.ModuleSet.Modules {
		for _, pkg := range mod.Packages {
			for _, iface := range pkg.Interfaces {
				valid, err := struc.Implements(iface)
				if valid {
					fmt.Printf("%v implements %v\n", struc.Name, iface.Name)
					ifaces = append(ifaces, iface)
				} else {
					fmt.Println(err.Error())
				}
			}
		}
	}
	return ifaces
}

func (struc *ParsedStruct) Implements(iface *ParsedInterface) (bool, error) {
outer:
	for _, method1 := range iface.Methods {
		for _, method2 := range struc.Methods {
			if reflect.DeepEqual(method1.Func, method2.Func) {
				continue outer
			}
		}
		// no match found
		return false, fmt.Errorf("struct %v does not implement %v because it lacks %v", struc.Name, iface.Name, method1.Func.String())
	}
	return true, nil
}

func (f *ParsedFunc) IsConstructor() (*ParsedInterface, error) {
	if len(f.Returns) == 0 {
		return nil, fmt.Errorf("%v has no return values", f.Func)
	}
	if len(f.Returns) == 1 {
		return nil, fmt.Errorf("%v must return two values but only returns one", f.Func)
	}
	if len(f.Returns) > 2 {
		return nil, fmt.Errorf("%v has too many return values", f.Func)
	}

	if ret1, isRet1Basic := f.Returns[1].Type.(*BasicType); !isRet1Basic || ret1.Name != "error" {
		return nil, fmt.Errorf("second retval of %v must be error", f.Func)
	}

	switch ret0 := f.Returns[0].Type.(type) {
	case *UserType:
		{
			return f.File.Package.Module.ModuleSet.GetInterface(ret0)
		}
	case *Pointer:
		{
			dst, isValid := ret0.PointerTo.(*UserType)
			if !isValid {
				return nil, fmt.Errorf("%v must be pointer to user type", f.Func)
			}
			struc, err := f.File.Package.Module.ModuleSet.GetStruct(dst)
			if err != nil {
				return nil, err
			}
			ifaces := struc.FindInterfaces()
			if len(ifaces) == 0 {
				return nil, fmt.Errorf("%v does not implement a service interface", struc.Name)
			}
			return ifaces[0], nil
		}
	default:
		return nil, fmt.Errorf("invalid return type to be a constructor")
	}
}

func (modules *ParsedModuleSet) GetStruct(t *UserType) (*ParsedStruct, error) {
	mod, modExists := modules.Modules[t.ModuleName]
	if !modExists {
		return nil, fmt.Errorf("%v defined in external module %v", t, t.ModuleName)
	}
	pkg, pkgExists := mod.Packages[t.PackageName]
	if !pkgExists {
		return nil, fmt.Errorf("%v has invalid pkg %v", t, t.PackageName)
	}
	struc, strucExists := pkg.Structs[t.Name]
	if strucExists {
		return struc, nil
	}
	return nil, fmt.Errorf("%v does not exist in package %v", t.Name, t.PackageName)
}

func (modules *ParsedModuleSet) GetInterface(t *UserType) (*ParsedInterface, error) {
	mod, modExists := modules.Modules[t.ModuleName]
	if !modExists {
		return nil, fmt.Errorf("%v defined in external module %v", t, t.ModuleName)
	}
	pkg, pkgExists := mod.Packages[t.PackageName]
	if !pkgExists {
		return nil, fmt.Errorf("%v has invalid pkg %v", t, t.PackageName)
	}
	iface, ifaceExists := pkg.Interfaces[t.Name]
	if ifaceExists {
		return iface, nil
	}
	return nil, fmt.Errorf("%v does not exist in package %v", t.Name, t.PackageName)
}

func (f *ParsedFunc) IsConstructorFor(iface *ParsedInterface) (bool, error) {
	return false, nil
}
