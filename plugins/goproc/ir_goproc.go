package goproc

import (
	"fmt"
	"path/filepath"
	"strings"

	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/blueprint"
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/core/irutil"
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/core/process"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/golang/gogen"
)

/*
This file contains the implementation of the golang.Process IRNode.

The `GenerateArtifacts` method generates the main method based on the process's contained nodes.

Most of the heavy lifting of code generation is done by the following:
* gogen/workspacebuilder
* gogen/modulebuilder
* gogen/graphbuilder

*/

var generatedModulePrefix = "gitlab.mpi-sws.org/cld/blueprint/plugins/golang/process"

// An IRNode representing a golang process.
// This is Blueprint's main implementation of Golang processes
type Process struct {
	blueprint.IRNode
	process.ProcessNode
	process.ArtifactGenerator

	InstanceName   string
	ArgNodes       []blueprint.IRNode
	ContainedNodes []blueprint.IRNode
}

// A Golang Process Node can either be given the child nodes ahead of time, or they can be added using AddArtifactNode / AddCodeNode
func newGolangProcessNode(name string) *Process {
	node := Process{}
	node.InstanceName = name
	return &node
}

func (node *Process) Name() string {
	return node.InstanceName
}

func (node *Process) String() string {
	var b strings.Builder
	b.WriteString(node.InstanceName)
	b.WriteString(" = GolangProcessNode(")
	var args []string
	for _, arg := range node.ArgNodes {
		args = append(args, arg.Name())
	}
	b.WriteString(strings.Join(args, ", "))
	b.WriteString(") {\n")
	var children []string
	for _, child := range node.ContainedNodes {
		children = append(children, child.String())
	}
	b.WriteString(blueprint.Indent(strings.Join(children, "\n"), 2))
	b.WriteString("\n}")
	return b.String()
}

func (node *Process) AddArg(argnode blueprint.IRNode) {
	node.ArgNodes = append(node.ArgNodes, argnode)
}

func (node *Process) AddChild(child blueprint.IRNode) error {
	node.ContainedNodes = append(node.ContainedNodes, child)
	return nil
}

func (node *Process) GenerateArtifacts(outputDir string) error {
	if gogen.IsDir(outputDir) {
		return fmt.Errorf("cannot built to %s, directory already exists", outputDir)
	}
	err := gogen.CheckDir(outputDir, true)
	if err != nil {
		return fmt.Errorf("unable to create %s for process %s due to %s", outputDir, node.Name(), err.Error())
	}

	// TODO: might end up building multiple times which is OK, so need a check here that we haven't already built this artifact, even if it was by a different (but identical) node
	cleanName := irutil.Clean(node.Name())
	workspaceDir := filepath.Join(outputDir, cleanName)
	workspace, err := gogen.NewWorkspaceBuilder(workspaceDir)
	if err != nil {
		return err
	}

	moduleName := generatedModulePrefix + "/" + cleanName
	module, err := gogen.NewModuleBuilder(workspace, cleanName, moduleName)
	if err != nil {
		return err
	}

	fileName := strings.ToLower(cleanName) + ".go"
	packagePath := "goproc"
	constructorName := "New" + strings.ToTitle(cleanName)

	code, err := gogen.NewGraphBuilder(module, fileName, packagePath, constructorName)
	if err != nil {
		return err
	}

	for _, node := range node.ContainedNodes {
		err := workspace.Visit(node)
		if err != nil {
			return err
		}
	}

	for _, node := range node.ContainedNodes {
		err := module.Visit(node)
		if err != nil {
			return err
		}
	}

	for _, node := range node.ContainedNodes {
		err := code.Visit(node)
		if err != nil {
			return err
		}
	}

	err = code.Build()
	if err != nil {
		return err
	}

	err = module.Finish()
	if err != nil {
		return err
	}

	err = workspace.Finish()
	if err != nil {
		return err
	}

	return nil
}
