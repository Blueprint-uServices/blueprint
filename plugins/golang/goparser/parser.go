// Package goparser parses go modules and extracts module, package, struct, and interface information.
//
// It is used by Blueprint to parse workflow specs.
package goparser

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	"golang.org/x/exp/slog"

	"github.com/blueprint-uservices/blueprint/blueprint/pkg/blueprint"
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/ir"
	"github.com/blueprint-uservices/blueprint/plugins/golang/gocode"
	"golang.org/x/exp/slices"
	"golang.org/x/mod/modfile"
)

type (

	/*
	   Things to bear in mind:
	    * A struct from one module can implement an interface from another
	    * Declared things are local to a package not a file
	    * Yet, imported packages are per-file
	    * To make sure things are resolved correctly across modules, packages, and files, we need to parse things breadth-first (first modules, then packages, etc.)
	    * import . "something" is likely to cause problems.  If only one package is imported like this, we can assume unresolved types come from that module; more than one we error
	    * any interface is valid for a workflow service.  typechecking function arguments is only needed when there is something like serialization; then there is a restriction on arg types
	    * not implemented yet: we don't currently support structs or interfaces that extend other structs/interfaces
	*/

	ParsedModule struct {
		ShortName string                    // Short name of the module
		Name      string                    // Fully qualified name of the module
		Version   string                    // Version of the module
		SrcDir    string                    // Fully qualified location of the module on the filesystem
		IsLocal   bool                      // Is this a local module or from the go cache?
		Modfile   *modfile.File             // The modfile File struct is sufficiently simple that we just use it directly
		Packages  map[string]*ParsedPackage // Map from fully qualified package name to ParsedPackage
	}

	ParsedPackage struct {
		Module        *ParsedModule
		Name          string                      // Fully qualified name of the package including module name
		ShortName     string                      // Shortname of the package (ie, the name used in an import statement)
		PackageDir    string                      // Subdirectory within the module containing the package
		SrcDir        string                      // Fully qualified location of the package on the filesystem
		Files         map[string]*ParsedFile      // Map from filename to ParsedFile
		Ast           *ast.Package                // The AST of the package
		DeclaredTypes map[string]gocode.UserType  // Types declared within this package
		Structs       map[string]*ParsedStruct    // Structs parsed from this package
		Interfaces    map[string]*ParsedInterface // Interfaces parsed from this package
		Funcs         map[string]*ParsedFunc      // Functions parsed from this package (does not include funcs with receiver types)
		Vars          map[string]*ParsedVar       // Vars declared in this package; we save their AST but don't process them
	}

	ParsedFile struct {
		Package          *ParsedPackage
		Name             string                   // Filename
		Path             string                   // Fully qualified path to the file
		PathInModule     string                   // Path within the module to the file
		AnonymousImports []*ParsedImport          // Import declarations that were imported with .
		NamedImports     map[string]*ParsedImport // Import declarations - map from shortname to fully qualified package import name
		Ast              *ast.File                // The AST of the file
	}

	ParsedStruct struct {
		File            *ParsedFile
		Ast             *ast.StructType
		Name            string
		Methods         map[string]*ParsedFunc  // Methods declared directly on this struct, does not include promoted methods (not implemented yet)
		FieldsList      []*ParsedField          // All fields in the order that they are declared
		Fields          map[string]*ParsedField // Named fields declared in this struct only, does not include promoted fields (not implemented yet)
		PromotedField   *ParsedField            // If there is a promoted field, stored here
		AnonymousFields []*ParsedField          // Subsequent anonymous fields
		TypeParams      []string                // Names of generic type parameters
	}

	ParsedInterface struct {
		File    *ParsedFile
		Ast     *ast.InterfaceType
		Name    string
		Methods map[string]*ParsedFunc
	}

	ParsedFunc struct {
		gocode.Func
		File *ParsedFile
		Ast  *ast.FuncType
	}

	// Currently we save var statements but don't do anything with them
	ParsedVar struct {
		File *ParsedFile
		Name string
		Ast  *ast.ValueSpec
	}

	ParsedImport struct {
		File    *ParsedFile
		Package string
	}

	ParsedField struct {
		gocode.Variable
		Struct   *ParsedStruct
		Position int
		Ast      *ast.Field
	}
)

