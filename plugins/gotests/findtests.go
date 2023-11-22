package gotests

import (
	"fmt"
	"go/ast"

	"gitlab.mpi-sws.org/cld/blueprint/plugins/golang/gocode"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/golang/goparser"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/workflow"
)

// Metadata about a runtime registry.ServiceRegistry variable
// declaration in the workflow spec's test code
type serviceRegistry struct {
	VarName    string              // the var name of the registry
	RegistryOf gocode.TypeName     // the type parameter of the service registry
	Var        *goparser.ParsedVar // the parsed var
}

type serviceRegistries struct {
	spec       *workflow.WorkflowSpec
	registries []*serviceRegistry
}

func (r *serviceRegistries) Get(t *gocode.UserType) []*serviceRegistry {
	var matches []*serviceRegistry
	for _, v := range r.registries {
		if userType, isUserType := v.RegistryOf.(*gocode.UserType); isUserType {
			if userType.Package == t.Package && userType.Name == t.Name {
				matches = append(matches, v)
			}
		}
	}
	return matches
}

// Finds all variables within the workflow spec code that are instances of
// the runtime registry.ServiceRegistry type.
//
// This is used by the gotests plugin to modify workflow unit tests to inject
// client instances during the Blueprint compilation process
//
// This implementation only currently matches vars that are declared with a
// direct call to registry.NewServiceRegistry, e.g.
//
//	xxx := registry.NewServiceRegistry[xxx](xxx)
func findWorkflowServiceRegistries() (*serviceRegistries, error) {
	var err error
	r := &serviceRegistries{}

	// Load the workflow spec
	if r.spec, err = workflow.GetSpec(); err != nil {
		return nil, err
	}

	// Find instances
	for _, mod := range r.spec.Parsed.Modules {
		for _, pkg := range mod.Packages {
			for _, v := range pkg.Vars {
				vType, isRegistry := isRegistryVar(v)
				if isRegistry {
					fmt.Printf("%v is a registry of %v\n", v.Name, vType)
					r.registries = append(r.registries, &serviceRegistry{
						VarName:    v.Name,
						RegistryOf: vType,
						Var:        v,
					})
				}
			}
		}
	}

	return r, nil
}

func isRegistryVar(v *goparser.ParsedVar) (gocode.TypeName, bool) {
	typeParam, isRegistryCall := isTestRegistryCall(v.Ast)
	if !isRegistryCall {
		return nil, false
	}

	return v.File.ResolveType(typeParam), true
}

// Check for an expr of the form:
// something := registry.NewServiceRegistry[user.UserService]()
// returns the ast.Expr for the type parameter of NewServiceRegistry
func isTestRegistryCall(v *ast.ValueSpec) (ast.Expr, bool) {
	if len(v.Values) != 1 {
		return nil, false
	}

	if call, isCall := v.Values[0].(*ast.CallExpr); isCall {
		if indx, isIndex := call.Fun.(*ast.IndexExpr); isIndex {
			if sel, isSel := indx.X.(*ast.SelectorExpr); isSel {
				if x, xIsIdent := sel.X.(*ast.Ident); xIsIdent {
					if x.Name == "registry" && sel.Sel.Name == "NewServiceRegistry" {
						return indx.Index, true
					}
				}
			}
		}
	}
	return nil, false
}
