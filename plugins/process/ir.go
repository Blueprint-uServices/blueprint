package process

import (
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/blueprint"
)

/*
process.Node is the base interface for any process

To support process artifact generation, the following IR interfaces are provided.
- process.ProvidesProcessArtifacts is for process nodes that collect files or
  run commands to collect together runnable process artifacts
- process.InstantiableProcess is for process nodes that can be run via a command

Most processes will implement both IR interfaces, but some might not need any
artifacts
*/

// This Node represents a process
type (

	/*
		The base IRNode interface for processes
	*/
	Node interface {
		blueprint.IRNode
		ImplementsProcessNode()
	}
)

type (

	/*
		For process nodes that want to provide code or other artifacts for their process.
		Methods on the ProcWorkspaceBuilder argument are used for collecting the artifacts
	*/
	ProvidesProcessArtifacts interface {
		AddProcessArtifacts(ProcWorkspaceBuilder) error
	}

	/*
		For process nodes that can be instantiated.
		Methods on the GraphBuilder argument are used for declaring commands to start processes
	*/
	InstantiableProcess interface {
		AddProcessInstance(GraphBuilder) error
	}
)

type (
	ProcWorkspaceInfo struct {
		Path string // fully-qualified path on the filesystem to the workspace
	}

	/*
		A workspace just contains the artifacts for a number of different processes.

		Process nodes can provide their artifacts using the methods on this interface
	*/
	ProcWorkspaceBuilder interface {
		blueprint.BuildContext

		Info() ProcWorkspaceInfo

		/*
			Creates a subdirectory in the workspace dir for a process node to collect
			its artifacts.
			Returns a fully qualified path on the local filesystem where artifacts will be
			collected.
		*/
		CreateProcessDir(name string) (string, error)

		/*
			Provides a build script that may be invoked to further collect or build process
			dependencies.
			For containers, the build script will be invoked from within the container.
			The script will be invoked from workspace's root directory.
		*/
		AddBuildScript(path string) error
	}

	GraphInfo struct {
		Workspace WorkspaceInfo
		FileName  string // Name of the file within the package
		FilePath  string // Fully-qualified path to the file on the local filesystem
		FuncName  string // Name of the function that builds the graph
	}

	/*
		The GraphBuilder accumulates the commands needed to start processes.
		It then creates the script needed to start all processes, that is typically
		used as the run command for a container image.
	*/
	GraphBuilder interface {
		blueprint.BuildContext

		Info() GraphInfo

		DeclareCommand(name string, cmd string, args []blueprint.IRNode) error

		// Generates a script somewhere that, when you run it, runs the graph??
		Build()
	}
)