// Parses the module directory specified by srcDir and returns a ParsedModule.
// All of the packages inside the module will be parsed.
// srcDir must contain a go.mod file; if it doesn't an error will be returned.
func parseModule(srcDir string) (*ParsedModule, error) {
	srcDir = filepath.Clean(srcDir)

	modfilePath := filepath.Join(srcDir, "go.mod")
	modfileData, err := os.ReadFile(modfilePath)
	if err != nil {
		return nil, blueprint.Errorf("unable to read workflow spec modfile %s due to %s", modfilePath, err.Error())
	}

	modf, err := modfile.Parse(modfilePath, modfileData, nil)
	if err != nil {
		return nil, blueprint.Errorf("unable to parse %s due to %s", modfilePath, err.Error())
	}

	nameSplits := strings.Split(modf.Module.Mod.Path, "/")

	mod := &ParsedModule{}
	mod.Modfile = modf
	mod.Name = modf.Module.Mod.Path
	mod.ShortName = nameSplits[len(nameSplits)-1]
	mod.Version = modf.Module.Mod.Version
	mod.SrcDir = srcDir
	mod.IsLocal = true
	mod.Packages = make(map[string]*ParsedPackage)

	// Try to update the module version if it's on our go path; doesn't work for generated modules.
	info, err := GetModuleInfo(mod.Name)
	if err == nil && info.Dir == srcDir {
		mod.Version = info.Version
	}

	if err := mod.Load(); err != nil {
		return mod, err
	}

	for _, pkg := range mod.Packages {
		if err := pkg.Parse(); err != nil {
			return mod, err
		}
	}
	slog.Info(fmt.Sprintf("Parsed %s version=%s local=%v", mod.Name, mod.Version, mod.IsLocal))

	return mod, nil
}

func (mod *ParsedModule) Load() error {
	// Find all packages within the module, parse them, save but don't process the AST
	var fset token.FileSet
	err := filepath.Walk(mod.SrcDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			return nil
		}

		pkgs, err := parser.ParseDir(&fset, path, nil, parser.ParseComments)
		if err != nil {
			return blueprint.Errorf("unable to parse package %v due to %s", path, err.Error())
		}

		for name, pkg := range pkgs {
			// Ignore _test packages
			if strings.HasSuffix(name, "_test") {
				continue
			}

			p := &ParsedPackage{}
			p.Ast = pkg
			p.ShortName = name
			p.Module = mod
			p.SrcDir = path
			p.PackageDir, err = filepath.Rel(mod.SrcDir, path)
			if err != nil {
				return blueprint.Errorf("%s should exist within %s but got %s", path, mod.SrcDir, err.Error())
			}
			p.Files = make(map[string]*ParsedFile)
			if p.PackageDir == "." {
				p.Name = mod.Name
			} else {
				p.Name = mod.Name + "/" + filepath.ToSlash(p.PackageDir)
			}
			p.DeclaredTypes = make(map[string]gocode.UserType)
			p.Interfaces = make(map[string]*ParsedInterface)
			p.Structs = make(map[string]*ParsedStruct)
			p.Funcs = make(map[string]*ParsedFunc)
			p.Vars = make(map[string]*ParsedVar)

			if existing, exists := mod.Packages[p.Name]; exists {
				return blueprint.Errorf("duplicate definition of package %v at %v and %v", p.Name, path, existing.SrcDir)
			}
			mod.Packages[p.Name] = p

			p.Load()
		}
		return nil
	})
	if err != nil {
		return err
	}

	return nil
}

func (pkg *ParsedPackage) Load() error {
	moduleDir := pkg.Module.SrcDir

	// Create files
	for filename, ast := range pkg.Ast.Files {
		f := &ParsedFile{}
		f.Ast = ast
		f.Name = filename
		f.AnonymousImports = nil
		f.NamedImports = make(map[string]*ParsedImport)
		f.Package = pkg
		f.Path = filepath.Join(pkg.SrcDir)
		var err error
		f.PathInModule, err = filepath.Rel(moduleDir, f.Path)
		if err != nil {
			return err
		}

		pkg.Files[filename] = f
	}

	// Load imports
	for _, f := range pkg.Files {
		err := f.LoadImports()
		if err != nil {
			return err
		}
	}

	// Load all of the declared structs and interfaces
	for _, f := range pkg.Files {
		err := f.LoadStructsAndInterfaces()
		if err != nil {
			return err
		}
	}

	// Load all of the functions.  Must be done
	// after loading structs and interfaces since function
	// definitions can be in different files to receiver types
	for _, f := range pkg.Files {
		err := f.LoadFuncs()
		if err != nil {
			return err
		}
		err = f.LoadVars()
		if err != nil {
			return err
		}
	}

	return nil
}

