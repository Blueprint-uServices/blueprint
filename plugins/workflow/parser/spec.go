package parser

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"log"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"unicode"

	"golang.org/x/exp/slog"
	"golang.org/x/mod/modfile"
)

// Data structs
type ArgInfo struct {
	Name string
	Type TypeInfo
}

func (a ArgInfo) String() string {
	str := ""
	if a.Name != "" {
		str += a.Name + " "
	}
	return str + a.Type.String()
}

func GetMapArg(argName string, keyName string, valName string) ArgInfo {
	return ArgInfo{Name: argName, Type: mapToType(keyName, valName)}
}

func GetContextArg(argName string) ArgInfo {
	return ArgInfo{Name: argName, Type: ctxType()}
}

func GetPointerArg(argName string, ptr_type string) ArgInfo {
	return ArgInfo{Name: argName, Type: pointerToType(ptr_type)}
}

func GetErrorArg(argName string) ArgInfo {
	return ArgInfo{Name: argName, Type: errType()}
}

func GetListArg(argName string, argType string) ArgInfo {
	return ArgInfo{Name: argName, Type: arrayToType(argType)}
}

func GetBasicArg(argName string, argType string) ArgInfo {
	return ArgInfo{Name: argName, Type: stringToType(argType)}
}

func GetVariadicArg(argName string, argType string) ArgInfo {
	return ArgInfo{Name: argName, Type: ellipsisToType(argType)}
}

type FuncInfo struct {
	Name    string // Name of the func
	Args    []ArgInfo
	Return  []ArgInfo
	Imports *ImportedPackages // Package
	Package *PackageInfo      // The package containing the func definition
	Public  bool
}

func (f FuncInfo) GetArgNames() []string {
	var retvals []string
	for _, arg := range f.Args {
		retvals = append(retvals, arg.Name)
	}
	return retvals
}

type ServiceInfo struct {
	Name    string
	Methods map[string]FuncInfo
	Package *PackageInfo
	Imports *ImportedPackages
}

type ImportedPackage struct {
	Name          string // The optional alias for the imported package
	ImportName    string // The fully qualified import package name
	Module        string // The module that the package exists within
	ModuleVersion string // The module version that the package exists within
	IsStandard    bool   // Is this a stdlib package
}

/*
Used when parsing files to figure out what the module and package dependencies are
of a source file
*/
type ImportedPackages struct {
	Imports map[string]ImportedPackage // Map from the package alias to ImportedPackage struct
}

type EnumInfo struct {
	Name     string
	Type     string // Has to be a Basic Type!
	PkgPath  string
	ValNames []string
}

type ImplInfo struct {
	Name             string
	Fields           []ArgInfo
	Methods          map[string]FuncInfo
	Interfaces       map[string]bool
	PkgPath          string
	Imports          *ImportedPackages
	Package          *PackageInfo
	ConstructorInfos []FuncInfo
}

func (d ImplInfo) String() string {
	var b strings.Builder
	b.WriteString(d.Name)
	b.WriteString("(")
	b.WriteString(strings.Join(d.ConstructorInfos[0].GetArgNames(), ", "))
	b.WriteString(")")
	return b.String()
}

type ModuleInfo struct {
	Name     string            // The fully qualified name of the module
	Version  string            // The version of the module
	Path     string            // For a local module, the path on the local filesystem to the module; otherwise ""
	Requires map[string]string // For a local module, other modules required by this module; otherwise empty
}

type PackageInfo struct {
	ShortName  string // The name of the package in the `package xxx` declaration of a file
	Name       string // The fully qualified name within the module, e.g. if it's in subdirs, it might be e.g. pkg/wiring/xxx
	Path       string // The path on the local filesystem to the package
	ImportName string // The fully qualified import used elsewhere, e.g. in the `import github.com/my/system/pkg/wiring/xxx`
	Module     *ModuleInfo
	Package    *ast.Package
}

func (pkg PackageInfo) String() string {
	b := strings.Builder{}
	b.WriteString(pkg.Module.String())
	b.WriteString("\nPackage " + pkg.ShortName + " at " + pkg.Name)
	b.WriteString("\nImport " + pkg.ImportName)
	return b.String()
}

