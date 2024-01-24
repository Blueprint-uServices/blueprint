package workflowspec

import "github.com/blueprint-uservices/blueprint/plugins/golang/goparser"

// Most plugins use a shared, cached workflow spec, so that we don't re-parse code constantly
var cached = New()

// Returns a shared / cached WorkflowSpec
func Get() *WorkflowSpec {
	return cached
}

// Parses & adds a module to the workflow spec search path
func AddModule(moduleName string) error {
	return cached.ParseModule(moduleName)
}

// Parses & adds the module containing T to the cached workflow spec.
func Add[T any]() error {
	modInfo, _, err := goparser.FindModule[T]()
	if err != nil {
		return err
	}
	_, err = cached.Modules.Add(modInfo)
	return err
}

// Gets a [WorkflowSpecService] for the specified type.
// Type parameter T should be a service defined in an application's workflow spec
// or a plugin's runtime directory.
//
// The definition of the service T will be acquired by parsing the module
// where T is defined.  Thus to utilize a specific version of T, set that
// version in the go.mod file of the wiring spec when requiring T's module.
//
// # Example Usage
//
//	leaf := workflowspec.GetService[leaf.LeafService]()
//
// # Internals
//
// By using type parameter T, it ensures that wherever T is defined, its module
// and version will be on the go path / within the go.mod.  By contrast,
// using [GetServiceByName] might fail if it names a package that doesn't exist
// in the local go cache / on the go path.
func GetService[T any]() (*Service, error) {
	modInfo, t, err := goparser.FindModule[T]()
	if err != nil {
		return nil, err
	}
	return cached.get(t.Package, t.Name, modInfo)
}

// Gets a [WorkflowSpecService] for the specified type.
// pkg and name should be the package and name of a service defined in an application's
// workflow spec or a plugin's runtime directory.
//
// # Example Usage
//
//	leaf := workflowspec.GetServiceByName("github.com/blueprint-uservices/blueprint/examples/leaf", "LeafService")
//
// # Internals
//
// This method is not as robust as [GetService] and it might fail if pkg isn't a local
// package or isn't a go.mod dependency.  Ensure the named package is in the go.mod file
// of the application.  Anonymously importing a package can help ensure it is not erased
// from your go.mod file, e.g.
//
//	import _ "github.com/blueprint-uservices/blueprint/examples/sockshop/tests"
func GetServiceByName(pkg, name string) (*Service, error) {
	modInfo, err := goparser.FindPackageModule(pkg)
	if err != nil {
		return nil, err
	}

	return cached.get(pkg, name, modInfo)
}