func (pkg *ParsedPackage) Parse() error {
	for _, iface := range pkg.Interfaces {
		for _, method := range iface.Methods {
			err := method.Parse()
			if err != nil {
				return err
			}
		}
	}
	for _, struc := range pkg.Structs {
		for _, method := range struc.Methods {
			err := method.Parse()
			if err != nil {
				return err
			}
		}
	}
	for _, struc := range pkg.Structs {
		for _, field := range struc.FieldsList {
			err := field.Parse()
			if err != nil {
				return err
			}
		}
	}
	for _, f := range pkg.Funcs {
		err := f.Parse()
		if err != nil {
			return err
		}
	}
	return nil
}

func (f *ParsedField) Parse() error {
	f.Type = f.Struct.File.ResolveType(f.Ast.Type, f.Struct.TypeParams...)
	if f.Type == nil {
		return blueprint.Errorf("unable to resolve the type of %v field %v", f.Struct.Name, f)
	}
	return nil
}

func (f *ParsedFunc) Parse() error {
	if f.Ast.Params != nil {
		for _, p := range f.Ast.Params.List {
			// Determine the argument's type
			argType := f.File.ResolveType(p.Type)
			if argType == nil {
				return blueprint.Errorf("%v unable to resolve type of argument %v", f.Name, p.Type)
			}

			if len(p.Names) == 0 {
				// This is an unnamed argument such as on an interface
				f.Arguments = append(f.Arguments, gocode.Variable{Type: argType})
			}

			for _, argName := range p.Names {
				// Can be 1 or more arguments sharing a type
				f.Arguments = append(f.Arguments, gocode.Variable{Name: argName.Name, Type: argType})
			}
		}
	}
	if f.Ast.Results != nil {
		for _, r := range f.Ast.Results.List {
			// Determine the retval's type
			retType := f.File.ResolveType(r.Type)
			if retType == nil {
				return blueprint.Errorf("%v unable to resolve type of retval %v", f.Name, r.Type)
			}

			if len(r.Names) == 0 {
				// An unnamed return value
				f.Returns = append(f.Returns, gocode.Variable{Type: retType})
			}

			for _, retName := range r.Names {
				// Can be 1 or more return values sharing a type
				f.Returns = append(f.Returns, gocode.Variable{Name: retName.Name, Type: retType})
			}
		}
	}
	return nil
}

/*
An ident can be:
  - a basic type, like int64, float32 etc.
  - any
  - a type declared locally within the file or package
  - a type imported with an `import . "package"` decl
  - a generic type from a struct or func's type params
*/
func (f *ParsedFile) ResolveIdent(name string, typeParams ...string) gocode.TypeName {
	if gocode.IsBasicType(name) {
		return &gocode.BasicType{Name: name}
	}

	if name == "any" {
		return &gocode.AnyType{}
	}

	local, isLocalType := f.Package.DeclaredTypes[name]
	if isLocalType {
		return &local
	}

	if slices.Contains(typeParams, name) {
		return &gocode.GenericTypeParam{ParamName: name}
	}

	if len(f.AnonymousImports) == 1 {
		// Assume (possibly erroneously) that this name just comes from the anonymous imports
		return &gocode.UserType{Package: f.AnonymousImports[0].Package, Name: name}
	}

	fmt.Printf("Unable to resolve ident %v in file %v\n", name, f.Name)

	return nil
}

func (f *ParsedFile) ResolveSelector(packageShortName string, name string) gocode.TypeName {
	pkg, isImported := f.NamedImports[packageShortName]
	if !isImported {
		fmt.Printf("Unable to resolve type %v.%v in file %v\n", packageShortName, name, f.Name)
		return nil
	}

	return &gocode.UserType{Package: pkg.Package, Name: name}
}

