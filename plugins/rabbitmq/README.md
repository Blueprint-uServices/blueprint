<!-- Code generated by gomarkdoc. DO NOT EDIT -->

# rabbitmq

```go
import "github.com/blueprint-uservices/blueprint/plugins/rabbitmq"
```

Package rabbitmq provides a plugin to generate and include a rabbitmq instance in a Blueprint application.

The package provides a built\-in rabbitmq container that provides the server\-side implementation and a go\-client for connecting to the client.

The applications must use a backend.Queue \(runtime/core/backend\) as the interface in the workflow.

## Index

- [func Container\(spec wiring.WiringSpec, name string, queue\_name string\) string](<#Container>)
- [type RabbitmqContainer](<#RabbitmqContainer>)
  - [func \(n \*RabbitmqContainer\) AddContainerArtifacts\(target docker.ContainerWorkspace\) error](<#RabbitmqContainer.AddContainerArtifacts>)
  - [func \(n \*RabbitmqContainer\) AddContainerInstance\(target docker.ContainerWorkspace\) error](<#RabbitmqContainer.AddContainerInstance>)
  - [func \(n \*RabbitmqContainer\) GenerateArtifacts\(outdir string\) error](<#RabbitmqContainer.GenerateArtifacts>)
  - [func \(n \*RabbitmqContainer\) GetInterface\(ctx ir.BuildContext\) \(service.ServiceInterface, error\)](<#RabbitmqContainer.GetInterface>)
  - [func \(n \*RabbitmqContainer\) Name\(\) string](<#RabbitmqContainer.Name>)
  - [func \(n \*RabbitmqContainer\) String\(\) string](<#RabbitmqContainer.String>)
- [type RabbitmqGoClient](<#RabbitmqGoClient>)
  - [func \(n \*RabbitmqGoClient\) AddInstantiation\(builder golang.NamespaceBuilder\) error](<#RabbitmqGoClient.AddInstantiation>)
  - [func \(n \*RabbitmqGoClient\) AddInterfaces\(builder golang.ModuleBuilder\) error](<#RabbitmqGoClient.AddInterfaces>)
  - [func \(n \*RabbitmqGoClient\) AddToWorkspace\(builder golang.WorkspaceBuilder\) error](<#RabbitmqGoClient.AddToWorkspace>)
  - [func \(n \*RabbitmqGoClient\) GetInterface\(ctx ir.BuildContext\) \(service.ServiceInterface, error\)](<#RabbitmqGoClient.GetInterface>)
  - [func \(n \*RabbitmqGoClient\) ImplementsGolangNode\(\)](<#RabbitmqGoClient.ImplementsGolangNode>)
  - [func \(n \*RabbitmqGoClient\) ImplementsGolangService\(\)](<#RabbitmqGoClient.ImplementsGolangService>)
  - [func \(n \*RabbitmqGoClient\) Name\(\) string](<#RabbitmqGoClient.Name>)
  - [func \(n \*RabbitmqGoClient\) String\(\) string](<#RabbitmqGoClient.String>)
- [type RabbitmqInterface](<#RabbitmqInterface>)
  - [func \(r \*RabbitmqInterface\) GetMethods\(\) \[\]service.Method](<#RabbitmqInterface.GetMethods>)
  - [func \(r \*RabbitmqInterface\) GetName\(\) string](<#RabbitmqInterface.GetName>)


<a name="Container"></a>
## func [Container](<https://github.com/Blueprint-uServices/blueprint/blob/main/plugins/rabbitmq/wiring.go#L19>)

```go
func Container(spec wiring.WiringSpec, name string, queue_name string) string
```

Container generate the IRNodes for a mysql server docker container that uses the latest mysql/mysql image and the clients needed by the generated application to communicate with the server.

<a name="RabbitmqContainer"></a>
## type [RabbitmqContainer](<https://github.com/Blueprint-uServices/blueprint/blob/main/plugins/rabbitmq/ir_container.go#L14-L21>)

Blueprint IR Node that represents the server side docker container

```go
type RabbitmqContainer struct {
    docker.Container
    backend.Queue

    InstanceName string
    BindAddr     *address.BindConfig
    Iface        *goparser.ParsedInterface
}
```

<a name="RabbitmqContainer.AddContainerArtifacts"></a>
### func \(\*RabbitmqContainer\) [AddContainerArtifacts](<https://github.com/Blueprint-uServices/blueprint/blob/main/plugins/rabbitmq/ir_container.go#L81>)

```go
func (n *RabbitmqContainer) AddContainerArtifacts(target docker.ContainerWorkspace) error
```



<a name="RabbitmqContainer.AddContainerInstance"></a>
### func \(\*RabbitmqContainer\) [AddContainerInstance](<https://github.com/Blueprint-uServices/blueprint/blob/main/plugins/rabbitmq/ir_container.go#L85>)

```go
func (n *RabbitmqContainer) AddContainerInstance(target docker.ContainerWorkspace) error
```



<a name="RabbitmqContainer.GenerateArtifacts"></a>
### func \(\*RabbitmqContainer\) [GenerateArtifacts](<https://github.com/Blueprint-uServices/blueprint/blob/main/plugins/rabbitmq/ir_container.go#L77>)

```go
func (n *RabbitmqContainer) GenerateArtifacts(outdir string) error
```



<a name="RabbitmqContainer.GetInterface"></a>
### func \(\*RabbitmqContainer\) [GetInterface](<https://github.com/Blueprint-uServices/blueprint/blob/main/plugins/rabbitmq/ir_container.go#L72>)

```go
func (n *RabbitmqContainer) GetInterface(ctx ir.BuildContext) (service.ServiceInterface, error)
```



<a name="RabbitmqContainer.Name"></a>
### func \(\*RabbitmqContainer\) [Name](<https://github.com/Blueprint-uServices/blueprint/blob/main/plugins/rabbitmq/ir_container.go#L68>)

```go
func (n *RabbitmqContainer) Name() string
```



<a name="RabbitmqContainer.String"></a>
### func \(\*RabbitmqContainer\) [String](<https://github.com/Blueprint-uServices/blueprint/blob/main/plugins/rabbitmq/ir_container.go#L64>)

```go
func (n *RabbitmqContainer) String() string
```



<a name="RabbitmqGoClient"></a>
## type [RabbitmqGoClient](<https://github.com/Blueprint-uServices/blueprint/blob/main/plugins/rabbitmq/ir_client.go#L18-L26>)

Blueprint IR Node that represents the generated client for the rabbitmq container

```go
type RabbitmqGoClient struct {
    golang.Service
    backend.Queue
    InstanceName string
    QueueName    *ir.IRValue
    Addr         *address.DialConfig
    Iface        *goparser.ParsedInterface
    Constructor  *gocode.Constructor
}
```

<a name="RabbitmqGoClient.AddInstantiation"></a>
### func \(\*RabbitmqGoClient\) [AddInstantiation](<https://github.com/Blueprint-uServices/blueprint/blob/main/plugins/rabbitmq/ir_client.go#L79>)

```go
func (n *RabbitmqGoClient) AddInstantiation(builder golang.NamespaceBuilder) error
```



<a name="RabbitmqGoClient.AddInterfaces"></a>
### func \(\*RabbitmqGoClient\) [AddInterfaces](<https://github.com/Blueprint-uServices/blueprint/blob/main/plugins/rabbitmq/ir_client.go#L75>)

```go
func (n *RabbitmqGoClient) AddInterfaces(builder golang.ModuleBuilder) error
```



<a name="RabbitmqGoClient.AddToWorkspace"></a>
### func \(\*RabbitmqGoClient\) [AddToWorkspace](<https://github.com/Blueprint-uServices/blueprint/blob/main/plugins/rabbitmq/ir_client.go#L71>)

```go
func (n *RabbitmqGoClient) AddToWorkspace(builder golang.WorkspaceBuilder) error
```



<a name="RabbitmqGoClient.GetInterface"></a>
### func \(\*RabbitmqGoClient\) [GetInterface](<https://github.com/Blueprint-uServices/blueprint/blob/main/plugins/rabbitmq/ir_client.go#L67>)

```go
func (n *RabbitmqGoClient) GetInterface(ctx ir.BuildContext) (service.ServiceInterface, error)
```



<a name="RabbitmqGoClient.ImplementsGolangNode"></a>
### func \(\*RabbitmqGoClient\) [ImplementsGolangNode](<https://github.com/Blueprint-uServices/blueprint/blob/main/plugins/rabbitmq/ir_client.go#L88>)

```go
func (n *RabbitmqGoClient) ImplementsGolangNode()
```



<a name="RabbitmqGoClient.ImplementsGolangService"></a>
### func \(\*RabbitmqGoClient\) [ImplementsGolangService](<https://github.com/Blueprint-uServices/blueprint/blob/main/plugins/rabbitmq/ir_client.go#L89>)

```go
func (n *RabbitmqGoClient) ImplementsGolangService()
```



<a name="RabbitmqGoClient.Name"></a>
### func \(\*RabbitmqGoClient\) [Name](<https://github.com/Blueprint-uServices/blueprint/blob/main/plugins/rabbitmq/ir_client.go#L44>)

```go
func (n *RabbitmqGoClient) Name() string
```



<a name="RabbitmqGoClient.String"></a>
### func \(\*RabbitmqGoClient\) [String](<https://github.com/Blueprint-uServices/blueprint/blob/main/plugins/rabbitmq/ir_client.go#L40>)

```go
func (n *RabbitmqGoClient) String() string
```



<a name="RabbitmqInterface"></a>
## type [RabbitmqInterface](<https://github.com/Blueprint-uServices/blueprint/blob/main/plugins/rabbitmq/ir_container.go#L24-L27>)

RabbitMQ interface exposed by the docker container.

```go
type RabbitmqInterface struct {
    service.ServiceInterface
    Wrapped service.ServiceInterface
}
```

<a name="RabbitmqInterface.GetMethods"></a>
### func \(\*RabbitmqInterface\) [GetMethods](<https://github.com/Blueprint-uServices/blueprint/blob/main/plugins/rabbitmq/ir_container.go#L33>)

```go
func (r *RabbitmqInterface) GetMethods() []service.Method
```



<a name="RabbitmqInterface.GetName"></a>
### func \(\*RabbitmqInterface\) [GetName](<https://github.com/Blueprint-uServices/blueprint/blob/main/plugins/rabbitmq/ir_container.go#L29>)

```go
func (r *RabbitmqInterface) GetName() string
```



Generated by [gomarkdoc](<https://github.com/princjef/gomarkdoc>)