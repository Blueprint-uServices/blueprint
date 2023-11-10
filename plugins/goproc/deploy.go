package goproc

import (
	"fmt"
	"strings"

	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/ir"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/golang"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/golang/gogen"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/goproc/goprocgen"
	"golang.org/x/exp/slog"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

/*
The default goproc deployer doesn't assume anything about the target environment, and simply
packages the go code into a process with a main method.  It is assumed that the user
or caller will install Go and any dependencies.

The builder used to generate the workspace is located in gogen/workspacebuilder.go
*/

type filesystemDeployer interface {
	ir.ArtifactGenerator
}

/*
Generates a golang process to a directory on the local filesystem.

This will collect and package all of the code for the contained Golang nodes
and generate a main.go method.

The output code will be runnable on the local filesystem, assuming the
user has configured the appropriate environment
*/
func (node *Process) GenerateArtifacts(workspaceDir string) error {
	slog.Info(fmt.Sprintf("Building goproc %s to %s", node.Name(), workspaceDir))
	workspace, err := gogen.NewWorkspaceBuilder(workspaceDir)
	if err != nil {
		return err
	}

	// Add relevant nodes to the workspace
	for _, node := range node.ContainedNodes {
		if n, valid := node.(golang.ProvidesModule); valid {
			if err := n.AddToWorkspace(workspace); err != nil {
				return err
			}
		}
	}

	// Create the module
	slog.Info(fmt.Sprintf("Creating module %v", node.ModuleName))
	module, err := gogen.NewModuleBuilder(workspace, node.ModuleName)
	if err != nil {
		return err
	}

	// Add and/or generate interfaces
	for _, node := range node.ContainedNodes {
		if n, valid := node.(golang.ProvidesInterface); valid {
			if err := n.AddInterfaces(module); err != nil {
				return err
			}
		}
	}

	// Generate constructors and function declarations
	for _, node := range node.ContainedNodes {
		if n, valid := node.(golang.GeneratesFuncs); valid {
			if err := n.GenerateFuncs(module); err != nil {
				return err
			}
		}
	}

	// Create the method to instantiate the graph
	graphFileName := strings.ToLower(node.ProcName) + ".go"
	procPackage := "goproc"
	constructorName := "New" + cases.Title(language.BritishEnglish).String(node.ProcName)
	graph, err := gogen.NewGraphBuilder(module, graphFileName, procPackage, constructorName)
	if err != nil {
		return err
	}

	// Add constructor invocations
	for _, node := range node.ContainedNodes {
		if n, valid := node.(golang.Instantiable); valid {
			if err := n.AddInstantiation(graph); err != nil {
				return err
			}
		}
	}

	// TODO: it's possible some metadata / address nodes are residing in this namespace.  They don't
	// get passed in as args, but need to be added to the graph nonetheless

	// Generate the graph code
	if err = graph.Build(); err != nil {
		return err
	}

	// Generate the main method
	err = goprocgen.GenerateMain(
		node.Name(),
		node.ArgNodes,
		node.ContainedNodes, // For now just instantiate all contained nodes
		module,
		fmt.Sprintf("%s/%s", module.Name, procPackage),
		fmt.Sprintf("%s.%s", procPackage, constructorName),
	)
	if err != nil {
		return err
	}

	// Complete workspace generation
	return workspace.Finish()
}