// If the expr is in the context of a generic struct or func,
// typeParams provides the additional named type params
func (f *ParsedFile) ResolveType(expr ast.Expr, typeParams ...string) gocode.TypeName {
	switch e := expr.(type) {
	case *ast.Ident:
		return f.ResolveIdent(e.Name, typeParams...)
	case *ast.ArrayType:
		return &gocode.Slice{SliceOf: f.ResolveType(e.Elt, typeParams...)}
	case *ast.MapType:
		return &gocode.Map{KeyType: f.ResolveType(e.Key, typeParams...), ValueType: f.ResolveType(e.Value, typeParams...)}
	case *ast.InterfaceType:
		return &gocode.InterfaceType{}
	case *ast.SelectorExpr:
		{
			x, isIdent := e.X.(*ast.Ident)
			if !isIdent {
				fmt.Printf("encountered invalid selector %v\n", expr)
				return nil
			}
			return f.ResolveSelector(x.Name, e.Sel.Name)
		}
	case *ast.StarExpr:
		return &gocode.Pointer{PointerTo: f.ResolveType(e.X, typeParams...)}
	case *ast.Ellipsis:
		return &gocode.Ellipsis{EllipsisOf: f.ResolveType(e.Elt, typeParams...)}
	case *ast.ChanType:
		switch e.Dir {
		case ast.SEND:
			return &gocode.SendChan{SendType: f.ResolveType(e.Value, typeParams...)}
		case ast.RECV:
			return &gocode.ReceiveChan{ReceiveType: f.ResolveType(e.Value, typeParams...)}
		default:
			return &gocode.Chan{ChanOf: f.ResolveType(e.Value, typeParams...)}
		}
	case *ast.FuncType:
		return &gocode.FuncType{}
	case *ast.StructType:
		return &gocode.StructType{}
	case *ast.IndexExpr:
		return &gocode.GenericType{BaseType: f.ResolveType(e.X, typeParams...)}
	case *ast.IndexListExpr:
		return &gocode.GenericType{BaseType: f.ResolveType(e.X, typeParams...)}
	default:
		fmt.Printf("unknown or invalid expr type %v %v\n", reflect.TypeOf(expr), expr)
	}
	return nil
}

func (f *ParsedFile) LoadImports() error {
	for _, imp := range f.Ast.Imports {
		i := &ParsedImport{}
		i.File = f
		i.Package = imp.Path.Value[1 : len(imp.Path.Value)-1] // Strip quotation marks

		// Imports can be one of the following:
		// - import "my.package"
		// - import myp "my.package"
		// - import . "my.package"
		var importedAs string
		if imp.Name == nil {
			splits := strings.Split(i.Package, "/")
			importedAs = splits[len(splits)-1]
		} else {
			importedAs = imp.Name.Name
		}

		if importedAs == "." {
			f.AnonymousImports = append(f.AnonymousImports, i)
		} else {
			f.NamedImports[importedAs] = i
		}

	}
	return nil
}

/*
Looks for:
  - structs defined in the file
  - interfaces defined in the file
  - other user types defined in the file

Does not:
  - look for function declarations
*/
func (f *ParsedFile) LoadStructsAndInterfaces() error {
	for _, decl := range f.Ast.Decls {
		// We are only looking for TYPE declarations. We don't do anything with IMPORT, CONST, or VAR declarations
		d, is_gendecl := decl.(*ast.GenDecl)
		if !is_gendecl || d.Tok != token.TYPE {
			continue
		}

		// Process the type declaration block
		for _, spec := range d.Specs {
			typespec, ok := spec.(*ast.TypeSpec)
			if !ok {
				return blueprint.Errorf("parsing error, expected typespec in decls of %v", f.Name)
			}

			var typeParams []string
			if typespec.TypeParams != nil {
				for _, field := range typespec.TypeParams.List {
					if field.Names != nil {
						for _, name := range field.Names {
							typeParams = append(typeParams, name.Name)
						}
					}
				}
			}

			// Save all types that are declared in the file
			u := gocode.UserType{Package: f.Package.Name, Name: typespec.Name.Name}
			f.Package.DeclaredTypes[u.Name] = u

			// Also specifically save interface and struct AST info which we later want to parse.
			// This ignores enums
			switch t := typespec.Type.(type) {
			case *ast.InterfaceType:
				{
					iface := &ParsedInterface{}
					iface.Ast = t
					iface.File = f
					iface.Name = typespec.Name.Name
					iface.Methods = make(map[string]*ParsedFunc)
					f.Package.Interfaces[iface.Name] = iface

					// Can load interface funcs immediately
					for _, methodDecl := range t.Methods.List {
						funcType, isFuncType := methodDecl.Type.(*ast.FuncType)
						if !isFuncType {
							return blueprint.Errorf("expected a function declaration in interface " + iface.Name)
						}

						method := &ParsedFunc{}
						method.Ast = funcType
						method.File = f
						method.Name = methodDecl.Names[0].Name
						iface.Methods[method.Name] = method
					}
				}
			case *ast.StructType:
				{
					struc := &ParsedStruct{}
					struc.Ast = t
					struc.File = f
					struc.Name = typespec.Name.Name
					struc.Methods = make(map[string]*ParsedFunc)
					struc.FieldsList = nil
					struc.Fields = make(map[string]*ParsedField)
					struc.PromotedField = nil
					struc.AnonymousFields = nil
					struc.TypeParams = typeParams

					f.Package.Structs[struc.Name] = struc

					if t.Fields != nil {
						for i, fieldDecl := range t.Fields.List {
							field := &ParsedField{}
							if fieldDecl.Names != nil {
								field.Name = fieldDecl.Names[0].Name
								struc.Fields[field.Name] = field
							} else if struc.PromotedField == nil {
								struc.PromotedField = field
							} else {
								struc.AnonymousFields = append(struc.AnonymousFields, field)
							}
							field.Position = i
							field.Struct = struc
							field.Ast = fieldDecl
							struc.FieldsList = append(struc.FieldsList, field)
						}
					}
				}
			}
		}
	}
	return nil
}

