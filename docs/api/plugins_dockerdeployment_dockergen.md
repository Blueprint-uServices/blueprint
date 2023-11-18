---
title: plugins/dockerdeployment/dockergen
---
# plugins/dockerdeployment/dockergen
```go
package dockergen // import "gitlab.mpi-sws.org/cld/blueprint/plugins/dockerdeployment/dockergen"
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

## TYPES

```go
type DockerComposeFile struct {
	WorkspaceName string
	WorkspaceDir  string
	FileName      string
	FilePath      string
	Instances     map[string]instance // Container instance declarations
```
```go
	// Has unexported fields.
}
```
## func NewDockerComposeFile
```go
func NewDockerComposeFile(workspaceName, workspaceDir, fileName string) *DockerComposeFile
```

## func 
```go
func (d *DockerComposeFile) AddBuildInstance(instanceName string, containerTemplateName string, args ...ir.IRNode) error
```

## func 
```go
func (d *DockerComposeFile) AddImageInstance(instanceName string, image string, args ...ir.IRNode) error
```

## func 
```go
func (d *DockerComposeFile) Generate() error
```

## func 
```go
func (d *DockerComposeFile) ResolveLocalDials() error
```


