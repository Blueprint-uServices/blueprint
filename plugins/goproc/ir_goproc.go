package goproc

import (
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/ir"
)

/*
goproc.Process is a node that represents a runnable Golang process.  It can contain any number of
other golang.Node IRNodes.  When it's compiled, the goproc.Process will generate a go module with
a runnable main method that instantiates and initializes the contained go nodes.  To achieve this,
the golang.Process also collects module dependencies from its contained nodes.

The `GenerateArtifacts` method generates the main method based on the process's contained nodes.

Most of the heavy lifting of code generation is done by the following:
* gogen/workspacebuilder
* gogen/modulebuilder
* gogen/namespacebuilder

*/

var generatedModulePrefix = "blueprint/goproc"

// An IRNode representing a golang process.
// This is Blueprint's main implementation of Golang processes
type Process struct {
	/* The implemented build targets for golang.Process nodes */
	filesystemDeployer /* Can be deployed as a basic go process; implemented in deploy.go */
	linuxDeployer      /* Can be deployed to linux; implemented in deploylinux.go */

	InstanceName   string
	ProcName       string
	ModuleName     string
	ArgNodes       []ir.IRNode
	ContainedNodes []ir.IRNode
}

// A Golang Process Node can either be given the child nodes ahead of time, or they can be added using AddArtifactNode / AddCodeNode
func newGolangProcessNode(name string) *Process {
	node := Process{}
	node.InstanceName = name
	node.ProcName = ir.CleanName(name)
	node.ModuleName = generatedModulePrefix + "/" + node.ProcName
	return &node
}

func (node *Process) Name() string {
	return node.InstanceName
}

func (node *Process) String() string {
	return ir.PrettyPrintNamespace(node.InstanceName, "GolangProcessNode", node.ArgNodes, node.ContainedNodes)
}

func (node *Process) AddArg(argnode ir.IRNode) {
	node.ArgNodes = append(node.ArgNodes, argnode)
}

func (node *Process) AddChild(child ir.IRNode) error {
	node.ContainedNodes = append(node.ContainedNodes, child)
	return nil
}
