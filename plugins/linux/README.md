<!-- Code generated by gomarkdoc. DO NOT EDIT -->

# linux

```go
import "github.com/blueprint-uservices/blueprint/plugins/linux"
```

Package linux defines compiler interfaces for use by plugins that generate and instantiate linux processes. The package does not provide any wiring spec functionality and is not directly used by Blueprint applications; only by other Blueprint plugins.

The noteworthy interfaces are as follows:

- [Process](<#Process>) is an interface for IRNodes that represent linux processes. If an IRNode implements this interface then it will ultimately be instantiated in a namespace that supports linux processes, such as a linux container.
- If a [Process](<#Process>) wants to include code, binaries, or other runnable artifacts, then the IRNode should implement the [ProvidesProcessArtifacts](<#ProvidesProcessArtifacts>) interface.
- If the [Process](<#Process>) can be instantiated \(e.g. by running a command\) then the IRNode should implement the [InstantiableProcess](<#InstantiableProcess>) interface.

Consult the following plugins for examples:

- The [goproc](<https://github.com/Blueprint-uServices/blueprint/tree/main/plugins/goproc>) plugin generates custom process artifacts and provides run commands to run the process \(e.g. the 'go run' command\)
- The [linuxcontainer](<https://github.com/Blueprint-uServices/blueprint/tree/main/plugins/linuxcontainer>) plugin implements a Process namespace that collects together Process nodes and generates run scripts and a Dockerfile if deploying to docker.

## Index

- [func EnvVar\(name string\) string](<#EnvVar>)
- [func FuncName\(name string\) string](<#FuncName>)
- [type InstantiableProcess](<#InstantiableProcess>)
- [type Process](<#Process>)
- [type ProcessWorkspace](<#ProcessWorkspace>)
- [type ProcessWorkspaceInfo](<#ProcessWorkspaceInfo>)
- [type ProvidesProcessArtifacts](<#ProvidesProcessArtifacts>)


<a name="EnvVar"></a>
## func [EnvVar](<https://github.com/blueprint-uservices/blueprint/blob/main/plugins/linux/util.go#L17>)

```go
func EnvVar(name string) string
```

A utility function to deterministically convert a string into a a valid linux environment variable name. This is done by converting all punctuation characters to underscores, and converting alphabetic characters to uppercase \(for convention\), e.g.

```
a.grpc_addr becomes A_GRPC_ADDR.
```

Punctuation is converted to underscores, and alpha are made uppercase.

<a name="FuncName"></a>
## func [FuncName](<https://github.com/blueprint-uservices/blueprint/blob/main/plugins/linux/util.go#L24>)

```go
func FuncName(name string) string
```

A utility function for use when using commands. Converts a string to a compatible command name. Punctuation is converted to underscores, and alpha are made uppercase.

<a name="InstantiableProcess"></a>
## type [InstantiableProcess](<https://github.com/blueprint-uservices/blueprint/blob/main/plugins/linux/ir.go#L53-L57>)

An optional interface for Process IRNodes to implement if the node wants to declare an instance of a process. The process can be started by using standard command\-line commands, or by running custom artifacts that were included by [ProvidesProcessArtifacts](<#ProvidesProcessArtifacts>)

```go
type InstantiableProcess interface {
    // The IRNode is being compiled into the provided target workspace, and should
    // use methods on target to declare how the process should be instantiated.
    AddProcessInstance(target ProcessWorkspace) error
}
```

<a name="Process"></a>
## type [Process](<https://github.com/blueprint-uservices/blueprint/blob/main/plugins/linux/ir.go#L29-L32>)

An IRNode interface that represents a linux process. If an IRNode implements this interface then it enables that IRNode to be instantiated within a process namespace, such as a linux container image.

```go
type Process interface {
    ir.IRNode
    ImplementsLinuxProcess()
}
```

<a name="ProcessWorkspace"></a>
## type [ProcessWorkspace](<https://github.com/blueprint-uservices/blueprint/blob/main/plugins/linux/ir.go#L79-L142>)

[ProcessWorkspace](<#ProcessWorkspace>) receives process artifacts and run commands from [Process](<#Process>) nodes during Blueprint's compilation process.

A [ProcessWorkspace](<#ProcessWorkspace>) instance will be provided to [Process](<#Process>) IRNodes that implement either the [ProvidesProcessArtifacts](<#ProvidesProcessArtifacts>) or [InstantiableProcess](<#InstantiableProcess>) interfaces. The process IRNodes can invoke methods on this workspace in order to add their artifacts to the build output.

After all [Process](<#Process>) instances have added their declarations to the ProcessWorkspace, the ProcessWorkspace will generate a build.sh that invokes any build scripts added by [Process](<#Process>) instances, and a run.sh that will run all of the processes.

The [docker](<https://github.com/Blueprint-uServices/blueprint/tree/main/plugins/docker>) plugin extends the [ProcessWorkspace](<#ProcessWorkspace>) interface to also enable [Process](<#Process>) IRNodes to add custom Dockerfile commands with a function AddDockerfileCommands. To use the docker extensions, the Process IRNode should typecheck the ProcessWorkspace. See the [docker](<https://github.com/Blueprint-uServices/blueprint/tree/main/plugins/docker>) plugin for more details.

```go
type ProcessWorkspace interface {
    ir.BuildContext

    Info() ProcessWorkspaceInfo

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
    DeclareRunCommand(name string, runfunc string, deps ...ir.IRNode) error

    /*
    	Indicates that we have completed building the workspace, and any finalization tasks
    	(e.g. generating build scripts) can run.

    	Only the plugin that created the workspace builder should call this method.
    */
    Finish() error

    ImplementsProcessWorkspace()
}
```

<a name="ProcessWorkspaceInfo"></a>
## type [ProcessWorkspaceInfo](<https://github.com/blueprint-uservices/blueprint/blob/main/plugins/linux/ir.go#L145-L148>)

Metadata about a [ProcessWorkspace](<#ProcessWorkspace>)

```go
type ProcessWorkspaceInfo struct {
    Path   string // fully-qualified path on the filesystem to the workspace
    Target string // the type of workspace being built
}
```

<a name="ProvidesProcessArtifacts"></a>
## type [ProvidesProcessArtifacts](<https://github.com/blueprint-uservices/blueprint/blob/main/plugins/linux/ir.go#L43-L47>)

An optional interface for Process IRNodes to implement if the node needs to generate custom artifacts \(e.g. generate code that then gets compiled/run\) \[target\] provides methods for doing so.

```go
type ProvidesProcessArtifacts interface {
    // The IRNode is being compiled into the provided target workspace, and should
    // use methods on target to add its process artifacts into the workspace.
    AddProcessArtifacts(target ProcessWorkspace) error
}
```

Generated by [gomarkdoc](<https://github.com/princjef/gomarkdoc>)