type SpecParser struct {
	srcDirs         []string
	logger          *log.Logger
	Services        map[string]*ServiceInfo
	Implementations map[string]*ImplInfo
	Functions       map[string][]*FuncInfo
	ExtraFunctions  []*FuncInfo
	RemoteTypes     map[string]*ImplInfo
	Enums           map[string]*EnumInfo
	Modules         map[string]*ModuleInfo
	Packages        map[string]*PackageInfo
}

// Printers
func (s *SpecParser) PrintServices() {
	for name, service := range s.Services {
		s.logger.Println(name, "has", len(service.Methods), "function(s):", service.Methods)
	}
}

func (s *SpecParser) PrintImplementations() {
	for name, impl := range s.Implementations {
		s.logger.Println(name, "has", len(impl.Fields), "field(s):", impl.Fields)
		s.logger.Println(name, "has", len(impl.Methods), "function(s):", impl.Methods)
		s.logger.Println(name, "implements", len(impl.Interfaces), "interface(s):", impl.Interfaces)
	}
}

func (s *SpecParser) PrintEnums() {
	for name, enum := range s.Enums {
		s.logger.Println("Enum", name, "has", len(enum.ValNames), "value(s):", enum.ValNames)
	}
}

func (s *SpecParser) PrintFunctions() {
	for k, v := range s.Functions {
		s.logger.Println(k, "has", len(v), "functions:", v)
	}
}

// Helper parser functions
func (s *SpecParser) getType(node *ast.Field) TypeInfo {
	switch eType := node.Type.(type) {
	case *ast.Ident:
		return stringToType(eType.Name)
	case *ast.ArrayType:
		switch eltType := eType.Elt.(type) {
		case *ast.Ident:
			return arrayToType(eltType.Name)
		case *ast.InterfaceType:
			return arrayToType("interface")
		case *ast.SelectorExpr:
			var selXName string
			var selName string
			switch selXType := eltType.X.(type) {
			case *ast.Ident:
				selXName = selXType.Name
			default:
				s.logger.Fatal(reflect.TypeOf(selXType), "is not a valid option for a selector")
			}
			selName = eltType.Sel.Name
			return arrayToType(selXName + "." + selName)
		default:
			s.logger.Fatal(reflect.TypeOf(eltType), " is not a valid Element Type for List")
		}
	case *ast.MapType:
		var keyName string
		var valName string
		switch keyType := eType.Key.(type) {
		case *ast.Ident:
			keyName = keyType.Name
		default:
			s.logger.Fatal(reflect.TypeOf(keyType), " is not a valid Key Type for Map")
		}
		switch valType := eType.Value.(type) {
		case *ast.Ident:
			valName = valType.Name
		case *ast.InterfaceType:
			valName = "interface"
		default:
			s.logger.Fatal(reflect.TypeOf(valType), " is not a valid Val Type for Map")
		}
		return mapToType(keyName, valName)
	case *ast.InterfaceType:
		return interfaceToType()
	case *ast.SelectorExpr:
		var selXName string
		var selName string
		switch selXType := eType.X.(type) {
		case *ast.Ident:
			selXName = selXType.Name
		default:
			s.logger.Fatal(reflect.TypeOf(selXType), "is not a valid option for a selector")
		}
		selName = eType.Sel.Name
		if selName == "Context" && selXName == "context" {
			return ctxType()
		}
		return stringToType(selXName + "." + selName)
	case *ast.StarExpr:
		switch xType := eType.X.(type) {
		case *ast.Ident:
			valName := xType.Name
			return pointerToType(valName)
		case *ast.SelectorExpr:
			var selXName string
			var selName string
			switch selXType := xType.X.(type) {
			case *ast.Ident:
				selXName = selXType.Name
			default:
				s.logger.Fatal(reflect.TypeOf(selXType), "is not a valid option for a selector")
			}
			selName = xType.Sel.Name
			return pointerToType(selXName + "." + selName)
		case *ast.IndexExpr:
			// We can have templated types now
			var exprName string
			var indexName string
			switch exprType := xType.X.(type) {
			case *ast.Ident:
				exprName = exprType.Name
			case *ast.SelectorExpr:
				var selXName string
				switch selXType := exprType.X.(type) {
				case *ast.Ident:
					selXName = selXType.Name
				default:
					s.logger.Fatal(reflect.TypeOf(selXType), " is not a valid option for an index expression")
				}
				exprName = selXName + "." + exprType.Sel.Name
			default:
				s.logger.Fatal(reflect.TypeOf(exprType), "is not a valid option for an index expression")
			}
			switch indexType := xType.Index.(type) {
			case *ast.Ident:
				indexName = indexType.Name
			case *ast.StarExpr:
				switch indexStarxtype := indexType.X.(type) {
				case *ast.Ident:
					indexName = "*" + indexStarxtype.Name
				default:
					s.logger.Fatal(reflect.TypeOf(indexStarxtype), " is not a valid option for an index expression")
				}
			default:
				s.logger.Fatal(reflect.TypeOf(indexType), "is not a valid option for an index expression")
			}
			return pointerToType(exprName + "[" + indexName + "]")
		default:
			s.logger.Fatal(reflect.TypeOf(xType), " is not a valid Val Type for a Pointer")
		}
	case *ast.Ellipsis:
		switch eltType := eType.Elt.(type) {
		case *ast.Ident:
			return ellipsisToType(eltType.Name)
		case *ast.InterfaceType:
			return ellipsisToType("interface")
		default:
			s.logger.Fatal(reflect.TypeOf(eltType), " is not a valid Element Type for Ellipsis")
		}
	case *ast.ChanType:
		switch valType := eType.Value.(type) {
		case *ast.Ident:
			valName := valType.Name
			return chanToType(valName)
		case *ast.SelectorExpr:
			var selXName string
			var selName string
			switch selXType := valType.X.(type) {
			case *ast.Ident:
				selXName = selXType.Name
			default:
				s.logger.Fatal(reflect.TypeOf(selXType), "is not a valid option for a selector")
			}
			selName = valType.Sel.Name
			return chanToType(selXName + "." + selName)
		default:
			s.logger.Fatal(reflect.TypeOf(valType), "is not a valid Val Type for a channel")
		}
	case *ast.FuncType:
		return funcType()
	default:
		s.logger.Fatal(reflect.TypeOf(eType), " is not currently supported by the Parser")
	}
	return TypeInfo{}
}

