---
title: plugins/linuxcontainer/dockergen
---
# plugins/linuxcontainer/dockergen
```go
package dockergen // import "gitlab.mpi-sws.org/cld/blueprint/plugins/linuxcontainer/dockergen"
```

## TYPES

```go
type Dockerfile struct {
	WorkspaceName string
	WorkspaceDir  string
	FilePath      string
	CustomProcs   map[string]string
	DefaultProcs  map[string]string
}
```
## func NewDockerfile
```go
func NewDockerfile(workspaceName, workspaceDir string) *Dockerfile
```

## func 
```go
func (d *Dockerfile) AddCustomCommands(procName string, commands string)
```

## func 
```go
func (d *Dockerfile) Generate(procDirs map[string]string) error
```


