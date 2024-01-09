// Package codegen implements the gotest plugin's code generation logic.
//
// Generates the blueprint_clients.go file that gets added to test packages.
// This is used internally by the gotest package.
//
// A blueprint_clients.go file is generated for any package where a [registry.ServiceRegistry]
// is used.  The file adds a static initialization block that registers an instance
// of the 'real' application client.
//
// [registry.ServiceRegistry]: https://github.com/blueprint-uservices/blueprint/tree/main/runtime/core/registry
package codegen

import (
	"fmt"
	"path/filepath"

	"github.com/blueprint-uservices/blueprint/plugins/golang/gocode"
	"github.com/blueprint-uservices/blueprint/plugins/golang/gogen"
	"golang.org/x/exp/slog"
)

// Used by the gotests plugin to generate the blueprint_clients.go file
type ClientBuilder struct {
	PackageShortName     string // The package name to use in the package declaration
	NamespaceConstructor string // The func that creates the namespace
	NamespaceName        string // The name to use for the namespace
	OutputDir            string // The output directory; should be the same as the package directory
	Clients              []*clientRegistration
	Imports              *gogen.Imports // Manages imports for us
}

type clientRegistration struct {
	ClientName        string          // The name of the client we're registering
	RegistryVarName   string          // The name of the ServiceRegistry variable within this package
	ClientType        gocode.TypeName // The client type managed by the ServiceRegistry
	NodeToInstantiate string          // The node in the namespace to instantiate when creating the client
}

// Create a new builder to generate a blueprint_clients.go file.
//
//   - outputDir points to an output directory
//   - packageName should correspond to the correct fully-qualified package name of the outputDir
//   - packageShortName should correspond to the name to use in the "package" declaration of the file
//   - namespacePackage is the package to import that contains namespaceConstructor
//   - namespaceConstructor is of the form shortname.Method - it is the method to call to build the client library
//   - namespaceName can be any name
func NewClientBuilder(packageName, packageShortName, namespaceConstructor, namespacePackage, namespaceName, outputDir string) *ClientBuilder {
	b := &ClientBuilder{
		PackageShortName:     packageShortName,
		NamespaceConstructor: namespaceConstructor,
		NamespaceName:        namespaceName,
		OutputDir:            outputDir,
		Imports:              gogen.NewImports(packageName),
	}

	b.Imports.AddPackages("context", namespacePackage)

	return b
}

// Add a client registration to the generated blueprint_clients.go file
//
//   - registryVar is a variable name within the output package of a ServiceRegistry[clientType]
//   - clientName can be any name
//   - nodeToInstantiate is the node within the namespace to Get to create the client
//   - clientType is the service interface being created.
func (b *ClientBuilder) AddClient(registryVar, clientName, nodeToInstantiate string, clientType gocode.TypeName) {
	r := &clientRegistration{
		ClientName:        clientName,
		RegistryVarName:   registryVar,
		ClientType:        clientType,
		NodeToInstantiate: nodeToInstantiate,
	}
	b.Clients = append(b.Clients, r)
}

// Generate the blueprint_clients.go file.
func (b *ClientBuilder) Build() error {
	filename := "blueprint_clients.go"

	slog.Info(fmt.Sprintf("Generating %v/%v.go", b.OutputDir, filename))
	outputFile := filepath.Join(b.OutputDir, filename)
	return gogen.ExecuteTemplateToFile("ServiceRegistryTestClientInit", initTestClientsTemplate, b, outputFile)
}

var initTestClientsTemplate = `
package {{ .PackageShortName }}

{{ .Imports }}

// Auto-generated code by the Blueprint gotests plugin.
func init() {
	// Initialize the clientlib early so that it can pick up command-line flags
	clientlib := {{ .NamespaceConstructor }}("{{ .NamespaceName }}")

	{{ range $_, $client := .Clients }}
	{{ .RegistryVarName }}.Register("{{ .ClientName }}", func(ctx context.Context) ({{ NameOf .ClientType }}, error) {
		// Build the client library
		namespace, err := clientlib.Build(ctx)
		if err != nil {
			return nil, err
		}

		// Get and return the client
		var client {{ NameOf .ClientType }}
		err = namespace.Get("{{ .NodeToInstantiate }}", &client)
		return client, err
	})
	{{end}}
}
`
