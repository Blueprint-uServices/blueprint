<!-- Code generated by gomarkdoc. DO NOT EDIT -->

# workloadgen

```go
import "github.com/blueprint-uservices/blueprint/examples/sockshop/workload/workloadgen"
```

## Index

- [type SimpleWorkload](<#SimpleWorkload>)
  - [func NewSimpleWorkload\(ctx context.Context, frontend frontend.Frontend\) \(SimpleWorkload, error\)](<#NewSimpleWorkload>)


<a name="SimpleWorkload"></a>
## type [SimpleWorkload](<https://github.com/blueprint-uservices/blueprint/blob/main/examples/sockshop/workload/workloadgen/workload.go#L15-L17>)

The WorkloadGen interface, which the Blueprint compiler will treat as a Workflow service

```go
type SimpleWorkload interface {
    ImplementsSimpleWorkload(context.Context) error
}
```

<a name="NewSimpleWorkload"></a>
### func [NewSimpleWorkload](<https://github.com/blueprint-uservices/blueprint/blob/main/examples/sockshop/workload/workloadgen/workload.go#L28>)

```go
func NewSimpleWorkload(ctx context.Context, frontend frontend.Frontend) (SimpleWorkload, error)
```



Generated by [gomarkdoc](<https://github.com/princjef/gomarkdoc>)
