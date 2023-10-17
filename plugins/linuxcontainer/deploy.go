package linuxcontainer

import (
	"fmt"

	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/core"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/linux"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/linux/procgen"
	"golang.org/x/exp/slog"
)

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
	workspace, err := procgen.NewProcWorkspaceBuilder(dir)
	if err != nil {
		return err
	}

	// Add all processes artifacts to the workspace
	for _, child := range node.ContainedNodes {
		if n, valid := child.(linux.ProvidesProcessArtifacts); valid {
			if err := n.AddProcessArtifacts(workspace); err != nil {
				return err
			}
		}
	}

	// Collect the scripts to run the processes
	graph, err := procgen.NewProcGraphBuilderImpl(workspace, node.Name(), "run.sh")
	if err != nil {
		return err
	}

	for _, child := range node.ContainedNodes {
		if n, valid := child.(linux.InstantiableProcess); valid {
			if err := n.AddProcessInstance(graph); err != nil {
				return err
			}
		}
	}

	for _, child := range node.ArgNodes {
		if err := graph.AddArg(child); err != nil {
			return err
		}

	}

	// TODO: it's possible some metadata / address nodes are residing in this namespace.  They don't
	// get passed in as args, but need to be added to the graph nonetheless

	// Generate the run.sh file
	if err := graph.Build(); err != nil {
		return err
	}

	return workspace.Finish()

}
