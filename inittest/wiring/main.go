package main

import (
	"fmt"
	"reflect"

	_ "github.com/blueprint-uservices/blueprint/examples/dsb_hotel/workflow/hotelreservation"
	"github.com/blueprint-uservices/blueprint/examples/leaf/wiring/specs"
	"github.com/blueprint-uservices/blueprint/examples/leaf/workflow/leaf"
	"github.com/blueprint-uservices/blueprint/examples/sockshop/workflow/user"
	"github.com/blueprint-uservices/blueprint/inittest/workflow/blah"
	"github.com/blueprint-uservices/blueprint/inittest/workflow/more"
	"github.com/blueprint-uservices/blueprint/inittest/workflow2/again"
	blah2 "github.com/blueprint-uservices/blueprint/inittest/workflow2/blah"
	"github.com/blueprint-uservices/blueprint/plugins/cmdbuilder"
	"github.com/blueprint-uservices/blueprint/plugins/golang/goparser"
	"golang.org/x/tools/go/packages"
)

type mystruc struct{}

func getFullyQualifiedName[T any]() (string, string) {
	t := reflect.TypeOf(new(T)).Elem()
	return t.PkgPath(), t.Name()
}

func getModuleInfo(pkgName string) (*packages.Module, error) {
	pkgs, err := packages.Load(&packages.Config{Mode: packages.NeedModule}, pkgName)
	if err != nil {
		return nil, err
	}
	if len(pkgs) != 1 {
		return nil, fmt.Errorf("expected 1 package for %s, got %d", pkgName, len(pkgs))
	}
	return pkgs[0].Module, nil
}

func print[T any]() {
	m, usertype, err := goparser.FindModule[T]()
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println(usertype, usertype.Package, usertype.Name)
	fmt.Println(m)
	fmt.Println()
}

func main() {

	print[leaf.LeafService]()
	print[mystruc]()
	print[blah.Blah]()
	print[blah2.Blah]()
	print[again.Again]()
	print[more.More]()
	print[user.UserService]()
	print[string]()

	mod, err := goparser.FindPackageModule("github.com/blueprint-uservices/blueprint/examples/dsb_hotel/workflow/hotelreservation")
	fmt.Println(mod, err)

	// set := goparser.Cache()
	// pkg, err := set.GetPackage("github.com/blueprint-uservices/blueprint/examples/dsb_hotel/workflow/hotelreservation")
	// fmt.Println(pkg, err)

	/*
		Next steps:

		  * put the workflow init stuff in a sub package so that it's not exposed to wiring by default
		  * we continue to use the old parsing code
		  * re-parse from scratch each time, but this time only with one module (mostly) + blueprint runtime, or two modules if iface and impl are both specified.  if impl specified, need a second load of the iface package.  cache this.
		  * extend the module struct to indicate if it's a cached or local module, handling the inconsistency with mod.Main and mod.Replace (perhaps just check the dirname to see if it matches a go cache)
		  * extend goparser to parse the embedded interface names of structs.  use this first when matching impls to ifaces -- if an impl claims to implement the iface, just match it.  conversely if an impl is requested, assume its iface is the first one listed (if not specified) (could even add an 'other interfaces' field?)
		  * special case / brute-force / fallback: allow adding a remote module + version to workflow spec search path; allow adding a relative local module and change to a warning if it's not a module (and print that the setup might be bad)


		Optimizations:

		  * cache / dont re-parse parsed modules
	*/

	if false {
		// Configure the location of our workflow spec
		// workflow.Init("../workflow")

		// Build a supported wiring spec
		name := "LeafApp"
		cmdbuilder.MakeAndExecute(
			name,
			specs.Docker,
		)
	}
}
