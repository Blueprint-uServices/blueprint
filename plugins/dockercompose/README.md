<!-- Code generated by gomarkdoc. DO NOT EDIT -->

# dockerdeployment

```go
import "github.com/blueprint-uservices/blueprint/plugins/dockerdeployment"
```

## Index

- [func AddContainerToDeployment\(spec wiring.WiringSpec, deploymentName, containerName string\)](<#AddContainerToDeployment>)
- [func NewDeployment\(spec wiring.WiringSpec, deploymentName string, containers ...string\) string](<#NewDeployment>)
- [func RegisterAsDefaultBuilder\(\)](<#RegisterAsDefaultBuilder>)
- [type Deployment](<#Deployment>)
  - [func \(deployment \*Deployment\) Accepts\(nodeType any\) bool](<#Deployment.Accepts>)
  - [func \(deployment \*Deployment\) AddEdge\(name string, edge ir.IRNode\) error](<#Deployment.AddEdge>)
  - [func \(deployment \*Deployment\) AddNode\(name string, node ir.IRNode\) error](<#Deployment.AddNode>)
  - [func \(node \*Deployment\) GenerateArtifacts\(dir string\) error](<#Deployment.GenerateArtifacts>)
  - [func \(node \*Deployment\) Name\(\) string](<#Deployment.Name>)
  - [func \(node \*Deployment\) String\(\) string](<#Deployment.String>)
- [type DeploymentNamespace](<#DeploymentNamespace>)


<a name="AddContainerToDeployment"></a>
## func [AddContainerToDeployment](<https://github.com/blueprint-uservices/blueprint/blob/main/plugins/dockerdeployment/wiring.go#L11>)

```go
func AddContainerToDeployment(spec wiring.WiringSpec, deploymentName, containerName string)
```

Adds a child node to an existing container deployment

<a name="NewDeployment"></a>
## func [NewDeployment](<https://github.com/blueprint-uservices/blueprint/blob/main/plugins/dockerdeployment/wiring.go#L17>)

```go
func NewDeployment(spec wiring.WiringSpec, deploymentName string, containers ...string) string
```

Adds a deployment that explicitly instantiates all of the containers provided. The deployment will also implicitly instantiate any of the dependencies of the containers

<a name="RegisterAsDefaultBuilder"></a>
## func [RegisterAsDefaultBuilder](<https://github.com/blueprint-uservices/blueprint/blob/main/plugins/dockerdeployment/defaults.go#L10>)

```go
func RegisterAsDefaultBuilder()
```

to trigger module initialization and register builders

<a name="Deployment"></a>
## type [Deployment](<https://github.com/blueprint-uservices/blueprint/blob/main/plugins/dockerdeployment/ir.go#L8-L15>)

A deployment is a collection of containers

```go
type Deployment struct {
    DeploymentName string
    Nodes          []ir.IRNode
    Edges          []ir.IRNode
    // contains filtered or unexported fields
}
```

<a name="Deployment.Accepts"></a>
### func \(\*Deployment\) [Accepts](<https://github.com/blueprint-uservices/blueprint/blob/main/plugins/dockerdeployment/wiring.go#L38>)

```go
func (deployment *Deployment) Accepts(nodeType any) bool
```

Implements \[wiring.NamespaceHandler\]

<a name="Deployment.AddEdge"></a>
### func \(\*Deployment\) [AddEdge](<https://github.com/blueprint-uservices/blueprint/blob/main/plugins/dockerdeployment/wiring.go#L44>)

```go
func (deployment *Deployment) AddEdge(name string, edge ir.IRNode) error
```

Implements \[wiring.NamespaceHandler\]

<a name="Deployment.AddNode"></a>
### func \(\*Deployment\) [AddNode](<https://github.com/blueprint-uservices/blueprint/blob/main/plugins/dockerdeployment/wiring.go#L50>)

```go
func (deployment *Deployment) AddNode(name string, node ir.IRNode) error
```

Implements \[wiring.NamespaceHandler\]

<a name="Deployment.GenerateArtifacts"></a>
### func \(\*Deployment\) [GenerateArtifacts](<https://github.com/blueprint-uservices/blueprint/blob/main/plugins/dockerdeployment/deploy_dockercompose.go#L50>)

```go
func (node *Deployment) GenerateArtifacts(dir string) error
```

Implements ir.ArtifactGenerator

<a name="Deployment.Name"></a>
### func \(\*Deployment\) [Name](<https://github.com/blueprint-uservices/blueprint/blob/main/plugins/dockerdeployment/ir.go#L18>)

```go
func (node *Deployment) Name() string
```

Implements IRNode

<a name="Deployment.String"></a>
### func \(\*Deployment\) [String](<https://github.com/blueprint-uservices/blueprint/blob/main/plugins/dockerdeployment/ir.go#L23>)

```go
func (node *Deployment) String() string
```

Implements IRNode

<a name="DeploymentNamespace"></a>
## type [DeploymentNamespace](<https://github.com/blueprint-uservices/blueprint/blob/main/plugins/dockerdeployment/wiring.go#L33-L35>)

A \[wiring.NamespaceHandler\] used to build container deployments

```go
type DeploymentNamespace struct {
    *Deployment
}
```

Generated by [gomarkdoc](<https://github.com/princjef/gomarkdoc>)