func (s *SpecParser) getFuncInfo(node *ast.FuncType, pkgInfo *PackageInfo, imports *ImportedPackages) FuncInfo {
	var args []ArgInfo
	var returns []ArgInfo
	if node.Params != nil {
		// Params are not nil which means there are parameters
		for _, param := range node.Params.List {
			paramName := param.Names[0].Name
			paramType := s.getType(param)
			arg := ArgInfo{Name: paramName, Type: paramType}
			args = append(args, arg)
		}
	}
	if node.Results != nil {
		// Results are not nil which means there are return vals
		for _, retParam := range node.Results.List {
			var resultName string
			if retParam.Names != nil {
				resultName = retParam.Names[0].Name
			}
			resultType := s.getType(retParam)
			result := ArgInfo{Name: resultName, Type: resultType}
			returns = append(returns, result)
		}
	}

	funcInfo := FuncInfo{}
	funcInfo.Args = args
	funcInfo.Return = returns
	funcInfo.Package = pkgInfo
	funcInfo.Imports = imports

	return funcInfo
}

// Parser Functions
func (s *SpecParser) parseTypes(path string, decl *ast.GenDecl, pkgInfo *PackageInfo, imports *ImportedPackages) {
	// TODO: fully qualify any imported user types
	for _, spec := range decl.Specs {
		typespec, ok := spec.(*ast.TypeSpec)
		if !ok {
			s.logger.Fatal("Parsing Error: Parser expected a type specification")
		}
		name := typespec.Name.Name
		switch t := typespec.Type.(type) {
		case *ast.InterfaceType:
			methods := make(map[string]FuncInfo)
			for _, method := range t.Methods.List {
				var funcInfo FuncInfo
				switch methodType := method.Type.(type) {
				case *ast.FuncType:
					funcName := method.Names[0].Name
					funcInfo = s.getFuncInfo(methodType, pkgInfo, imports)
					funcInfo.Name = funcName
					funcInfo.Public = true
					methods[funcName] = funcInfo
				default:
					s.logger.Fatal("Parsing Error: Expected a function declaration in Interface Type and not", reflect.TypeOf(methodType))
				}
			}
			serviceInfo := ServiceInfo{Name: name, Methods: methods, Package: pkgInfo, Imports: imports}
			s.Services[name] = &serviceInfo

		case *ast.StructType:
			var fields []ArgInfo
			if t.Fields.List != nil {
				for _, field := range t.Fields.List {
					var fieldName string
					if field.Names != nil {
						fieldName = field.Names[0].Name
					}
					typeString := s.getType(field)
					fieldInfo := ArgInfo{Name: fieldName, Type: typeString}
					fields = append(fields, fieldInfo)
				}
			}
			implInfo := ImplInfo{Name: name, Fields: fields, PkgPath: path, Imports: imports, Package: pkgInfo}
			s.Implementations[name] = &implInfo

		case *ast.Ident:
			// Potentially could be an enum!!!! (Otherwise it is just a simple typedef)
			enumInfo := EnumInfo{Name: name, Type: t.Name, PkgPath: path}
			s.Enums[name] = &enumInfo
		}
	}
}

