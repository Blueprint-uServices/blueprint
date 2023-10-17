package linuxcontainer

import (
	"fmt"

	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/blueprint"
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/core"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/linux"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/linuxcontainer/workspace"
	"golang.org/x/exp/slog"
)

func init() {
	RegisterBuilders()
}

// to trigger module initialization and register builders
func RegisterBuilders() {
	blueprint.RegisterDefaultNamespace[linux.Process]("linuxcontainer", buildDefaultProcessWorkspace)
}

/*
The default linux container deployer doesn't assume anything about the target environment,
nor the existence of a container manager.  The deployer simply packages all process
artifacts together along with a linux build script and a linux run script.
It is assumed that the user will manually invoke the build script or pre-install dependencies
and manually call the run script.
*/

type BasicLinuxContainer interface {
	core.ArtifactGenerator
}

/*
Collects process artifacts into a directory on the local filesystem and
generates a build.sh and run.sh script.

The output processes will be runnable in the local environment.
*/
func (node *Container) GenerateArtifacts(dir string) error {
	slog.Info(fmt.Sprintf("Collecting process artifacts for %s in %s", node.Name(), dir))
	workspace := workspace.NewBasicWorkspace(node.Name(), dir)
	return node.generateArtifacts(workspace)
}

/*
The basic build process for any container of processes.

Deployment targets like Docker will extend the linuxgen.BasicWorkspace
to offer extra platform-specific commands.

Process nodes that implement AddProcessArtifacts and AddProcessInstance
can typecheck the workspace to utilize those platform-specific commands.
*/
func (node *Container) generateArtifacts(workspace linux.ProcessWorkspace) error {
	// Add all processes artifacts to the workspace
	for _, child := range node.ContainedNodes {
		if n, valid := child.(linux.ProvidesProcessArtifacts); valid {
			if err := n.AddProcessArtifacts(workspace); err != nil {
				return err
			}
		}
	}

	// Collect the scripts to run the processes
	for _, child := range node.ContainedNodes {
		if n, valid := child.(linux.InstantiableProcess); valid {
			if err := n.AddProcessInstance(workspace); err != nil {
				return err
			}
		}
	}

	// // Tell the workspace the nodes it should expect as args
	// for _, child := range node.ArgNodes {
	// 	workspace.AddArg(child)
	// }

	// TODO: it's possible some metadata / address nodes are residing in this namespace.  They don't
	// get passed in as args, but need to be added to the graph nonetheless
	return workspace.Finish()
}

func buildDefaultProcessWorkspace(outputDir string, nodes []blueprint.IRNode) error {
	ctr := newLinuxContainerNode("default")
	ctr.ContainedNodes = nodes
	return ctr.GenerateArtifacts(outputDir)
}
