package clientpool

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/blueprint-uservices/blueprint/blueprint/pkg/blueprint/stringutil"
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/coreplugins/service"
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/ir"
	"github.com/blueprint-uservices/blueprint/plugins/golang"
	"github.com/blueprint-uservices/blueprint/plugins/golang/gocode"
	"github.com/blueprint-uservices/blueprint/plugins/golang/gogen"
	"golang.org/x/exp/slog"
)

// Blueprint IR node representing a ClientPool that uses [N] instances of [Client]
type ClientPool struct {
	golang.Service
	golang.GeneratesFuncs

	PoolName string
	N        int
	Client   golang.Service
	Edges    []ir.IRNode
	Nodes    []ir.IRNode
}

// Implements ir.IRNode
func (pool *ClientPool) Name() string {
	return pool.PoolName
}

// Implements ir.IRNode
func (pool *ClientPool) String() string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("%v = ClientPool(%v, %v) {\n", pool.PoolName, pool.Client.Name(), pool.N))
	var children []string
	for _, child := range pool.Nodes {
		children = append(children, child.String())
	}
	b.WriteString(stringutil.Indent(strings.Join(children, "\n"), 2))
	b.WriteString("\n}")
	return b.String()
}

// Implements golang.Service service.ServiceNode
func (pool *ClientPool) GetInterface(ctx ir.BuildContext) (service.ServiceInterface, error) {
	/* ClientPool doesn't modify the client's interface and doesn't introduce new interfaces */
	return pool.Client.GetInterface(ctx)
}

// Implements golang.Service golang.ProvidesInterface
func (pool *ClientPool) AddInterfaces(module golang.ModuleBuilder) error {
	/* ClientPool doesn't modify the client's interface and doesn't introduce new interfaces */
	for _, node := range pool.Nodes {
		if n, valid := node.(golang.ProvidesInterface); valid {
			if err := n.AddInterfaces(module); err != nil {
				return err
			}
		}
	}
	return nil
}

// Implements golang.GeneratesFuncs
func (pool *ClientPool) GenerateFuncs(module golang.ModuleBuilder) error {
	/* Only generate clientpool code for the wrapped client once */
	iface, err := golang.GetGoInterface(module, pool.Client)
	if err != nil {
		return err
	}
	if module.Visited(iface.Name + "_ClientPool") {
		return nil
	}

	// Make sure we have all necessary code of contained nodes
	for _, node := range pool.Nodes {
		if n, valid := node.(golang.GeneratesFuncs); valid {
			if err := n.GenerateFuncs(module); err != nil {
				return err
			}
		}
	}

	// First generate the namespace code used by the clientpool
	args, err := pool.getTemplateArgs(module)
	if err != nil {
		return err
	}
	namespaceBuilder, err := gogen.NewNamespaceBuilder(module, args.Service.BaseName+"_PoolClient", args.ClientFileName, args.PackageShortName, args.ClientConstructor)
	if err != nil {
		return err
	}

	// Add instantiation code for everything within the pool
	for _, node := range pool.Nodes {
		if inst, canInstantiate := node.(golang.Instantiable); canInstantiate {
			if err := inst.AddInstantiation(namespaceBuilder); err != nil {
				return err
			}
		}
	}

	// Generate the namespace code
	if err = namespaceBuilder.Build(); err != nil {
		return err
	}

	// Generate the client pool code
	poolFileName := filepath.Join(module.Info().Path, args.PackageShortName, args.PoolFileName)
	return gogen.ExecuteTemplateToFile("clientpool_client_constructor", poolTemplate, args, poolFileName)
}

// Implements golang.Service golang.Instantiable
func (pool *ClientPool) AddInstantiation(builder golang.NamespaceBuilder) error {
	if builder.Visited(pool.PoolName) {
		return nil
	}

	args, err := pool.getTemplateArgs(builder.Module())
	if err != nil {
		return err
	}

	builder.Import(args.PackageName)

	slog.Info(fmt.Sprintf("Instantiating ClientPool %v in %v/%v", pool.PoolName, builder.Info().Package.PackageName, builder.Info().FileName))
	code, err := gogen.ExecuteTemplate("clientpool", buildPoolTemplate, args)
	if err != nil {
		return err
	}
	return builder.Declare(pool.PoolName, code)
}

func (pool *ClientPool) getTemplateArgs(module golang.ModuleBuilder) (*templateArgs, error) {
	var err error
	args := &templateArgs{}
	args.Service, err = golang.GetGoInterface(module, pool.Client)
	if err != nil {
		return nil, err
	}
	args.WrappedClient = pool.Client.Name()
	args.InstanceName = pool.PoolName
	args.MaxClients = pool.N
	args.PoolName = args.Service.Name + "_ClientPool"
	args.PackageShortName = "pool"
	args.PackageName = module.Info().Name + "/" + args.PackageShortName
	args.ClientFileName = args.Service.BaseName + "_pool_client.go"
	args.PoolFileName = args.Service.BaseName + "_pool.go"
	args.ClientConstructor = fmt.Sprintf("New_%v_PoolClient", args.Service.BaseName)
	args.PoolConstructor = fmt.Sprintf("New_%v_Pool", args.Service.BaseName)
	args.Imports = gogen.NewImports(args.PackageName)

	args.Imports.AddPackages(
		"context", "fmt",
		"github.com/blueprint-uservices/blueprint/runtime/plugins/clientpool",
		"github.com/blueprint-uservices/blueprint/runtime/plugins/golang",
	)
	return args, nil
}

type (
	templateArgs struct {
		InstanceName      string
		WrappedClient     string
		MaxClients        int
		PoolName          string
		PackageShortName  string
		PackageName       string
		ClientFileName    string
		PoolFileName      string
		ClientConstructor string
		PoolConstructor   string
		Service           *gocode.ServiceInterface
		Imports           *gogen.Imports
	}
)

var buildPoolTemplate = `func(n *golang.Namespace) (any, error) {
		return pool.{{.PoolConstructor}}(n), nil
	}`

var poolTemplate = `// This file is auto-generated by the Blueprint clientpool plugin
package {{.PackageShortName}}

{{.Imports}}

type {{.PoolName}} struct {
	clients *clientpool.ClientPool[{{NameOf .Service.UserType}}]
}

func {{.PoolConstructor}}(parent *golang.Namespace) *{{.PoolName}} {
	i := 0
	createClient := func() ({{NameOf .Service.UserType}}, error) {
		clientName := fmt.Sprintf("{{.InstanceName}}.%v", i)
		n, err := {{.ClientConstructor}}(clientName).BuildWithParent(parent)
		if err != nil {
			return nil, err
		}
		i++
		var client {{NameOf .Service.UserType}}
		err = n.Get("{{.WrappedClient}}", &client)
		return client, err
	}
	clients := clientpool.NewClientPool({{.MaxClients}}, createClient)
	return &{{.PoolName}}{clients: clients}
}

{{$service := .Service -}}
{{$receiver := .PoolName -}}
{{ range $_, $f := .Service.Methods }}
func (pool *{{$receiver}}) {{SignatureWithRetVars $f}} {
	client, err := pool.clients.Pop(ctx)
	if err != nil {
		return
	}
	defer pool.clients.Push(client)
	return client.{{$f.Name}}({{ArgVars $f "ctx"}})
}
{{end}}


`
