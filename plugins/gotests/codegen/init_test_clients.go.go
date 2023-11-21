package codegen

import (
	"fmt"
	"path/filepath"

	"gitlab.mpi-sws.org/cld/blueprint/plugins/golang/gocode"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/golang/gogen"
	"golang.org/x/exp/slog"
)

// packageDir is a fully qualified output directory containing the test code and registry variable.
// packageName is the fully qualified name of the output package
// registryVar is the variable name within the package that is a ServiceRegistry
// clientName is the name of the client to register in the registry
// nodeToInstantiate is the node to instantiate when the client gets created
// clientType is the type of the client and the type parameter of the ServiceRegistry
func AddClientToTests(packageDir, packageName, packageShortName, registryVar, clientName, nodeToInstantiate string, clientType gocode.TypeName) error {
	filename := fmt.Sprintf("init_%v_%v_client.go", clientName, registryVar)

	templateArgs := TestClientArgs{
		Package:           packageName,
		PackageShortName:  packageShortName,
		RegistryVarName:   registryVar,
		ClientName:        clientName,
		NodeToInstantiate: nodeToInstantiate,
		ClientType:        clientType,
		Imports:           gogen.NewImports(packageName),
	}

	templateArgs.Imports.AddPackages("context")

	slog.Info(fmt.Sprintf("Generating %v/%v.go", packageDir, filename))
	outputFile := filepath.Join(packageDir, filename)
	return gogen.ExecuteTemplateToFile("ServiceRegistryTestClientInit", initTestClientsTemplate, templateArgs, outputFile)
}

type TestClientArgs struct {
	Package           string          // The name of the package containing the TestRegistry
	PackageShortName  string          // The package name used in the package declaration
	RegistryVarName   string          // The name of the variable in the test code containing the registry.ServiceRegistry
	ClientName        string          // The name to give that client that will be added to the registry
	NodeToInstantiate string          // The node in the namespace to instantiate when creating the client
	ClientType        gocode.TypeName // The type of the registry and client

	Imports *gogen.Imports // Manages imports for us
}

var initTestClientsTemplate = `
package {{ .PackageShortName }}

{{ .Imports }}

func init() {
	{{ .RegistryVarName }}.Register("{{ .ClientName }}", func(ctx context.Context) ({{ NameOf .ClientType }}, error) {
		// TODO
		return nil, nil
	})
}
`
