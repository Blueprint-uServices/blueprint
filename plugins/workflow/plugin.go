package workflow

import (
	"fmt"
	"path/filepath"
	"runtime"
	"strings"

	cp "github.com/otiai10/copy"
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/core/service"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/golang"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/workflow/parser"
)

var workflowSpecPaths []string
var spec *parser.SpecParser

// Golang workflow must be initialized with a path to the workflow code, relative to the calling file
func Init(path string) {
	_, callingFile, _, _ := runtime.Caller(1)
	dir, _ := filepath.Split(callingFile)
	workflowPath := filepath.Join(dir, path)
	workflowSpecPaths = append(workflowSpecPaths, workflowPath)
	spec = nil
}

func getSpec() (*parser.SpecParser, error) {
	if spec == nil {
		var fqPaths []string
		for _, path := range workflowSpecPaths {
			fqPath, err := filepath.Abs(path)
			if err != nil {
				return nil, fmt.Errorf("invalid workflow spec path %s due to %s", fqPath, err.Error())
			}
			fqPaths = append(fqPaths, fqPath)
		}

		// TODO: tidy up legacy spec parser
		spec = parser.NewSpecParser(fqPaths...)
		err := spec.ParseSpec()
		if err != nil {
			spec = nil
			return nil, err
		}
	}
	return spec, nil
}

func CopyWorkflowSpec(dstPath string) error {
	for _, srcPath := range workflowSpecPaths {
		err := cp.Copy(srcPath, dstPath)
		if err != nil {
			return err
		}
	}
	return nil
}

// Convert from parser representation to IR representation
func argsToVars(as []parser.ArgInfo) (vs []service.Variable) {
	for _, a := range as {
		v := service.Variable{Name: a.Name, Type: a.Type.String()}
		vs = append(vs, v)
	}
	return
}

func funcToDecl(f parser.FuncInfo) (d service.ServiceMethodDeclaration) {
	d.Name = f.Name
	d.Args = argsToVars(f.Args)
	d.Ret = argsToVars(f.Return)
	return
}

// Finds the service with the specified type in the workflow spec.
// This method searches the WorkflowSpecPath and returns an error if not found.
func findService(serviceType string) (*golang.GolangServiceDetails, error) {
	spec, err := getSpec()
	if err != nil {
		return nil, err
	}

	if impl, exists := spec.Implementations[serviceType]; exists {
		s := golang.GolangServiceDetails{}

		for iface, _ := range impl.Interfaces {
			service := spec.Services[iface]
			s.Interface.Name = service.Name
			for _, method := range service.Methods {
				s.Interface.Methods = append(s.Interface.Methods, funcToDecl(method))
			}
			s.InterfacePackage = service.Package
			break
		}

		/*
		 TODO:  the spec parser needs to correctly do the following:
		  - get the module name that contains the service interface.
		    this
		  - get the module name that contains the instance constructor (if different)

		*/

		s.ImplName = impl.Name
		s.ImplConstructor.Name = impl.ConstructorInfos[0].Name
		s.ImplConstructor.Args = argsToVars(impl.ConstructorInfos[0].Args)
		s.ImplConstructor.Ret = argsToVars(impl.ConstructorInfos[0].Return)

		s.ImplPackage = impl.Package

		// TODO: handle the cases of 0 and >1 constructors and interfaces

		fmt.Printf("Impl:\n\n%v\n\n", s)

		return &s, nil
	} else {
		return nil, fmt.Errorf("unable to find workflow spec service %s in any of the following locations: %s", serviceType, strings.Join(workflowSpecPaths, "; "))
	}
}
