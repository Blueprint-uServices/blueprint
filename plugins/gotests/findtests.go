package gotests

import (
	"fmt"
	"go/ast"
	"strings"

	"github.com/blueprint-uservices/blueprint/plugins/golang/gocode"
	"github.com/blueprint-uservices/blueprint/plugins/golang/goparser"
	"github.com/blueprint-uservices/blueprint/plugins/workflow/workflowspec"
	"golang.org/x/exp/slog"
)

// Metadata about a runtime registry.ServiceRegistry variable
// declaration in the workflow spec's test code
type serviceRegistry struct {
	VarName    string                   // the var name of the registry
	RegistryOf gocode.TypeName          // the type parameter of the service registry
	Iface      *gocode.ServiceInterface // the interface of the type parameter of the service registry
	Var        *goparser.ParsedVar      // the parsed var
}

type serviceRegistries struct {
	registries []*serviceRegistry
}

func (r *serviceRegistries) Get(iface *gocode.ServiceInterface) []*serviceRegistry {
	var matches []*serviceRegistry
	for _, v := range r.registries {
		if iface.UserType.Equals(v.RegistryOf) {
			// A registry exists paramaterized with the same interface as iface
			matches = append(matches, v)
		} else if iface.Contains(v.Iface) {
			// A registry exists paramaterized with a sub-interface of iface
			matches = append(matches, v)
		}
	}
	return matches
}

var verbose = false

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
	r := &serviceRegistries{}

	// Load the workflow spec
	modset := workflowspec.Get().Modules

	// Find instances
	names := []string{}
	for _, mod := range modset.Modules {
		for _, pkg := range mod.Packages {
			for _, v := range pkg.Vars {
				vType, isRegistry := isRegistryVar(v)
				if isRegistry {
					registryInfo := &serviceRegistry{
						VarName:    v.Name,
						RegistryOf: vType,
						Var:        v,
					}

					// Try to extract the interface of the registry
					if u, isU := vType.(*gocode.UserType); isU {
						srv, err := workflowspec.GetServiceByName(u.Package, u.Name)
						if err == nil {
							registryInfo.Iface = srv.Iface.ServiceInterface(nil)
							if verbose {
								slog.Info(fmt.Sprintf("Found %v registry %v\n", registryInfo.Iface, v.Name))
							}
						}
					}

					r.registries = append(r.registries, registryInfo)
					names = append(names, v.Name)
				}
			}
		}
	}

	if verbose {
		slog.Info(fmt.Sprintf("Found %v service registries [%v]", len(r.registries), strings.Join(names, ", ")))
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
