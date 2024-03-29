<!-- Code generated by gomarkdoc. DO NOT EDIT -->

# clientpool

```go
import "github.com/blueprint-uservices/blueprint/plugins/clientpool"
```

Package clientpool is a plugin that wraps the client side of a service to use a pool of N clients, disallowing callers from making more than N concurrent oustanding calls each to the service.

### Wiring Spec Usage

To use the clientpool plugin in your wiring spec, simply apply it to an application\-level service instance:

```
clientpool.Create(spec, "my_service", 10)
```

### Description

When applied, the clientpool plugin instantiates N instances of clients to a service, and callers have exclusive access to a client when making a call. This effectively limits the caller\-side to only having N outstanding calls at a time, with any extra calls blocking until a previous call completes and a client becomes available. By contrast, the default Blueprint behavior is for all callers to share a single client that allows an unlimited number of concurrent calls.

After applying the clientpool plugin to a service, you can continue to apply application\-level modifiers to the service.

### Artifacts Generated

During compilation, the clientpool plugin will generate a client\-side wrapper class. The plugin also utilizes some code in the [runtime/plugins/clientpool](<https://github.com/Blueprint-uServices/blueprint/tree/main/runtime/plugins/clientpool>) package.

## Index

- [func Create\(spec wiring.WiringSpec, serviceName string, numClients int\)](<#Create>)
- [type ClientPool](<#ClientPool>)
  - [func \(pool \*ClientPool\) Accepts\(nodeType any\) bool](<#ClientPool.Accepts>)
  - [func \(pool \*ClientPool\) AddEdge\(name string, edge ir.IRNode\) error](<#ClientPool.AddEdge>)
  - [func \(pool \*ClientPool\) AddInstantiation\(builder golang.NamespaceBuilder\) error](<#ClientPool.AddInstantiation>)
  - [func \(pool \*ClientPool\) AddInterfaces\(module golang.ModuleBuilder\) error](<#ClientPool.AddInterfaces>)
  - [func \(pool \*ClientPool\) AddNode\(name string, node ir.IRNode\) error](<#ClientPool.AddNode>)
  - [func \(pool \*ClientPool\) GenerateFuncs\(module golang.ModuleBuilder\) error](<#ClientPool.GenerateFuncs>)
  - [func \(pool \*ClientPool\) GetInterface\(ctx ir.BuildContext\) \(service.ServiceInterface, error\)](<#ClientPool.GetInterface>)
  - [func \(pool \*ClientPool\) Name\(\) string](<#ClientPool.Name>)
  - [func \(pool \*ClientPool\) String\(\) string](<#ClientPool.String>)


<a name="Create"></a>
## func [Create](<https://github.com/blueprint-uservices/blueprint/blob/main/plugins/clientpool/wiring.go#L48>)

```go
func Create(spec wiring.WiringSpec, serviceName string, numClients int)
```

Create can be used by wiring specs to add a clientpool to the client side of a service.

This will modify the client\-side of serviceName so that all calls are made using a pool of numClients clients.

At runtime, clients to serviceName will only be able to have up to numClients concurrent calls outstanding, before subsequent calls block.

serviceName must be an application\-level service instance, e.g. clientpool must be applied to the service before deploying the service over RPC or to a process.

After calling [Create](<#Create>) you can continue to apply application\-level modifiers to serviceName.

<a name="ClientPool"></a>
## type [ClientPool](<https://github.com/blueprint-uservices/blueprint/blob/main/plugins/clientpool/ir.go#L18-L27>)

Blueprint IR node representing a ClientPool that uses \[N\] instances of \[Client\]

```go
type ClientPool struct {
    golang.Service
    golang.GeneratesFuncs

    PoolName string
    N        int
    Client   golang.Service
    Edges    []ir.IRNode
    Nodes    []ir.IRNode
}
```

<a name="ClientPool.Accepts"></a>
### func \(\*ClientPool\) [Accepts](<https://github.com/blueprint-uservices/blueprint/blob/main/plugins/clientpool/wiring.go#L78>)

```go
func (pool *ClientPool) Accepts(nodeType any) bool
```

Implements \[wiring.NamespaceHandler\]

<a name="ClientPool.AddEdge"></a>
### func \(\*ClientPool\) [AddEdge](<https://github.com/blueprint-uservices/blueprint/blob/main/plugins/clientpool/wiring.go#L84>)

```go
func (pool *ClientPool) AddEdge(name string, edge ir.IRNode) error
```

Implements \[wiring.NamespaceHandler\]

<a name="ClientPool.AddInstantiation"></a>
### func \(\*ClientPool\) [AddInstantiation](<https://github.com/blueprint-uservices/blueprint/blob/main/plugins/clientpool/ir.go#L116>)

```go
func (pool *ClientPool) AddInstantiation(builder golang.NamespaceBuilder) error
```

Implements golang.Service golang.Instantiable

<a name="ClientPool.AddInterfaces"></a>
### func \(\*ClientPool\) [AddInterfaces](<https://github.com/blueprint-uservices/blueprint/blob/main/plugins/clientpool/ir.go#L54>)

```go
func (pool *ClientPool) AddInterfaces(module golang.ModuleBuilder) error
```

Implements golang.Service golang.ProvidesInterface

<a name="ClientPool.AddNode"></a>
### func \(\*ClientPool\) [AddNode](<https://github.com/blueprint-uservices/blueprint/blob/main/plugins/clientpool/wiring.go#L90>)

```go
func (pool *ClientPool) AddNode(name string, node ir.IRNode) error
```

Implements \[wiring.NamespaceHandler\]

<a name="ClientPool.GenerateFuncs"></a>
### func \(\*ClientPool\) [GenerateFuncs](<https://github.com/blueprint-uservices/blueprint/blob/main/plugins/clientpool/ir.go#L67>)

```go
func (pool *ClientPool) GenerateFuncs(module golang.ModuleBuilder) error
```

Implements golang.GeneratesFuncs

<a name="ClientPool.GetInterface"></a>
### func \(\*ClientPool\) [GetInterface](<https://github.com/blueprint-uservices/blueprint/blob/main/plugins/clientpool/ir.go#L48>)

```go
func (pool *ClientPool) GetInterface(ctx ir.BuildContext) (service.ServiceInterface, error)
```

Implements golang.Service service.ServiceNode

<a name="ClientPool.Name"></a>
### func \(\*ClientPool\) [Name](<https://github.com/blueprint-uservices/blueprint/blob/main/plugins/clientpool/ir.go#L30>)

```go
func (pool *ClientPool) Name() string
```

Implements ir.IRNode

<a name="ClientPool.String"></a>
### func \(\*ClientPool\) [String](<https://github.com/blueprint-uservices/blueprint/blob/main/plugins/clientpool/ir.go#L35>)

```go
func (pool *ClientPool) String() string
```

Implements ir.IRNode

Generated by [gomarkdoc](<https://github.com/princjef/gomarkdoc>)
