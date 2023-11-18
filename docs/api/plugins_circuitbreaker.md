---
title: plugins/circuitbreaker
---
# plugins/circuitbreaker
```go
package circuitbreaker // import "gitlab.mpi-sws.org/cld/blueprint/plugins/circuitbreaker"
```

## FUNCTIONS

## func AddCircuitBreaker
```go
func AddCircuitBreaker(spec wiring.WiringSpec, serviceName string, min_reqs int64, failure_rate float64, interval string)
```
Adds circuit breaker functionality to all clients of the specified service.
Uses a blueprint.WiringSpec. Circuit breaker trips when `failure_rate`
percentage of requests fail. Minimum number of requests for the circuit to
break is specified using `min_reqs`. The circuit breaker counters are reset
after `interval` duration. Usage:

    AddCircuitBreaker(spec, "serviceA", 1000, 0.1, "1s")


## TYPES

```go
type CircuitBreakerClient struct {
	golang.Service
	golang.GeneratesFuncs
	golang.Instantiable
```
```go
	InstanceName string
	Wrapped      golang.Service
```
```go
	Min_Reqs    int64
	FailureRate float64
	Interval    string
	// Has unexported fields.
}
```
## func 
```go
func (node *CircuitBreakerClient) AddInstantiation(builder golang.GraphBuilder) error
```

## func 
```go
func (node *CircuitBreakerClient) AddInterfaces(builder golang.ModuleBuilder) error
```

## func 
```go
func (node *CircuitBreakerClient) GenerateFuncs(builder golang.ModuleBuilder) error
```

## func 
```go
func (node *CircuitBreakerClient) GetInterface(ctx ir.BuildContext) (service.ServiceInterface, error)
```

## func 
```go
func (node *CircuitBreakerClient) ImplementsGolangNode()
```

## func 
```go
func (node *CircuitBreakerClient) Name() string
```

## func 
```go
func (node *CircuitBreakerClient) String() string
```


