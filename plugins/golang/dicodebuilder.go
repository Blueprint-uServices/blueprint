package golang

import (
	"fmt"
	"path/filepath"
	"strings"
)

/*
This is used for accumulating DI code definitions and generating a file that
constructs and builds the objects
*/
type DICodeBuilder struct {
	VisitTracker
	FileName     string            // The short name of the file
	Module       *ModuleBuilder    // The module containing this file
	PackagePath  string            // The package path within the module
	Package      string            // The package name in the package declaration within the file
	FuncName     string            // The name of the function to generaet
	Imports      map[string]string // Import declarations in the file; map of shortname to full package import name
	Declarations map[string]string // The DI declarations
}

/*
This method is used by plugins if they want to generate code that instantiates other nodes.

After the DI declarations of nodes have been added to the code builder, plugins must call
Generate to finish building the file within the module.
*/
func NewDICodeBuilder(module *ModuleBuilder, fileName, packagePath, funcName string) (*DICodeBuilder, error) {
	err := checkDir(module.ModuleDir, false)
	if err != nil {
		return nil, fmt.Errorf("unable to generate %s for module %s due to %s", fileName, module.ShortName, err.Error())
	}

	packageDir := filepath.Join(module.ModuleDir, packagePath)
	err = checkDir(packageDir, true)
	if err != nil {
		return nil, fmt.Errorf("unable to generate %s for module %s due to %s", fileName, module.ShortName, err.Error())
	}

	builder := &DICodeBuilder{}
	builder.visited = make(map[string]any)
	builder.FileName = fileName
	builder.Module = module
	builder.PackagePath = packagePath
	splits := strings.Split(packagePath, "/")
	builder.Package = splits[len(splits)-1]
	builder.Imports = make(map[string]string)
	builder.Declarations = make(map[string]string)
	builder.FuncName = funcName
	return builder, nil
}

/*
Adds an import to the generated file; this is necessary for any types declared in other
packages that are going to be used in a DI declaration.  This method returns the type
alias that should be used in the generated code.  By default the type alias is just the
package name, but if there are multiple different imports with the same package name, then
aliases will be created
*/
func (code *DICodeBuilder) Import(packageName string) string {
	splits := strings.Split(packageName, "/")
	shortName := splits[len(splits)-1]
	suffix := 0
	name := shortName
	for {
		if _, nameInUse := code.Imports[name]; !nameInUse {
			code.Imports[name] = packageName
			return name
		}
		suffix += 1
		name = fmt.Sprintf("%s%v", shortName, suffix)
	}
}

/*
Adds DI declaration code to the generated file
*/
func (code *DICodeBuilder) Declare(name, buildFunc string) error {
	if _, exists := code.Declarations[name]; exists {
		return fmt.Errorf("generated file %s encountered redeclaration of %s", code.FileName, name)
	}
	code.Declarations[name] = buildFunc
	return nil
}

/*
Generates the file within its module
*/
func (code *DICodeBuilder) Finish() error {
	// TODO generate go file with method that receives map as input, gives map as output
	return nil
}