func (s *SpecParser) parseFunctions(decl *ast.FuncDecl, pkgInfo *PackageInfo, imports *ImportedPackages) {
	if decl.Recv == nil {
		// If the receiver is Nil then EITHER this function is not associated with a struct OR this function is a constructor for a struct
		funcInfo := s.getFuncInfo(decl.Type, pkgInfo, imports)
		funcInfo.Name = decl.Name.Name
		s.ExtraFunctions = append(s.ExtraFunctions, &funcInfo)
		return
	}
	var recvName string
	switch recvType := decl.Recv.List[0].Type.(type) {
	case *ast.StarExpr:
		switch starType := recvType.X.(type) {
		case *ast.Ident:
			recvName = starType.Name
		}
	case *ast.Ident:
		recvName = recvType.Name
	default:
		s.logger.Fatal("The receiver for a function should either be a star expression or an identifier")
	}
	name := decl.Name.Name
	funcInfo := s.getFuncInfo(decl.Type, pkgInfo, imports)
	funcInfo.Name = name
	runes := []rune(name)
	funcInfo.Public = unicode.IsUpper(runes[0])
	if v, ok := s.Functions[recvName]; ok {
		s.Functions[recvName] = append(v, &funcInfo)
	} else {
		s.Functions[recvName] = []*FuncInfo{&funcInfo}
	}
}

func (s *SpecParser) associateFunctions() {
	for implName, functions := range s.Functions {
		if serviceInfo, ok := s.Implementations[implName]; ok {
			impl_functions := make(map[string]FuncInfo)
			for _, function := range functions {
				impl_functions[function.Name] = *function
			}
			serviceInfo.Methods = impl_functions
		} else {
			s.logger.Println("Unable to find the receiver (possibly an enum type)", implName)
		}
	}

	// Find constructors
	for _, function := range s.ExtraFunctions {
		for _, retArg := range function.Return {
			retArgName := retArg.Type.String()
			retArgName = strings.ReplaceAll(retArgName, "*", "")
			if sinfo, ok := s.Implementations[retArgName]; ok {
				sinfo.ConstructorInfos = append(sinfo.ConstructorInfos, *function)
			}
		}
	}
}

func (s *SpecParser) parseConstBlock(path string, t *ast.GenDecl, pkgInfo *PackageInfo, imports *ImportedPackages) {
	var names []string
	var eInfo *EnumInfo
	for idx, spec := range t.Specs {
		s.logger.Println(spec)
		switch stype := spec.(type) {
		case *ast.ValueSpec:
			name := stype.Names[0].Name
			names = append(names, name)
			if idx == 0 {
				// Check if the const block is an enum
				if stype.Type == nil {
					// No Type was attached so not an enum
					return
				}
				type_expr := stype.Type.(*ast.Ident)
				type_name := type_expr.Name
				if v, ok := s.Enums[type_name]; ok {
					eInfo = v
				}
			}
			// Not sure if we really need to store the values or not....
		}
	}
	if eInfo != nil {
		eInfo.ValNames = names
	}
}