/*
Looks for:
  - vars declared
*/
func (f *ParsedFile) LoadVars() error {
	for _, decl := range f.Ast.Decls {
		// We are only looking for VAR declarations.
		d, is_gendecl := decl.(*ast.GenDecl)
		if !is_gendecl || d.Tok != token.VAR {
			continue
		}

		for _, spec := range d.Specs {
			valspec, ok := spec.(*ast.ValueSpec)
			if !ok {
				return blueprint.Errorf("parsing error, expected valuespec in decls of %v", f.Name)
			}

			// Save the AST of the var for later use
			f.Package.Vars[valspec.Names[0].Name] = &ParsedVar{
				File: f,
				Name: valspec.Names[0].Name,
				Ast:  valspec,
			}
		}
	}
	return nil
}

/*
Assumes that all structs and interfaces have been loaded for the package containing the file.

Loads the names of all funcs.  If the func has a receiver type, then it is saved as a method on the
appropriate struct; if it does not have a receiver type, then it is saved as a package func.

This does not parse the arguments or returns of the func
*/
func (f *ParsedFile) LoadFuncs() error {
	for _, decl := range f.Ast.Decls {
		// We are only looking for FuncDecls
		d, is_funcdecl := decl.(*ast.FuncDecl)
		if !is_funcdecl {
			continue
		}

		fun := &ParsedFunc{}
		fun.Ast = d.Type
		fun.File = f
		fun.Name = d.Name.Name

		if d.Recv == nil {
			// This function is not associated with a struct, but it might still be a constructor
			// We will associate constructors to structs later
			f.Package.Funcs[fun.Name] = fun
			continue
		}

		// This doesn't work with generic types

		// Pull out the name of the receiver struct
		var receiverName string
		switch receiverType := d.Recv.List[0].Type.(type) {
		case *ast.StarExpr:
			{
				switch pointerReceiverType := receiverType.X.(type) {
				case *ast.Ident:
					{
						receiverName = pointerReceiverType.Name
					}
				default:
					{
						return blueprint.Errorf("unable to parse receiver type of function %v", fun.Name)
					}
				}
			}
		case *ast.Ident:
			{
				// Declared as func(receiver MyType) funcName(...) {}
				receiverName = receiverType.Name
			}
		}

		// Associate the func with the receiver struct
		struc, exists := f.Package.Structs[receiverName]
		if !exists {
			return blueprint.Errorf("function declared for receiver %v that does not exist in package", receiverName)
		}
		struc.Methods[fun.Name] = fun
	}

	return nil
}

func indent(str string, amount int) string {
	lines := strings.Split(str, "\n")
	for i, line := range lines {
		lines[i] = strings.Repeat(" ", amount) + line
	}
	return strings.Join(lines, "\n")
}

func (iface *ParsedInterface) Type() *gocode.UserType {
	return &gocode.UserType{
		Name:    iface.Name,
		Package: iface.File.Package.Name,
	}
}

