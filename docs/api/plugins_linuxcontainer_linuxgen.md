---
title: plugins/linuxcontainer/linuxgen
---
# plugins/linuxcontainer/linuxgen
```go
package linuxgen // import "gitlab.mpi-sws.org/cld/blueprint/plugins/linuxcontainer/linuxgen"
```

## FUNCTIONS

## func ExecuteTemplate
```go
func ExecuteTemplate(name string, body string, args any) (string, error)
```
## func ExecuteTemplateToFile
```go
func ExecuteTemplateToFile(name string, body string, args any, filename string) error
```
## func GenerateRunFunc
```go
func GenerateRunFunc(name string, runfunc string, deps ...ir.IRNode) (string, error)
```

## TYPES

```go
type BuildScript struct {
	WorkspaceDir string
	FileName     string
	FilePath     string
	Scripts      map[string]*scriptInfo
}
```
## func NewBuildScript
```go
func NewBuildScript(workspaceDir, fileName string) *BuildScript
```
Creates a new build.sh that will invoke multiple build scripts of processes
in subdirectories of the workspace

## func 
```go
func (d *BuildScript) Add(filePath string) error
```
Adds a process's build script to the workspace's build.sh

filePath should be a fully qualified path to a build script that resides
within a subdirectory of the workspace

Returns an error if the script resides outside of the workspace

## func 
```go
func (d *BuildScript) GenerateBuildScript() error
```

```go
type RunScript struct {
	WorkspaceName string
	WorkspaceDir  string
	FileName      string
	FilePath      string
	RunFuncs      map[string]string    // Function bodies provided by processes
	AllNodes      map[string]ir.IRNode // All nodes seen by this run script
	Args          map[string]ir.IRNode // Arguments that will be set in calling the environment
}
```
## func NewRunScript
```go
func NewRunScript(workspaceName, workspaceDir, fileName string) *RunScript
```
Creates a new run.sh that will check environment variables are set and
invokes the run scripts of the processes within the workspace

## func 
```go
func (run *RunScript) Add(procName, runfunc string, deps ...ir.IRNode)
```

## func 
```go
func (run *RunScript) GenerateRunScript() error
```

## func 
```go
func (run *RunScript) Require(node ir.IRNode)
```
Indicate that the specified node is required within this namespace; either
it is built by its own runfunc, or it must be provided as argument.

We use this so that the generated run.sh knows which environment variables
will be needed or used by the processes it runs.