func parseImports(module *ModuleInfo, imports []*ast.ImportSpec) (*ImportedPackages, error) {
	packages := &ImportedPackages{}
	packages.Imports = make(map[string]ImportedPackage)
	for _, imp := range imports {
		pkg := ImportedPackage{}
		pkg.ImportName = imp.Path.Value[1 : len(imp.Path.Value)-1]
		if imp.Name != nil {
			pkg.Name = imp.Name.Name
		} else {
			splits := strings.Split(pkg.ImportName, "/")
			pkg.Name = splits[len(splits)-1]
		}
		pkg.IsStandard = isStandardPackage(pkg.ImportName)
		if !pkg.IsStandard {
			mod, ver, err := module.FindRequires(pkg.ImportName)
			pkg.Module = mod
			pkg.ModuleVersion = ver
			if err != nil {
				return nil, err
			}
		}
		packages.Imports[pkg.Name] = pkg
	}
	return packages, nil
}

func (s *SpecParser) parsePackages(pkgs map[string]*PackageInfo) error {
	for path, pkg := range pkgs {
		for _, file := range pkg.Package.Files {
			imports, err := parseImports(pkg.Module, file.Imports)
			if err != nil {
				return err
			}

			for _, decl := range file.Decls {
				// Check if it is a GeneralDeclaration Block
				switch t := decl.(type) {
				case *ast.GenDecl:
					if t.Tok == token.TYPE {
						s.parseTypes(path, t, pkg, imports)
					} else if t.Tok == token.CONST {
						s.parseConstBlock(path, t, pkg, imports)
					}
				case *ast.FuncDecl:
					s.parseFunctions(t, pkg, imports)
				}
			}
		}
		// Associate functions with implementations
		s.associateFunctions()
		s.Functions = make(map[string][]*FuncInfo)
		s.ExtraFunctions = []*FuncInfo{}
	}
	return nil
}

func matchingMethod(f1 FuncInfo, f2 FuncInfo) bool {
	if (len(f1.Args) != len(f2.Args)) || (len(f1.Return) != len(f2.Return)) {
		return false
	}
	for idx, arg1 := range f1.Args {
		if !isSameType(arg1.Type, f2.Args[idx].Type) {
			return false
		}
	}
	for idx, ret1 := range f1.Return {
		if !isSameType(ret1.Type, f2.Return[idx].Type) {
			return false
		}
	}
	return true
}

func matchingMethods(iface *ServiceInfo, impl *ImplInfo) bool {
	for funcName, funcInfo := range iface.Methods {
		if implFunc, ok := impl.Methods[funcName]; !ok {
			// Implementation doesn't have a method that interface
			return false
		} else {
			// Implementation has the function name but the signature doesn't match
			if !matchingMethod(funcInfo, implFunc) {
				return false
			}
		}
	}
	return true
}

func (s *SpecParser) associateImplementations() {
	for _, impl := range s.Implementations {
		ifaces := make(map[string]bool)
		for _, iface := range s.Services {
			implements := matchingMethods(iface, impl)
			if implements {
				ifaces[iface.Name] = true
			}
		}
		impl.Interfaces = ifaces
	}
}

func (s *SpecParser) parseRemoteTypeStructs() {
	for name, impl := range s.Implementations {
		if _, ok := impl.Interfaces["Remote"]; ok {
			s.RemoteTypes[name] = impl
		}
	}
}

// Creates New Parser
func NewSpecParser(srcDirs ...string) *SpecParser {
	return &SpecParser{
		srcDirs:         srcDirs,
		logger:          log.Default(),
		Services:        make(map[string]*ServiceInfo),
		Implementations: make(map[string]*ImplInfo),
		Functions:       make(map[string][]*FuncInfo),
		ExtraFunctions:  []*FuncInfo{},
		RemoteTypes:     make(map[string]*ImplInfo),
		Packages:        make(map[string]*PackageInfo),
		Enums:           make(map[string]*EnumInfo),
		Modules:         make(map[string]*ModuleInfo),
	}
}

