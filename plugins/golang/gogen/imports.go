package gogen

import (
	"fmt"
	"strings"

	"gitlab.mpi-sws.org/cld/blueprint/plugins/golang/gocode"
)

/*
A helper struct for managing imports in generated golang files.

Used by plugins like the GRPC plugin.

The string representation of the Imports struct is
the import declaration.

The NameOf method provides the correctly qualified name for
the specified userType
*/
type Imports struct {
	localPackage string              // The name of the package where the importing is happening
	packages     map[string]string   // Map from fully qualified package name to imported name
	named        map[string]string   // Map of fully qualified package names to chosen import name
	anonymous    map[string]struct{} // Map of fully qualified package names with default import name
	seen         map[string]struct{} // Map from imported name to fully qualified package name
}

/*
Creates a new ImportedPackages struct, treating the provided
fully-qualified packageName as the "current" package
*/
func NewImports(packageName string) *Imports {
	imports := &Imports{}
	imports.localPackage = packageName
	imports.packages = make(map[string]string)
	imports.named = make(map[string]string)
	imports.anonymous = make(map[string]struct{})
	imports.seen = make(map[string]struct{})
	return imports
}

func (imports *Imports) AddPackage(pkg string) string {
	if pkg == imports.localPackage {
		return ""
	}
	if importName, exists := imports.packages[pkg]; exists {
		return importName
	}

	splits := strings.Split(pkg, "/")
	defaultImportName := splits[len(splits)-1]
	importName := defaultImportName
	i := 2
	for {
		if _, exists := imports.seen[importName]; exists {
			importName = fmt.Sprintf("%v%v", defaultImportName, i)
			i += 1
		} else {
			imports.packages[pkg] = importName
			imports.seen[importName] = struct{}{}
			if importName == defaultImportName {
				imports.anonymous[pkg] = struct{}{}
			} else {
				imports.named[pkg] = importName
			}
			return importName
		}
	}
}

func (imports *Imports) AddType(typeName gocode.TypeName) {
	switch t := typeName.(type) {
	case *gocode.BuiltinType:
		{
			imports.AddPackage(t.Package)
		}
	case *gocode.UserType:
		{
			imports.AddPackage(t.PackageName)
		}
	case *gocode.Pointer:
		{
			imports.AddType(t.PointerTo)
		}
	case *gocode.Slice:
		{
			imports.AddType(t.SliceOf)
		}
	case *gocode.Map:
		{
			imports.AddType(t.KeyType)
			imports.AddType(t.ValueType)
		}
	case *gocode.Chan:
		{
			imports.AddType(t.ChanOf)
		}
	case *gocode.ReceiveChan:
		{
			imports.AddType(t.ReceiveType)
		}
	case *gocode.SendChan:
		{
			imports.AddType(t.SendType)
		}
	}
}

func (imports *Imports) String() string {
	var b strings.Builder
	b.WriteString("import (\n")
	for pkg, _ := range imports.anonymous {
		b.WriteString(fmt.Sprintf("\t\"%s\"\n", pkg))
	}
	for pkg, importName := range imports.named {
		b.WriteString(fmt.Sprintf("\t%s \"%s\"\n", importName, pkg))
	}
	b.WriteString(")")
	return b.String()
}

func (imports *Imports) Qualify(pkg string, name string) string {
	if pkg == imports.localPackage {
		return name
	}
	if importedName, isImported := imports.packages[pkg]; isImported {
		return importedName + "." + name
	}
	return pkg + "." + name
}

func (imports *Imports) NameOf(typeName gocode.TypeName) string {
	switch t := typeName.(type) {
	case *gocode.BuiltinType:
		{
			return imports.Qualify(t.Package, t.Name)
		}
	case *gocode.UserType:
		{
			return imports.Qualify(t.PackageName, t.Name)
		}
	case *gocode.Pointer:
		{
			return "*" + imports.NameOf(t.PointerTo)
		}
	case *gocode.Slice:
		{
			return "[]" + imports.NameOf(t.SliceOf)
		}
	case *gocode.Map:
		{
			return fmt.Sprintf("map[%s]%s", imports.NameOf(t.KeyType), imports.NameOf(t.ValueType))
		}
	case *gocode.Chan:
		{
			return "chan " + imports.NameOf(t.ChanOf)
		}
	case *gocode.ReceiveChan:
		{
			return "<-chan " + imports.NameOf(t.ReceiveType)
		}
	case *gocode.SendChan:
		{
			return "chan<- " + imports.NameOf(t.SendType)
		}
	default:
		{
			return typeName.String()
		}
	}

}
