<!-- Code generated by gomarkdoc. DO NOT EDIT -->

# circuitbreaker

```go
import "github.com/blueprint-uservices/blueprint/plugins/circuitbreaker"
```

Package circuitbreaker provides a Blueprint modifier for the client side of service calls.

The plugin wraps clients with a circuitbreaker that blocks any new requests from being sent out over a connection if the failure rate exceeds a provided number in a fixed duration. The block is removed after the completion of the fixed duration interval.

## Index

- [func AddCircuitBreaker\(spec wiring.WiringSpec, serviceName string, min\_reqs int64, failure\_rate float64, interval string\)](<#AddCircuitBreaker>)
- [type CircuitBreakerClient](<#CircuitBreakerClient>)
  - [func \(node \*CircuitBreakerClient\) AddInstantiation\(builder golang.NamespaceBuilder\) error](<#CircuitBreakerClient.AddInstantiation>)
  - [func \(node \*CircuitBreakerClient\) AddInterfaces\(builder golang.ModuleBuilder\) error](<#CircuitBreakerClient.AddInterfaces>)
  - [func \(node \*CircuitBreakerClient\) GenerateFuncs\(builder golang.ModuleBuilder\) error](<#CircuitBreakerClient.GenerateFuncs>)
  - [func \(node \*CircuitBreakerClient\) GetInterface\(ctx ir.BuildContext\) \(service.ServiceInterface, error\)](<#CircuitBreakerClient.GetInterface>)
  - [func \(node \*CircuitBreakerClient\) ImplementsGolangNode\(\)](<#CircuitBreakerClient.ImplementsGolangNode>)
  - [func \(node \*CircuitBreakerClient\) Name\(\) string](<#CircuitBreakerClient.Name>)
  - [func \(node \*CircuitBreakerClient\) String\(\) string](<#CircuitBreakerClient.String>)


<a name="AddCircuitBreaker"></a>
## func [AddCircuitBreaker](<https://github.com/blueprint-uservices/blueprint/blob/main/plugins/circuitbreaker/wiring.go#L22>)

```go
func AddCircuitBreaker(spec wiring.WiringSpec, serviceName string, min_reqs int64, failure_rate float64, interval string)
```

Adds circuit breaker functionality to all clients of the specified service. Uses a \[blueprint.WiringSpec\]. Circuit breaker trips when \`failure\_rate\` percentage of requests fail. Minimum number of requests for the circuit to break is specified using \`min\_reqs\`. The circuit breaker counters are reset after \`interval\` duration. Usage:

```
AddCircuitBreaker(spec, "serviceA", 1000, 0.1, "1s")
```

<a name="CircuitBreakerClient"></a>
## type [CircuitBreakerClient](<https://github.com/blueprint-uservices/blueprint/blob/main/plugins/circuitbreaker/ir.go#L15-L27>)

Blueprint IR node representing a CircuitBreaker

```go
type CircuitBreakerClient struct {
    golang.Service
    golang.GeneratesFuncs
    golang.Instantiable

    InstanceName string
    Wrapped      golang.Service

    Min_Reqs    int64
    FailureRate float64
    Interval    string
    // contains filtered or unexported fields
}
```

<a name="CircuitBreakerClient.AddInstantiation"></a>
### func \(\*CircuitBreakerClient\) [AddInstantiation](<https://github.com/blueprint-uservices/blueprint/blob/main/plugins/circuitbreaker/ir.go#L77>)

```go
func (node *CircuitBreakerClient) AddInstantiation(builder golang.NamespaceBuilder) error
```



<a name="CircuitBreakerClient.AddInterfaces"></a>
### func \(\*CircuitBreakerClient\) [AddInterfaces](<https://github.com/blueprint-uservices/blueprint/blob/main/plugins/circuitbreaker/ir.go#L56>)

```go
func (node *CircuitBreakerClient) AddInterfaces(builder golang.ModuleBuilder) error
```



<a name="CircuitBreakerClient.GenerateFuncs"></a>
### func \(\*CircuitBreakerClient\) [GenerateFuncs](<https://github.com/blueprint-uservices/blueprint/blob/main/plugins/circuitbreaker/ir.go#L64>)

```go
func (node *CircuitBreakerClient) GenerateFuncs(builder golang.ModuleBuilder) error
```



<a name="CircuitBreakerClient.GetInterface"></a>
### func \(\*CircuitBreakerClient\) [GetInterface](<https://github.com/blueprint-uservices/blueprint/blob/main/plugins/circuitbreaker/ir.go#L60>)

```go
func (node *CircuitBreakerClient) GetInterface(ctx ir.BuildContext) (service.ServiceInterface, error)
```



<a name="CircuitBreakerClient.ImplementsGolangNode"></a>
### func \(\*CircuitBreakerClient\) [ImplementsGolangNode](<https://github.com/blueprint-uservices/blueprint/blob/main/plugins/circuitbreaker/ir.go#L29>)

```go
func (node *CircuitBreakerClient) ImplementsGolangNode()
```



<a name="CircuitBreakerClient.Name"></a>
### func \(\*CircuitBreakerClient\) [Name](<https://github.com/blueprint-uservices/blueprint/blob/main/plugins/circuitbreaker/ir.go#L31>)

```go
func (node *CircuitBreakerClient) Name() string
```



<a name="CircuitBreakerClient.String"></a>
### func \(\*CircuitBreakerClient\) [String](<https://github.com/blueprint-uservices/blueprint/blob/main/plugins/circuitbreaker/ir.go#L35>)

```go
func (node *CircuitBreakerClient) String() string
```



Generated by [gomarkdoc](<https://github.com/princjef/gomarkdoc>)
