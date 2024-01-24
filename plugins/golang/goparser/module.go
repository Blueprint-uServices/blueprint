package goparser

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/blueprint-uservices/blueprint/blueprint/pkg/blueprint"
	"github.com/blueprint-uservices/blueprint/plugins/golang/gocode"
	"golang.org/x/tools/go/packages"
)

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

func FindPackageModule(pkgName string) (*ModuleInfo, error) {
	if m, ok := cache[pkgName]; ok {
		return m, nil
	}

	pkgs, err := packages.Load(&packages.Config{Mode: packages.NeedModule}, pkgName)
	if err != nil {
		return nil, blueprint.Errorf("could not find package %v; is it in your go.mod? %v", pkgName, err)
	}
	if len(pkgs) != 1 {
		return nil, blueprint.Errorf("expected 1 package for %s, got %d", pkgName, len(pkgs))
	}
	mod := pkgs[0].Module
	if mod == nil {
		return nil, blueprint.Errorf("nil module for package %s", pkgName)
	}
	splits := strings.Split(pkgName, "/")
	info := &ModuleInfo{
		ShortName: splits[len(splits)-1],
		Path:      mod.Path,
		Version:   mod.Version,
		Dir:       mod.Dir,
		IsLocal:   isLocal(mod),
		GoModule:  mod,
	}
	cache[pkgName] = info
	return info, nil
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

func FindModule[T any]() (*ModuleInfo, *gocode.UserType, error) {
	t := reflect.TypeOf(new(T)).Elem()
	if t.PkgPath() == "" || t.Name() == "" {
		return nil, nil, blueprint.Errorf("type %v is predeclared or not defined and thus has no module", t)
	}
	mod, err := FindPackageModule(t.PkgPath())
	usertype := &gocode.UserType{Package: t.PkgPath(), Name: t.Name()}
	return mod, usertype, err
}

func print(mod *packages.Module) {
	fmt.Printf("Module %s\n", mod.Path)
	fmt.Printf("  Mod ver: %s\n", mod.Version)
	fmt.Printf("  Go mod: %s\n", mod.GoMod)
	fmt.Printf("  Go ver: %s\n", mod.GoVersion)
	fmt.Printf("  Mod dir: %s\n", mod.Dir)
	fmt.Printf("  GoMod dir: %s\n", mod.GoMod)
	fmt.Printf("  Is main: %v\n", mod.Main)
	fmt.Printf("  Is replace: %v\n", mod.Replace != nil)
	fmt.Printf("  Is indirect: %v\n", mod.Indirect)
}
