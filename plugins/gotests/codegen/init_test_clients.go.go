package codegen

import (
	"fmt"
	"path/filepath"

	"gitlab.mpi-sws.org/cld/blueprint/plugins/golang/gocode"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/golang/gogen"
	"golang.org/x/exp/slog"
)

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

func (b *ClientBuilder) AddClient(registryVar, clientName, nodeToInstantiate string, clientType gocode.TypeName) {
	r := &clientRegistration{
		ClientName:        clientName,
		RegistryVarName:   registryVar,
		ClientType:        clientType,
		NodeToInstantiate: nodeToInstantiate,
	}
	b.Clients = append(b.Clients, r)
}

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
