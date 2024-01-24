package goparser

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/blueprint-uservices/blueprint/blueprint/pkg/blueprint"
	"github.com/blueprint-uservices/blueprint/plugins/golang/gocode"
	"golang.org/x/tools/go/packages"
)

// Metadata about a module, such as its version and location on the local file system.
type ModuleInfo struct {
	ShortName string           // The last part of the module path
	Path      string           // Fully qualified name of the module
	Version   string           // Version of the module
	Dir       string           // Directory containing the module source
	IsLocal   bool             // True if the module is local (ie with a replace directive), false if it's from gocache
	GoModule  *packages.Module // The underlying golang module from [golang.org/x/tools/go/packages]
}

func (m *ModuleInfo) String() string {
	return fmt.Sprintf("Module %s Version: %s\n  IsLocal: %v, Dir: %s", m.Path, m.Version, m.IsLocal, m.Dir)
}

var cache = make(map[string]*ModuleInfo)

// Get the info for a module.  Better than reading the go.mod.
// Better than calling FindPackageModule because the root of the module
// doesn't need to be a golang package.
func GetModuleInfo(moduleName string) (*ModuleInfo, error) {
	if m, ok := cache[moduleName]; ok {
		return m, nil
	}

	if err := findModules(moduleName + "..."); err != nil {
		return nil, err
	}

	if m, ok := cache[moduleName]; ok {
		return m, nil
	}
	return nil, blueprint.Errorf("no valid module found for %v", moduleName)
}

// Get the module info for a package.
func FindPackageModule(pkgName string) (*ModuleInfo, error) {
	if m, ok := cache[pkgName]; ok {
		return m, nil
	}

	if err := findModules(pkgName); err != nil {
		return nil, err
	}

	if m, ok := cache[pkgName]; ok {
		if err := findModules(m.Path + "..."); err != nil {
			return nil, err
		}
		return m, nil
	}
	return nil, blueprint.Errorf("no valid module found for %v", pkgName)
}

func findModules(pkgName string) error {
	pkgs, err := packages.Load(&packages.Config{Mode: packages.NeedModule, Tests: true}, pkgName)
	if err != nil {
		return blueprint.Errorf("could not find package %v; is it in your go.mod? %v", pkgName, err)
	}

	for _, pkg := range pkgs {
		mod := pkg.Module
		if mod == nil {
			continue
		}
		splits := strings.Split(mod.Path, "/")
		info := &ModuleInfo{
			ShortName: splits[len(splits)-1],
			Path:      mod.Path,
			Version:   mod.Version,
			Dir:       mod.Dir,
			IsLocal:   isLocal(mod),
			GoModule:  mod,
		}
		cache[pkg.ID] = info
		cache[mod.Path] = info
	}
	return nil
}

// Somewhat hacky way of figuring out of the module is a local filesystem module vs. from the go cache.
// This is typically determined by the use of a replace directive, but for some reason the
// packages.Module replace isn't always true when using go workspaces
//
// Currently mod.Replace == true when it is a valid replace
// Sometimes mod.Main == true and mod.Replace == false when it is a replace through go.work
// Lastly we assume that go cache modules have a different path for the gomod than the dir, so if
// the gomod lives in the mod dir then it's a local module.
func isLocal(mod *packages.Module) bool {
	return mod.Main || mod.Replace != nil || strings.HasPrefix(mod.GoMod, mod.Dir)
}

// Finds and returns the module info for a type.
func FindModule[T any]() (*ModuleInfo, *gocode.UserType, error) {
	t := reflect.TypeOf(new(T)).Elem()
	// We also should support pointer types. This is necessary when the constructor of a service returns the type of a pointer to the service implementation instead of the interface.
	if t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	if t.PkgPath() == "" || t.Name() == "" {
		return nil, nil, blueprint.Errorf("type %v is predeclared or not defined and thus has no module", t)
	}
	mod, err := FindPackageModule(t.PkgPath())
	usertype := &gocode.UserType{Package: t.PkgPath(), Name: t.Name()}
	return mod, usertype, err
}
