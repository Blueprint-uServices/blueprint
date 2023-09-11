package gogen

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/blueprint"
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/core/irutil"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/golang"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/golang/gocode"
)

/*
Implements the GraphBuilder interface defined in gogen/ir.go

GraphBuilder is useful for any golang namespace node, such as the goproc plugin, that wants to be able
to instantiate Golang IR nodes.

The GraphBuilder accumulates code from child nodes and stubs for instantiating the code.

The GraphBuilder can then be used to generate a source file that constructs the graph.

The typical usage of a `GraphBuilder` is to:

 1. Create a new `GraphBuilder` with the `NewGraphBuilder` method

 2. Collect code definitions from child nodes by calling `Instantiable.AddInstantiation` on those nodes

 3. Generate the final output by calling `GraphBuilder.Finish`
*/
type GraphBuilderImpl struct {
	golang.GraphBuilder
	tracker      irutil.VisitTrackerImpl
	FileName     string             // The short name of the file
	FilePath     string             // The fully qualified path to the file
	module       *ModuleBuilderImpl // The module containing this file
	PackagePath  string             // The package path within the module
	Package      string             // The package name in the package declaration within the file
	FuncName     string             // The name of the function to generaet
	Imports      *Imports           // Import declarations in the file; map of shortname to full package import name
	Declarations map[string]string  // The DI declarations
}

/*
Create a new GraphBuilder
*/
func NewGraphBuilder(module *ModuleBuilderImpl, fileName, packagePath, funcName string) (*GraphBuilderImpl, error) {
	err := CheckDir(module.ModuleDir, false)
	if err != nil {
		return nil, fmt.Errorf("unable to generate %s for module %s due to %s", fileName, module.ShortName, err.Error())
	}

	packageDir := filepath.Join(module.ModuleDir, packagePath)
	err = CheckDir(packageDir, true)
	if err != nil {
		return nil, fmt.Errorf("unable to generate %s for module %s due to %s", fileName, module.ShortName, err.Error())
	}

	builder := &GraphBuilderImpl{}
	builder.FileName = fileName
	builder.FilePath = filepath.Join(packageDir, fileName)
	builder.module = module
	builder.PackagePath = packagePath
	splits := strings.Split(packagePath, "/")
	builder.Package = splits[len(splits)-1]
	builder.Imports = NewImports(packagePath)
	builder.Declarations = make(map[string]string)
	builder.FuncName = funcName

	// Add the runtime module as a dependency, in case it hasn't already
	builder.module.workspace.AddLocalModuleRelative("runtime", "../../../runtime")
	builder.module.Require("gitlab.mpi-sws.org/cld/blueprint/runtime", "v0.0.0")
	builder.Imports.AddPackage("gitlab.mpi-sws.org/cld/blueprint/runtime/plugins/golang")

	return builder, nil
}

func (code *GraphBuilderImpl) Visit(node blueprint.IRNode) error {
	if instantiable, ok := node.(golang.Instantiable); ok {
		return instantiable.AddInstantiation(code)
	}
	return nil
}

func (code *GraphBuilderImpl) Module() golang.ModuleBuilder {
	return code.module
}

func (code *GraphBuilderImpl) Import(packageName string) string {
	return code.Imports.AddPackage(packageName)
}

func (code *GraphBuilderImpl) ImportType(typeName gocode.TypeName) string {
	code.Imports.AddType(typeName)
	return code.Imports.NameOf(typeName)
}

func (code *GraphBuilderImpl) Declare(name, buildFuncSrc string) error {
	if _, exists := code.Declarations[name]; exists {
		return fmt.Errorf("generated file %s encountered redeclaration of %s", code.FileName, name)
	}
	code.Declarations[name] = buildFuncSrc
	return nil
}

func (code *GraphBuilderImpl) Visited(name string) bool {
	return code.tracker.Visited(name)
}

var diFuncTemplate = `package {{.Package}}

{{.Imports}}

func {{ .FuncName }}(args map[string]string) golang.Graph {
	g := golang.NewGraph()

	for k := range args {
		g.Define(k, func(ctr golang.Container) (any, error) { return args[k], nil })
	}

	{{ range $defName, $buildFunc := .Declarations }}
	g.Define("{{ $defName }}", {{ $buildFunc }})
	{{ end }}

	return g
}
`

/*
Generates the file within its module
*/
func (code *GraphBuilderImpl) Build() error {
	t, err := template.New(code.FuncName).Parse(diFuncTemplate)
	if err != nil {
		return err
	}

	f, err := os.OpenFile(code.FilePath, os.O_CREATE, 0755)
	if err != nil {
		return err
	}

	return t.Execute(f, code)
}
