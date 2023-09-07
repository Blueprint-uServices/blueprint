package goparser

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"

	"gitlab.mpi-sws.org/cld/blueprint/plugins/golang/gocode"
	"golang.org/x/mod/modfile"
)

/*
Parses go modules and extracts module, package, struct, and interface info
*/

/*
A set of modules on the local filesystem that contain workflow spec interfaces
and implementations.  It is allowed for a workflow spec implementation in one
package to use the interface defined in another package.  However, currently,
it is not possible to use workflow spec nodes whose interface or implementation
comes entirely from an external module (ie. a module that exists only as a
'require' directive of a go.mod)
*/
type ParsedModuleSet struct {
	Modules map[string]*ParsedModule // Map from FQ module name to module object
}

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

type ParsedModule struct {
	ModuleSet *ParsedModuleSet
	Name      string                    // Fully qualified name of the module
	Version   string                    // Version of the module
	SrcDir    string                    // Fully qualified location of the module on the filesystem
	Modfile   *modfile.File             // The modfile File struct is sufficiently simple that we just use it directly
	Packages  map[string]*ParsedPackage // Map from fully qualified package name to ParsedPackage
}

type ParsedPackage struct {
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
}

type ParsedFile struct {
	Package          *ParsedPackage
	Name             string                   // Filename
	Path             string                   // Fully qualified path to the file
	AnonymousImports []*ParsedImport          // Import declarations that were imported with .
	NamedImports     map[string]*ParsedImport // Import declarations - map from shortname to fully qualified package import name
	Ast              *ast.File                // The AST of the file
}

type ParsedStruct struct {
	File            *ParsedFile
	Ast             *ast.StructType
	Name            string
	Methods         map[string]*ParsedFunc  // Methods declared directly on this struct, does not include promoted methods (not implemented yet)
	FieldsList      []*ParsedField          // All fields in the order that they are declared
	Fields          map[string]*ParsedField // Named fields declared in this struct only, does not include promoted fields (not implemented yet)
	PromotedField   *ParsedField            // If there is a promoted field, stored here
	AnonymousFields []*ParsedField          // Subsequent anonymous fields
}

type ParsedInterface struct {
	File    *ParsedFile
	Ast     *ast.InterfaceType
	Name    string
	Methods map[string]*ParsedFunc
}

type ParsedFunc struct {
	gocode.Func
	File *ParsedFile
	Ast  *ast.FuncType
}

type ParsedImport struct {
	gocode.Source
	File *ParsedFile
}

type ParsedField struct {
	gocode.Variable
	Struct   *ParsedStruct
	Position int
	Ast      *ast.Field
}

func ParseWorkspace(workspaceDir string) (*ParsedModuleSet, error) {
	entries, err := os.ReadDir(workspaceDir)
	if err != nil {
		return nil, fmt.Errorf("unable to read workspace directory %v due to %v", workspaceDir, err.Error())
	}

	var moduleDirs []string
	for _, e := range entries {
		if e.IsDir() {
			moduleDirs = append(moduleDirs, filepath.Join(workspaceDir, e.Name()))
		}
	}

	return ParseModules(moduleDirs...)
}

func ParseModules(srcDirs ...string) (*ParsedModuleSet, error) {
	// Create the module set
	set := &ParsedModuleSet{}
	set.Modules = make(map[string]*ParsedModule)

	// Load the modules
	for _, srcDir := range srcDirs {
		mod, err := set.AddModule(srcDir)
		if err != nil {
			return nil, err
		}
		err = mod.Load()
		if err != nil {
			return nil, err
		}
	}

	// Parse the modules
	for _, mod := range set.Modules {
		for _, pkg := range mod.Packages {
			err := pkg.Parse()
			if err != nil {
				return nil, err
			}
		}
	}

	return set, nil
}

