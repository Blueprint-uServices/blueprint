<!-- Code generated by gomarkdoc. DO NOT EDIT -->

# thrift

```go
import "github.com/blueprint-uservices/blueprint/plugins/thrift"
```

Package thrift implements a Blueprint plugin that enables any Golang service to be deployed using a Thrift server.

To use the plugin in a Blueprint wiring spec, import this package and use the [Deploy](<#Deploy>) method, i.e.

```
import "github.com/blueprint-uservices/blueprint/plugins/thrift"
thrift.Deploy(spec, "my_service")
```

See the documentation for [Deploy](<#Deploy>) for more information about its behavior.

The plugin implements thrift code generation, as well as generating a server\-side handler and a client\-side library that calls the server. This is implemented within the \[thriftcodegen\] pacakge.

To use this plugin, the thrift compiler and version\-matching go bindings are required to be installed on the machine that is compiling the Blueprint wiring spec. Installation instructions can be found: https://thrift.apache.org/download

## Index

- [func Deploy\(spec wiring.WiringSpec, serviceName string\)](<#Deploy>)
- [type ThriftInterface](<#ThriftInterface>)
  - [func \(thrift \*ThriftInterface\) GetMethods\(\) \[\]service.Method](<#ThriftInterface.GetMethods>)
  - [func \(thrift \*ThriftInterface\) GetName\(\) string](<#ThriftInterface.GetName>)


<a name="Deploy"></a>
## func [Deploy](<https://github.com/blueprint-uservices/blueprint/blob/main/plugins/thrift/wiring.go#L41>)

```go
func Deploy(spec wiring.WiringSpec, serviceName string)
```

Deploys \`serviceName\` as a Thrift server.

Typically serviceName should be the name of a workflow service that was initially defined using \[workflow.Define\].

Like many other modifiers, Thrift modifies the service at the golang level, by generating server\-side handler code and a client\-side library. However, Thrift should be the last golang\-level modifier applied to a service, because thereafter communication between the client and server is no longer at the golang level, but at the network level.

Deploying a service with Thrift increases the visibility of the service within the application. By default, any other service running in any other container or namespace can now contact this service.

<a name="ThriftInterface"></a>
## type [ThriftInterface](<https://github.com/blueprint-uservices/blueprint/blob/main/plugins/thrift/ir_thrift_server.go#L30-L33>)



```go
type ThriftInterface struct {
    service.ServiceInterface
    Wrapped service.ServiceInterface
}
```

<a name="ThriftInterface.GetMethods"></a>
### func \(\*ThriftInterface\) [GetMethods](<https://github.com/blueprint-uservices/blueprint/blob/main/plugins/thrift/ir_thrift_server.go#L39>)

```go
func (thrift *ThriftInterface) GetMethods() []service.Method
```



<a name="ThriftInterface.GetName"></a>
### func \(\*ThriftInterface\) [GetName](<https://github.com/blueprint-uservices/blueprint/blob/main/plugins/thrift/ir_thrift_server.go#L35>)

```go
func (thrift *ThriftInterface) GetName() string
```



Generated by [gomarkdoc](<https://github.com/princjef/gomarkdoc>)
