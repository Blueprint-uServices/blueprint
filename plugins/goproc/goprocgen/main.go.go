package goprocgen

import (
	"fmt"
	"path/filepath"

	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/blueprint"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/golang"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/golang/gogen"
	"golang.org/x/exp/slog"
)

/*
Generates a main.go file in the provided module.  The main method will
call the graphConstructor provided to create and instantiate nodes.
*/
func GenerateMain(
	name string,
	argNodes []blueprint.IRNode,
	nodesToInstantiate []blueprint.IRNode,
	module golang.ModuleBuilder,
	graphPackage string,
	graphConstructor string) error {

	// Generate the main.go
	mainArgs := mainTemplateArgs{
		Name:             name,
		GraphPackage:     graphPackage,
		GraphConstructor: graphConstructor,
		Args:             nil,
		Instantiate:      nil,
	}

	// Expect command-line arguments for all argNodes specified
	for _, arg := range argNodes {
		mainArgs.Args = append(mainArgs.Args, mainArg{
			Name: arg.Name(),
			Doc:  arg.String(),
			Var:  blueprint.CleanName(arg.Name()),
		})
	}

	// Instantiate the nodes specified
	for _, node := range nodesToInstantiate {
		if _, isInstantiable := node.(golang.Instantiable); isInstantiable {
			mainArgs.Instantiate = append(mainArgs.Instantiate, node.Name())
		}
	}
	slog.Info(fmt.Sprintf("Generating %v/main.go", module.Info().Name))
	mainFileName := filepath.Join(module.Info().Path, "main.go")
	return gogen.ExecuteTemplateToFile("goprocMain", mainTemplate, mainArgs, mainFileName)
}

type mainArg struct {
	Name string
	Doc  string
	Var  string
}

type mainTemplateArgs struct {
	Name             string
	GraphPackage     string
	GraphConstructor string
	Args             []mainArg
	Instantiate      []string
}

var mainTemplate = `// This file is auto-generated by the Blueprint goproc plugin
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strings"

	"{{.GraphPackage}}"
	"golang.org/x/exp/slog"
)

var missingArgs []string
func checkArg(name, value string) {
	if value == "" {
		missingArgs = append(missingArgs, name)
	} else {
		slog.Info(fmt.Sprintf("%v = %v", name, value))
	}
}

func main() {
	slog.Info("Running {{.Name}}")
	{{- range $i, $arg := .Args}}
	{{$arg.Var}} := flag.String("{{$arg.Name}}", "", "Argument automatically generated from Blueprint IR: {{$arg.Doc}}")
	{{end}}

	flag.Parse()

	{{range $i, $arg := .Args -}}
	checkArg("{{$arg.Name}}", *{{$arg.Var}})
	{{end}}
	if len(missingArgs) > 0 {
		slog.Error(fmt.Sprintf("Missing required arguments: \n  %v", strings.Join(missingArgs, "\n  ")))
		os.Exit(1)
	}
	
	graphArgs := map[string]string{
		{{- range $i, $arg := .Args}}
		"{{$arg.Name}}": *{{$arg.Var}},
		{{- end}}
	}

	ctx, cancel := context.WithCancel(context.Background())
	graph, err := {{.GraphConstructor}}(ctx, cancel, graphArgs, nil, "{{.Name}}")
	if err != nil {
		slog.Error(err.Error())
		return
	}

	var node any
	{{range $i, $node := .Instantiate -}}
	if err = graph.Get("{{$node}}", &node); err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}
	{{end}}
	
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt)
	go func() {
		for sig := range signals {
			slog.Info(fmt.Sprintf("{{.Name}} received %v\n", sig))
			cancel()
		}
	}()

	graph.WaitGroup().Wait()

	slog.Info("{{.Name}} exiting")
}`