func (set *ParsedModuleSet) AddModule(srcDir string) (*ParsedModule, error) {
	srcDir = filepath.Clean(srcDir)
	modfilePath := filepath.Join(srcDir, "go.mod")
	modfileData, err := os.ReadFile(modfilePath)
	if err != nil {
		return nil, fmt.Errorf("unable to read workflow spec modfile %s due to %s", modfilePath, err.Error())
	}

	modf, err := modfile.Parse(modfilePath, modfileData, nil)
	if err != nil {
		return nil, fmt.Errorf("unable to parse %s due to %s", modfilePath, err.Error())
	}

	mod := &ParsedModule{}
	mod.ModuleSet = set
	mod.Modfile = modf
	mod.Name = modf.Module.Mod.Path
	mod.Version = modf.Module.Mod.Version
	mod.SrcDir = srcDir
	mod.Packages = make(map[string]*ParsedPackage)

	if existing, exists := set.Modules[mod.Name]; exists {
		return nil, fmt.Errorf("duplicate definition of module %v at %v and %v", mod.Name, existing.SrcDir, srcDir)
	}
	set.Modules[mod.Name] = mod

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
			return fmt.Errorf("unable to parse package %v due to %s", path, err.Error())
		}

		for name, pkg := range pkgs {
			p := &ParsedPackage{}
			p.Ast = pkg
			p.ShortName = name
			p.Module = mod
			p.SrcDir = path
			p.PackageDir, err = filepath.Rel(mod.SrcDir, path)
			if err != nil {
				return fmt.Errorf("%s should exist within %s but got %s", path, mod.SrcDir, err.Error())
			}
			p.Files = make(map[string]*ParsedFile)
			p.Name = mod.Name + "/" + filepath.ToSlash(p.PackageDir)
			p.DeclaredTypes = make(map[string]gocode.UserType)
			p.Interfaces = make(map[string]*ParsedInterface)
			p.Structs = make(map[string]*ParsedStruct)
			p.Funcs = make(map[string]*ParsedFunc)

			if existing, exists := mod.Packages[p.Name]; exists {
				return fmt.Errorf("duplicate definition of package %v at %v and %v", p.Name, path, existing.SrcDir)
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
	// Create files
	for filename, ast := range pkg.Ast.Files {
		f := &ParsedFile{}
		f.Ast = ast
		f.Name = filename
		f.AnonymousImports = nil
		f.NamedImports = make(map[string]*ParsedImport)
		f.Package = pkg
		f.Path = filepath.Join(pkg.SrcDir)

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
	}

	return nil
}

func (pkg *ParsedPackage) Parse() error {
	fmt.Println("Parsing " + pkg.Name)
	for _, iface := range pkg.Interfaces {
		for _, method := range iface.Methods {
			err := method.Parse()
			if err != nil {
				return err
			}
			fmt.Printf("Parsed %v.%v\n", iface.Name, method.String())
		}
	}
	for _, struc := range pkg.Structs {
		for _, method := range struc.Methods {
			err := method.Parse()
			if err != nil {
				return err
			}
			fmt.Printf("Parsed %v.%v\n", struc.Name, method.String())
		}
	}
	for _, struc := range pkg.Structs {
		for _, field := range struc.FieldsList {
			err := field.Parse()
			if err != nil {
				return err
			}
		}
		fmt.Println(struc.String())
	}
	for _, f := range pkg.Funcs {
		err := f.Parse()
		if err != nil {
			return err
		}
		fmt.Printf("Parsed %v\n", f.String())
	}
	return nil
}

func (f *ParsedField) Parse() error {
	f.Type = f.Struct.File.ResolveType(f.Ast.Type)
	if f.Type == nil {
		return fmt.Errorf("unable to resolve the type of %v field %v", f.Struct.Name, f)
	}
	fmt.Printf("Parsed %v\n", f.String())
	return nil
}

func (f *ParsedFunc) Parse() error {
	if f.Ast.Params != nil {
		for _, p := range f.Ast.Params.List {
			arg := gocode.Variable{}
			if len(p.Names) > 0 {
				arg.Name = p.Names[0].Name
			}
			arg.Type = f.File.ResolveType(p.Type)
			if arg.Type == nil {
				return fmt.Errorf("%v unable to resolve type of argument %v", f.Name, p.Type)
			}
			f.Arguments = append(f.Arguments, arg)
		}
	}
	if f.Ast.Results != nil {
		for _, r := range f.Ast.Results.List {
			ret := gocode.Variable{}
			if len(r.Names) > 0 {
				ret.Name = r.Names[0].Name
			}
			ret.Type = f.File.ResolveType(r.Type)
			if ret.Type == nil {
				return fmt.Errorf("%v unable to resolve type of retval %v", f.Name, r.Type)
			}
			f.Returns = append(f.Returns, ret)
		}
	}
	return nil
}

/*
An ident can be:
  - a basic type, like int64, float32 etc.
  - a type declared locally within the file or package
  - a type imported with an `import . "package"` decl
*/
func (f *ParsedFile) ResolveIdent(name string) gocode.TypeName {
	if gocode.IsBasicType(name) {
		return &gocode.BasicType{Name: name}
	}

	local, isLocalType := f.Package.DeclaredTypes[name]
	if isLocalType {
		return &local
	}

	if len(f.AnonymousImports) == 1 {
		// Assume (possibly erroneously) that this name just comes from the anonymous imports
		return &gocode.UserType{Source: f.AnonymousImports[0].Source, Name: name}
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

	if gocode.IsBuiltinPackage(pkg.PackageName) {
		return &gocode.BuiltinType{Package: pkg.PackageName, Name: name}
	} else {
		return &gocode.UserType{Source: pkg.Source, Name: name}
	}
}

func (f *ParsedFile) ResolveType(expr ast.Expr) gocode.TypeName {
	switch e := expr.(type) {
	case *ast.Ident:
		return f.ResolveIdent(e.Name)
	case *ast.ArrayType:
		return &gocode.Slice{SliceOf: f.ResolveType(e.Elt)}
	case *ast.MapType:
		return &gocode.Map{KeyType: f.ResolveType(e.Key), ValueType: f.ResolveType(e.Value)}
	case *ast.InterfaceType:
		return &gocode.InterfaceType{InterfaceAst: e}
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
		// TODO: here indexexpr should be supported to handle templated types
		return &gocode.Pointer{PointerTo: f.ResolveType(e.X)}
	case *ast.Ellipsis:
		return &gocode.Ellipsis{EllipsisOf: f.ResolveType(e.Elt)}
	case *ast.ChanType:
		switch e.Dir {
		case ast.SEND:
			return &gocode.SendChan{SendType: f.ResolveType(e.Value)}
		case ast.RECV:
			return &gocode.ReceiveChan{ReceiveType: f.ResolveType(e.Value)}
		default:
			return &gocode.Chan{ChanOf: f.ResolveType(e.Value)}
		}
	case *ast.FuncType:
		return &gocode.FuncType{FuncAst: e}
	default:
		fmt.Printf("unknown or invalid expr type %v\n", e)
	}
	return nil
}

func (f *ParsedFile) LoadImports() error {
	for _, imp := range f.Ast.Imports {
		i := &ParsedImport{}
		i.File = f
		i.PackageName = imp.Path.Value[1 : len(imp.Path.Value)-1] // Strip quotation marks
		i.FindSourceModule()                                      // Set the source module

		// Imports can be one of the following:
		// - import "my.package"
		// - import myp "my.package"
		// - import . "my.package"
		var importedAs string
		if imp.Name == nil {
			splits := strings.Split(i.PackageName, "/")
			importedAs = splits[len(splits)-1]
		} else {
			importedAs = imp.Name.Name
		}

		if importedAs == "." {
			f.AnonymousImports = append(f.AnonymousImports, i)
			fmt.Println("Import anonymous package " + i.PackageName)
		} else {
			f.NamedImports[importedAs] = i
			fmt.Println("Import " + importedAs + " \"" + i.PackageName + "\"")
		}

	}
	return nil
}

func (i *ParsedImport) FindSourceModule() error {
	mod := i.File.Package.Module

	if gocode.IsBuiltinPackage(i.PackageName) {
		fmt.Println("builtin import " + i.PackageName)
		return nil
	}

	if strings.HasPrefix(i.PackageName, mod.Name) {
		// Is a module-local import
		i.ModuleName = mod.Name
		i.ModuleVersion = mod.Version
		fmt.Println("module local import " + i.PackageName)
		return nil
	}

	// Try to match it to a modfile requires statement
	for _, req := range mod.Modfile.Require {
		if strings.HasPrefix(i.PackageName, req.Mod.Path) {
			i.ModuleName = req.Mod.Path
			i.ModuleVersion = req.Mod.Version
			fmt.Println("import " + i.PackageName + " from module " + i.ModuleName)
			return nil
		}
	}

	return fmt.Errorf("unable to find the source module of imported package %v", i.PackageName)
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
				return fmt.Errorf("parsing error, expected typespec in decls of %v", f.Name)
			}

			// Save all types that are declared in the file
			u := gocode.UserType{}
			u.ModuleName = f.Package.Module.Name
			u.ModuleVersion = f.Package.Module.Version
			u.PackageName = f.Package.Name
			u.Name = typespec.Name.Name
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
					fmt.Println("Found interface " + u.Name)

					// Can load interface funcs immediately
					for _, methodDecl := range t.Methods.List {
						funcType, isFuncType := methodDecl.Type.(*ast.FuncType)
						if !isFuncType {
							return fmt.Errorf("expected a function declaration in interface " + iface.Name)
						}

						method := &ParsedFunc{}
						method.Ast = funcType
						method.File = f
						method.Name = methodDecl.Names[0].Name
						iface.Methods[method.Name] = method
						fmt.Printf("Found method %v.%v\n", iface.Name, method.Name)
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
					f.Package.Structs[struc.Name] = struc
					fmt.Println("Found struct " + u.Name)

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
			default:
				fmt.Println("Found userType " + u.Name)
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
			fmt.Println("Package func " + fun.Name)
			continue
		}

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
						return fmt.Errorf("unable to parse receiver type of function %v", fun.Name)
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
			return fmt.Errorf("function declared for receiver %v that does not exist in package", receiverName)
		}
		struc.Methods[fun.Name] = fun
		fmt.Printf("Method %v.%v\n", receiverName, fun.Name)
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

func (pkg *ParsedPackage) Source() gocode.Source {
	return gocode.Source{
		ModuleName:    pkg.Module.Name,
		ModuleVersion: pkg.Module.Version,
		PackageName:   pkg.Name,
	}
}

func (iface *ParsedInterface) Type() *gocode.UserType {
	return &gocode.UserType{
		Source: iface.File.Package.Source(),
		Name:   iface.Name,
	}
}

func (struc *ParsedStruct) Type() *gocode.UserType {
	return &gocode.UserType{
		Source: struc.File.Package.Source(),
		Name:   struc.Name,
	}
}

func (f *ParsedFunc) Source() gocode.Source {
	return f.File.Package.Source()
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
	return b.String()
}

func (f *ParsedFile) String() string {
	b := strings.Builder{}
	b.WriteString("File Name=" + f.Name + "\n")
	b.WriteString("Imports=")
	for as, imp := range f.NamedImports {
		b.WriteString("\n")
		b.WriteString(indent("import "+as+" \""+imp.PackageName+"\"", 2))
	}
	for _, imp := range f.AnonymousImports {
		b.WriteString("\n")
		b.WriteString(indent("import . \""+imp.PackageName+"\"", 2))
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

func (f *ParsedField) String() string {
	if f.Name == "" {
		return f.Type.String()
	} else {
		return f.Name + " " + f.Type.String()
	}
}