func ReadModfile(srcDir string) (*ModuleInfo, error) {
	srcDir = filepath.Clean(srcDir)
	modfilePath := filepath.Join(srcDir, "go.mod")
	modfileData, err := os.ReadFile(modfilePath)
	if err != nil {
		return nil, fmt.Errorf("unable to read workflow spec modfile %s due to %s", modfilePath, err.Error())
	}

	mod, err := modfile.Parse(modfilePath, modfileData, nil)
	if err != nil {
		return nil, fmt.Errorf("unable to parse %s due to %s", modfilePath, err.Error())
	}

	modInfo := &ModuleInfo{}
	modInfo.Name = mod.Module.Mod.Path
	modInfo.Path = srcDir
	modInfo.Version = mod.Module.Mod.Version
	modInfo.Requires = make(map[string]string)
	for _, req := range mod.Require {
		modInfo.Requires[req.Mod.Path] = req.Mod.Version
	}
	return modInfo, nil
}

/*
Finds the required module that contains the specified import name. Does so by
matching on the module prefix of the import name.  Returns an error if not found
*/
func (modInfo *ModuleInfo) FindRequires(importName string) (string, string, error) {
	for reqName, reqVersion := range modInfo.Requires {
		if strings.HasPrefix(importName, reqName) {
			return reqName, reqVersion, nil
		}
	}

	return "", "", fmt.Errorf("unable to find package in %v requires that imports %v", modInfo.Name, importName)
}

func (modInfo *ModuleInfo) String() string {
	b := strings.Builder{}
	b.WriteString("Module " + modInfo.Name + " " + modInfo.Version + "\n")
	for reqName, reqVersion := range modInfo.Requires {
		b.WriteString("  requires " + reqName + " " + reqVersion + "\n")
	}
	return b.String()
}

// Exported Parser function
func (s *SpecParser) ParseSpec() error {
	var fset token.FileSet
	for _, srcdir := range s.srcDirs {
		slog.Info("Parsing workflow spec module at " + srcdir)
		// JM: parse modfiles first.  each srcdir should be a golang module
		module, err := ReadModfile(srcdir)
		if err != nil {
			return err
		}
		s.Modules[module.Name] = module

		err = filepath.Walk(srcdir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if !info.IsDir() {
				return nil
			}
			pkgs, err := parser.ParseDir(&fset, path, nil, parser.ParseComments)
			if err != nil {
				return fmt.Errorf("unable to parse directory %s due to %s", path, err.Error())
			}

			for k, v := range pkgs {

				pkg := &PackageInfo{}
				pkg.ShortName = k
				pkg.Package = v
				pkg.Module = module
				pkg.Path = filepath.Clean(path)

				relPath, err := filepath.Rel(module.Path, pkg.Path)
				if err != nil {
					return fmt.Errorf("%s should exist within %s but got %s", pkg.Path, module.Path, err.Error())
				}

				pkg.Name = filepath.ToSlash(relPath)
				pkg.ImportName = module.Name + "/" + pkg.Name

				s.Packages[path] = pkg

				s.logger.Printf("Found package %s (import %s)", pkg.ShortName, pkg.ImportName)
			}
			return nil
		})
		if err != nil {
			return err
		}
	}
	err := s.parsePackages(s.Packages)
	if err != nil {
		return fmt.Errorf("unable to parse workflow packages due to %s", err.Error())
	}

	s.associateImplementations()
	s.parseRemoteTypeStructs()

	// s.logger.Println("# Total Service Declarations Found:", len(s.Services))
	// s.logger.Println("# Total Remote Type Declarations Found:", len(s.RemoteTypes))
	// s.PrintServices()
	// s.PrintImplementations()
	// s.PrintEnums()

	for name, info := range s.Services {
		s.logger.Printf("Found workflow service %s in %s\n", name, info.Package.ImportName)
	}

	return nil
}
