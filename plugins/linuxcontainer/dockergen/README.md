<!-- Code generated by gomarkdoc. DO NOT EDIT -->

# dockergen

```go
import "github.com/blueprint-uservices/blueprint/plugins/linuxcontainer/dockergen"
```

## Index

- [type Dockerfile](<#Dockerfile>)
  - [func NewDockerfile\(workspaceName, workspaceDir string\) \*Dockerfile](<#NewDockerfile>)
  - [func \(d \*Dockerfile\) AddCustomCommands\(procName string, commands string\)](<#Dockerfile.AddCustomCommands>)
  - [func \(d \*Dockerfile\) Generate\(procDirs map\[string\]string\) error](<#Dockerfile.Generate>)


<a name="Dockerfile"></a>
## type [Dockerfile](<https://github.com/blueprint-uservices/blueprint/blob/main/plugins/linuxcontainer/dockergen/dockerfile_.go#L11-L17>)



```go
type Dockerfile struct {
    WorkspaceName string
    WorkspaceDir  string
    FilePath      string
    CustomProcs   map[string]string
    DefaultProcs  map[string]string
}
```

<a name="NewDockerfile"></a>
### func [NewDockerfile](<https://github.com/blueprint-uservices/blueprint/blob/main/plugins/linuxcontainer/dockergen/dockerfile_.go#L19>)

```go
func NewDockerfile(workspaceName, workspaceDir string) *Dockerfile
```



<a name="Dockerfile.AddCustomCommands"></a>
### func \(\*Dockerfile\) [AddCustomCommands](<https://github.com/blueprint-uservices/blueprint/blob/main/plugins/linuxcontainer/dockergen/dockerfile_.go#L29>)

```go
func (d *Dockerfile) AddCustomCommands(procName string, commands string)
```



<a name="Dockerfile.Generate"></a>
### func \(\*Dockerfile\) [Generate](<https://github.com/blueprint-uservices/blueprint/blob/main/plugins/linuxcontainer/dockergen/dockerfile_.go#L33>)

```go
func (d *Dockerfile) Generate(procDirs map[string]string) error
```



Generated by [gomarkdoc](<https://github.com/princjef/gomarkdoc>)
