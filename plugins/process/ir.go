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
		Methods on the ProcGraphBuilder argument are used for declaring commands to start processes
	*/
	InstantiableProcess interface {
		AddProcessInstance(ProcGraphBuilder) error
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
			This will typically be invoked from e.g. within a Container (e.g a Dockerfile),
			rather than on the host machine environment.

			path must refer to a script that resides within a process dir in this workspace;
			if not an error will be returned.

			When it does get invoked, the script will be invoked from the process dir in
			which it resides.
		*/
		AddBuildScript(path string) error

		/*
			Indicates that we have completed building the workspace, and any finalization tasks
			(e.g. generating build scripts) can run.

			Only the plugin that created the workspace builder should call this method.
		*/
		Finish() error
	}

	ProcGraphInfo struct {
		Workspace ProcWorkspaceInfo
		Name      string // Name of the graph
		FileName  string // Name of the file
		FileDir   string // Dir within the workspace containing the file
		FilePath  string // Path to the file within the workspace
	}

	/*
		The ProcGraphBuilder accumulates the commands needed to start processes.
		It then creates the script needed to start all processes, that is typically
		used as the run command for a container image.
	*/
	ProcGraphBuilder interface {
		blueprint.BuildContext

		Info() ProcGraphInfo

		/*
			A plugin can provide the shell command(s) to run its process.

			Name is just the name of the IRNode representing the process.  Other IRNodes
			that want to instantiate the process will use this name to look it up.

			If the process has dependencies on other IRNodes, they can be provided with
			the deps argument.  The generated code will ensure that the dependencies
			get instantiated first before the runfunc is executed.

			runfunc is a bash function declaration for running the process.
			The runfunc should adhere to the following:
			 - should be defined with syntax like function my_func() { ... }
			 - for any dependencies (config values, addresses, pids, etc.) they can be
			   accessed from environment variable with the corresponding name.  e.g.
			   a.grpc.addr will be in A_GRPC_ADDR.  The mapping from node name to
			   env variable name is implemented by process.EnvVar(name)
			 - the function must set an environment variable for Name with the result
			   of the runfunc.  Typically, this means setting the PID of a started process
			   e.g. MY_GOLANG_PROC=$!
			 - the function must return a return code that will be checked
			 - when it is invoked, the runfunc will be invoked from the root of the
			   proc workspace
			 - the runfunc will be renamed to prevent name clashes between IRNodes
		*/
		DeclareRunCommand(name string, runfunc string, deps ...blueprint.IRNode) error
	}
)
