package gocodegen

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/core/irutil"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/golang"
)

type DICodeBuilderImpl struct {
	golang.DICodeBuilder
	tracker      irutil.VisitTrackerImpl
	FileName     string             // The short name of the file
	FilePath     string             // The fully qualified path to the file
	module       *ModuleBuilderImpl // The module containing this file
	PackagePath  string             // The package path within the module
	Package      string             // The package name in the package declaration within the file
	FuncName     string             // The name of the function to generaet
	Imports      map[string]string  // Import declarations in the file; map of shortname to full package import name
	Declarations map[string]string  // The DI declarations
}

/*
Plugins rarely need to instantiate their own DICodeBuilder; this is typically only needed if an IRNode
is a namespace (for example, if it is the main method of a Golang process).

The DICodeBuilder can be used to accumulate code definitions from child nodes, if those child nodes implement
the Instantiable interface.  The `DICodeBuilder“ then combines those code definitions and generates a method
that instantiates those child nodes.

The typical usage of a `DICodeBuilder` is to:

 1. Create a new `DICodeBuilder` with the `NewDICodeBuilder` method

 2. Collect code definitions from child nodes by calling `Instantiable.AddInstantiation` on those nodes

 3. Generate the final output by calling `DICodeBuilder.Finish`
*/
func NewDICodeBuilder(module *ModuleBuilderImpl, fileName, packagePath, funcName string) (*DICodeBuilderImpl, error) {
	err := CheckDir(module.ModuleDir, false)
	if err != nil {
		return nil, fmt.Errorf("unable to generate %s for module %s due to %s", fileName, module.ShortName, err.Error())
	}

	packageDir := filepath.Join(module.ModuleDir, packagePath)
	err = CheckDir(packageDir, true)
	if err != nil {
		return nil, fmt.Errorf("unable to generate %s for module %s due to %s", fileName, module.ShortName, err.Error())
	}

	builder := &DICodeBuilderImpl{}
	builder.FileName = fileName
	builder.FilePath = filepath.Join(packageDir, fileName)
	builder.module = module
	builder.PackagePath = packagePath
	splits := strings.Split(packagePath, "/")
	builder.Package = splits[len(splits)-1]
	builder.Imports = make(map[string]string)
	builder.Declarations = make(map[string]string)
	builder.FuncName = funcName

	// Add the runtime module as a dependency, in case it hasn't already
	builder.module.workspace.AddLocalModuleRelative("runtime", "../../../runtime")
	builder.module.Require("gitlab.mpi-sws.org/cld/blueprint/runtime", "v0.0.0")

	return builder, nil
}

func (code *DICodeBuilderImpl) Module() golang.ModuleBuilder {
	return code.module
}

func (code *DICodeBuilderImpl) Import(packageName string) string {
	splits := strings.Split(packageName, "/")
	shortName := splits[len(splits)-1]
	suffix := 0
	name := shortName
	for {
		if pkg, nameInUse := code.Imports[name]; !nameInUse || pkg == packageName {
			code.Imports[name] = packageName
			return name
		}
		suffix += 1
		name = fmt.Sprintf("%s%v", shortName, suffix)
	}
}
func (code *DICodeBuilderImpl) Declare(name, buildFuncSrc string) error {
	if _, exists := code.Declarations[name]; exists {
		return fmt.Errorf("generated file %s encountered redeclaration of %s", code.FileName, name)
	}
	code.Declarations[name] = buildFuncSrc
	return nil
}

func (code *DICodeBuilderImpl) Visited(name string) bool {
	return code.tracker.Visited(name)
}

var diFuncTemplate = `package {{.Package}}

import (
	{{ range $importAs, $package := .Imports }}{{ $importAs }} "{{ $package }}"
	{{ end }}
	"gitlab.mpi-sws.org/cld/blueprint/runtime/plugins/golang"
)

func {{ .FuncName }}(args map[string]string) golang.Graph {
	g := golang.NewGraph()

	{{ range $defName, $buildFunc := .Declarations }}
	g.Define("{{ $defName }}" ,{{ $buildFunc }})
	{{ end }}

	return g
}

`

// `module {{.Name}}

// go 1.20

// {{ range $name, $version := .Requires }}require {{ $name }} {{ $version }}
// {{ end }}
// `

/*
Generates the file within its module
*/
func (code *DICodeBuilderImpl) Finish() error {
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