func (struc *ParsedStruct) Type() *gocode.UserType {
	return &gocode.UserType{
		Name:    struc.Name,
		Package: struc.File.Package.Name,
	}
}

func (iface *ParsedInterface) ServiceInterface(ctx ir.BuildContext) *gocode.ServiceInterface {
	methods := make(map[string]gocode.Func)
	for name, method := range iface.Methods {
		methods[name] = gocode.Func{
			Name:      method.Name,
			Arguments: method.Arguments[1:],
			Returns:   method.Returns[:len(method.Returns)-1],
		}
	}
	return &gocode.ServiceInterface{
		UserType: *iface.Type(),
		BaseName: (*iface.Type()).Name,
		Methods:  methods,
	}
}

func (f *ParsedFunc) AsConstructor() *gocode.Constructor {
	return &gocode.Constructor{
		Func:    f.Func,
		Package: f.File.Package.Name,
	}
}

func (set *ParsedModuleSet) String() string {
	var modStrings []string
	for _, mod := range set.Modules {
		modStrings = append(modStrings, mod.String())
	}
	return strings.Join(modStrings, "\n")
}

func (mod *ParsedModule) String() string {
	b := strings.Builder{}
	b.WriteString("Module name=" + mod.Name + "\n")
	b.WriteString("Module srcDir=" + mod.SrcDir + "\n")
	b.WriteString("Module packages=")
	for _, pkg := range mod.Packages {
		b.WriteString("\n")
		b.WriteString(indent(pkg.String(), 2))
	}
	return b.String()
}

func (pkg *ParsedPackage) String() string {
	b := strings.Builder{}
	b.WriteString("Package Name=" + pkg.Name + "\n")
	b.WriteString("Package ShortName=" + pkg.ShortName + "\n")
	b.WriteString("Package PackageDir=" + pkg.PackageDir + "\n")
	b.WriteString("Package SrcDir=" + pkg.SrcDir + "\n")
	b.WriteString("Package Files=")
	for _, f := range pkg.Files {
		b.WriteString("\n")
		b.WriteString(indent(f.String(), 2))
	}
	b.WriteString("\nPackage Interfaces=")
	for _, iface := range pkg.Interfaces {
		b.WriteString("\n")
		b.WriteString(indent(iface.Name, 2))
	}
	b.WriteString("\nPackage Structs=")
	for _, struc := range pkg.Structs {
		b.WriteString("\n")
		b.WriteString(indent(struc.Name, 2))
	}
	return b.String()
}

func (f *ParsedFile) String() string {
	b := strings.Builder{}
	b.WriteString("File Name=" + f.Name + "\n")
	b.WriteString("Imports=")
	for as, imp := range f.NamedImports {
		b.WriteString("\n")
		b.WriteString(indent("import "+as+" \""+imp.Package+"\"", 2))
	}
	for _, imp := range f.AnonymousImports {
		b.WriteString("\n")
		b.WriteString(indent("import . \""+imp.Package+"\"", 2))
	}
	return b.String()
}

func (f *ParsedStruct) String() string {
	b := strings.Builder{}
	b.WriteString("type " + f.Name + " struct {\n")
	for _, field := range f.FieldsList {
		b.WriteString("  " + field.String() + "\n")
	}
	b.WriteString("}")
	return b.String()
}

func (i *ParsedInterface) String() string {
	b := strings.Builder{}
	b.WriteString("type " + i.Name + " interface {\n")
	for _, method := range i.Methods {
		b.WriteString("  " + method.String() + "\n")
	}
	b.WriteString("}")
	return b.String()
}

func (f *ParsedFunc) String() string {
	b := strings.Builder{}
	b.WriteString(f.Name + "(")
	var args []string
	for _, arg := range f.Arguments {
		args = append(args, arg.String())
	}
	b.WriteString(strings.Join(args, ", "))
	b.WriteString(")")
	var rets []string
	for _, arg := range f.Returns {
		rets = append(rets, arg.String())
	}
	b.WriteString("(")
	b.WriteString(strings.Join(rets, ", "))
	b.WriteString(")")
	return b.String()
}

func (f *ParsedField) String() string {
	if f.Name == "" {
		return f.Type.String()
	} else {
		return f.Name + " " + f.Type.String()
	}
}
