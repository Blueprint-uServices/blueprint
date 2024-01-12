package goproc

import (
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/ir"
)

var generatedModulePrefix = "blueprint/goproc"

// An IRNode representing a golang process, which is a collection of application-level golang instances.
type Process struct {
	/* The implemented build targets for golang.Process nodes */
	filesystemDeployer /* Can be deployed as a basic go process; implemented in deploy.go */
	linuxDeployer      /* Can be deployed to linux; implemented in deploylinux.go */

	InstanceName   string
	ProcName       string
	ModuleName     string
	Nodes          []ir.IRNode
	Edges          []ir.IRNode
	metricProvider ir.IRNode
	logger         ir.IRNode
}

func newGolangProcessNode(name string) *Process {
	proc := Process{
		InstanceName: name,
		ProcName:     ir.CleanName(name),
	}
	proc.ModuleName = generatedModulePrefix + "/" + proc.ProcName
	return &proc
}

// Implements ir.IRNode
func (proc *Process) Name() string {
	return proc.InstanceName
}

// Implements ir.IRNode
func (proc *Process) String() string {
	return ir.PrettyPrintNamespace(proc.InstanceName, "GolangProcessNode", proc.Edges, proc.Nodes)
}
