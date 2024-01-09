package linuxcontainer

import (
	"fmt"
	"path/filepath"

	"github.com/blueprint-uservices/blueprint/blueprint/pkg/blueprint/ioutil"
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/ir"
	"github.com/blueprint-uservices/blueprint/plugins/linux"
	"github.com/blueprint-uservices/blueprint/plugins/linuxcontainer/linuxgen"
	"golang.org/x/exp/slog"
)

/*
A collection of processes can, in their simplest form, just be output
to a directory on the local filesystem.
*/

type (
	/*
		The default linux container deployer doesn't assume anything about the target environment,
		nor the existence of a container manager.  The deployer simply packages all process
		artifacts together along with a linux build script and a linux run script, into
		an output directory.
		It is assumed that the user will manually invoke the build script or pre-install dependencies
		and manually call the run script.
	*/
	filesystemDeployer interface {
		ir.ArtifactGenerator
	}

	/*
	   The base implementation of the linux.ProcessWorkspace defined in linux/ir.go

	   This workspace performs the basic actions that are (presumed to be) common
	   to all process workspaces:
	    (a) gather each process's artifacts into process subdirectories
	    (b) allow each process to declare a run command
	    (c) allow each process to provide a build file
	    (d) generate a root build.sh that invokes each process's build file
	    (e) generate a root run.sh that invokes each process's run command

	   Note that the Docker process workspace extends this workspace to enable
	   processes to additionally provide Dockerfile build commands in lieu of
	   a build.sh script
	*/
	filesystemWorkspace struct {
		ir.VisitTrackerImpl

		info linux.ProcessWorkspaceInfo

		ProcDirs map[string]string // map from proc name to directory

		Build *linuxgen.BuildScript
		Run   *linuxgen.RunScript
	}
)

/*
Implements ir.ArtifactGenerator

This is the starting point for generating process workspace artifacts.

Collects process artifacts into a directory on the local filesystem and
generates a build.sh and run.sh script.

The output processes will be runnable in the local environment.
*/
func (node *Container) GenerateArtifacts(dir string) error {
	slog.Info(fmt.Sprintf("Collecting process artifacts for %s in %s", node.Name(), dir))
	workspace := NewBasicWorkspace(node.Name(), dir)
	return node.generateArtifacts(workspace)
}

/*
The basic build process for any container of processes.

Deployment targets like Docker will extend the linuxgen.basicWorkspace
to offer extra platform-specific commands.

Process nodes that implement AddProcessArtifacts and AddProcessInstance
can typecheck the workspace to utilize those platform-specific commands.
*/
func (node *Container) generateArtifacts(workspace linux.ProcessWorkspace) error {
	// Add all processes artifacts to the workspace
	for _, child := range node.Nodes {
		if n, valid := child.(linux.ProvidesProcessArtifacts); valid {
			if err := n.AddProcessArtifacts(workspace); err != nil {
				return err
			}
		}
	}

	// Collect the scripts to run the processes
	for _, child := range node.Nodes {
		if n, valid := child.(linux.InstantiableProcess); valid {
			if err := n.AddProcessInstance(workspace); err != nil {
				return err
			}
		}
	}

	// TODO: it's possible some metadata / address nodes are residing in this namespace.  They don't
	// get passed in as args, but need to be added to the namespace nonetheless
	return workspace.Finish()
}

// Creates a BasicWorkspace, which is the simplest process workspace
// that can write processes to an output directory
func NewBasicWorkspace(name string, dir string) *filesystemWorkspace {
	return &filesystemWorkspace{
		info: linux.ProcessWorkspaceInfo{
			Path:   filepath.Clean(dir),
			Target: "basic",
		},
		Build:    linuxgen.NewBuildScript(dir, "build.sh"),
		Run:      linuxgen.NewRunScript(name, dir, "run.sh"),
		ProcDirs: make(map[string]string),
	}
}

// Implements linux.ProcessWorkspace
func (workspace *filesystemWorkspace) Info() linux.ProcessWorkspaceInfo {
	return workspace.info
}

// Implements linux.ProcessWorkspace
//
// Creates a subdirectory for a process to output its artifacts.
// Saves the metadata about the process
func (ws *filesystemWorkspace) CreateProcessDir(name string) (string, error) {
	path, err := ioutil.CreateNodeDir(ws.info.Path, name)
	ws.ProcDirs[ir.CleanName(name)] = path
	return path, err
}

// Implements linux.ProcessWorkspace
//
// Adds a build script provided by a process
func (ws *filesystemWorkspace) AddBuildScript(path string) error {
	return ws.Build.Add(path)
}

// Implements linux.ProcessWorkspace
//
// Adds a command to the run.sh file for running the specified process node
func (ws *filesystemWorkspace) DeclareRunCommand(name string, runfunc string, deps ...ir.IRNode) error {
	// Generate the runfunc
	runfunc_impl, err := linuxgen.GenerateRunFunc(name, runfunc, deps...)
	ws.Run.Add(name, runfunc_impl, deps...)
	return err
}

// Implements linux.ProcessWorkspace
//
// # Creates a build.sh and a run.sh file in the root of the proc workspace
//
// When invoked, the build.sh file will sequentially invoke any
// build scripts that were provided by processes in the workspace.
//
// The build.sh will typically be invoked by e.g. a Dockerfile
func (ws *filesystemWorkspace) Finish() error {
	// Generate the build.sh
	if err := ws.Build.GenerateBuildScript(); err != nil {
		return err
	}

	// Generate the run.sh
	return ws.Run.GenerateRunScript()
}

func (ws *filesystemWorkspace) ImplementsBuildContext()     {}
func (ws *filesystemWorkspace) ImplementsProcessWorkspace() {}
