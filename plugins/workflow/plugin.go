package workflow

import (
	"fmt"
	"strings"

	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/core/service"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/golang"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/workflow/parser"
)

var workflowSpecPaths []string
var spec *parser.SpecParser

// Golang workflow must be initialized with a path to the workflow code
func Init(path string) {
	workflowSpecPaths = append(workflowSpecPaths, path)
	spec = nil
}

func getSpec() (*parser.SpecParser, error) {
	if spec == nil {
		spec = parser.NewSpecParser(workflowSpecPaths...)
		err := spec.ParseSpec()
		if err != nil {
			spec = nil
			return nil, err
		}
	}
	return spec, nil
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
		s.Name = impl.Name
		s.Package.Name = spec.PathPkgs[impl.PkgPath]
		s.Package.Path = impl.PkgPath

		// TODO: handle the cases of 0 and >1 constructors and interfaces

		s.Constructor.Name = impl.ConstructorInfos[0].Name
		s.Constructor.Args = argsToVars(impl.ConstructorInfos[0].Args)
		s.Constructor.Ret = argsToVars(impl.ConstructorInfos[0].Return)

		for iface, _ := range impl.Interfaces {
			service := spec.Services[iface]
			s.Interface.Name = service.Name
			for _, method := range service.Methods {
				s.Interface.Methods = append(s.Interface.Methods, funcToDecl(method))
			}
		}
		return &s, nil
	} else {
		return nil, fmt.Errorf("unable to find workflow spec service %s in any of the following locations: %s", serviceType, strings.Join(workflowSpecPaths, "; "))
	}
}
