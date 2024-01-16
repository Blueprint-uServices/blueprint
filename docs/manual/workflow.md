# Workflow

A **Workflow Spec** or just **Workflow** defines the core business logic of an application.  For example, in a social network application, the workflow defines how users can upload posts, view their timeline feed, follow other users, etc.

In Blueprint, a workflow is implemented without reference to any of the libraries of infrastructure needed to deploy the workflow.  A workflow does not need to bind to an RPC library like gRPC or implement a mechanism like retries.  Instead, these are integrated into the workflow code later by **Blueprint's Compiler**.

While developing an application's workflow, the philosophy should be to assume *nothing* about exactly how the application will be deployed.  Services might be deployed into different processes running on the same machine; in containers distributed across a cluster; or even, directly combined into a single monolith application.

## Project Layout

A Blueprint application will likely comprise several golang modules, primarily for the application's workflow spec and wiring spec(s).  By convention, we recommend placing these modules in sibling directories (e.g. `workflow` and `wiring` directories).
The[Sock Shop](../../examples/sockshop/) application demonstrates this structure and convention.

The `workflow` subdirectory will contain your workflow implementation.  Your workflow module will likely want a dependency on the `github.com/blueprint-uservices/blueprint/runtime` module.

Later, you may choose to also create a `tests` module for [Workflow Tests](workflow_tests.md) and a `workload` module for a custom [Workload Generator](../../plugins/workload), though these are not needed yet.

## Workflow Services

A Workflow consists of a number of inter-related **Services**.  A service is akin to a microservice or a class that provides some public methods; other services can call those methods.

Define a service by declaring an interface with some methods:
```
type EchoService interface {
    // Echoes the provided message back to the caller
    Echo(ctx context.Context, message string) (string, error)
}
```

Implement the service with a struct
```
type echoServiceImpl struct {}

func NewEchoService(ctx context.Context) (EchoService, error) {
    return &echoServiceImpl{}, nil
}

func (s *echoServiceImpl) Echo(ctx context.Context, message string) (string, error) {
    return message, nil
}
```

The above is sufficient to compile the `EchoService` to a process, docker container, etc. and make use of any of the plugins offered by Blueprint.

### Rules

Blueprint requires the following from workflow services:

 * A service must be defined by an interface, e.g. `EchoService`
 * The first argument of all service methods is a `context.Context`
 * The final return value of all service methods is an `error`
 * A service constructor must be defined that returns a service instance, e.g. `NewEchoService`
 * The first argument of a constructor must be a `context.Context`
 * The return value of a constructor must be the service instance and an error, e.g. `(EchoService, error)`

A workflow can import and make use of any 3rd party libraries it desires.

## Calling other Workflow Services

A service can make calls to other services.  To do so, the service needs a reference to those other services.

```
type MultiEchoer struct {
    // Calls the EchoService n times
    MultiEcho(ctx context.Context, message string, times int) (string, error)
}

type multiEchoerImpl struct {
    echo EchoService
}

func NewMultiEchoer(ctx context.Context, echo EchoService) (MultiEchoer, error) {
    return &multiEchoerImpl{echo: echo}
}

func (s *multiEchoerImpl) MultiEcho(ctx context.Context, message string, times int) (string, error) {
    var b strings.Builder
    for i := 0; i < times; i++ {
        echoed, err := s.echo.Echo(ctx, message)
        if err != nil {
            return "", err
        }
        b.WriteString(echoed + "\n")
    }
    return b.String(), nil
}
```

In the above, `MultiEchoer` calls `Echo` by directly invoking `s.echo.Echo`.

The above is sufficient for the `MultiEchoer` service to be compiled and deployed, and to call the `EchoService` even if running in a different process, machine, or container.

### Rules for Calling other Services

* If a service calls another service, it can only receive a reference to the other service as a constructor argument, e.g. `NewMultiEchoer(ctx context.Context, echo EchoService)`.  It cannot instantiate the other service directly.

## Backends

Some services want to persist data in backends, such as in a database, or make use of other features like a cache.  Backends behave much like services: they have an interface, and Blueprint is responsible for compiling them.

Several backends are defined in Blueprint's `runtime` module.  To make use of them, use the following import:

```
import "github.com/blueprint-uservices/blueprint/runtime/core/backend"
```

We can update the MultiEchoer to use a [`backend.Cache`](../../runtime/core/backend/cache.go) and attempt to lookup cached entries.

```
type MultiEchoer struct {
    // Calls the EchoService n times
    MultiEcho(ctx context.Context, message string, times int) (string, error)
}

type multiEchoerImpl struct {
    echo EchoService
    cache backend.Cache
}

func NewMultiEchoer(ctx context.Context, echo EchoService, cache backend.Cache) (MultiEchoer, error) {
    return &multiEchoerImpl{echo: echo, cache: cache}
}

func (s *multiEchoerImpl) MultiEcho(ctx context.Context, message string, times int) (string, error) {
    var b strings.Builder
    for i := 0; i < times; i++ {
        var echoed string
        if err := s.cache.Get(ctx, message, &echoed); err != nil {
            // not present in cache; call EchoService
            echoed, err = s.echo.Echo(ctx, message)
            if err != nil {
                return "", err
            }
            s.cache.Put(ctx, message, echoed)
        }
        b.WriteString(echoed + "\n")
    }
    return b.String(), nil
}
```

### Rules for backends

Backends do not impose any additional rules.  Like services, they must be passed as constructor arguments.

### List of backends

The [runtime/core](../../runtime/core) package provides the interfaces for a number of commonplace backends.

```
import "github.com/blueprint-uservices/blueprint/runtime/core/backend"
```

* `backend.Cache` an interface for key-value caches; implementations for use in Wiring Specs include [simplecache](../../plugins/simple) and [memcached](../../plugins/memcached)
* `backend.Queue` an interface for queues with push/pop; implementations for use in Wiring Specs include [simplequeue](../../plugins/simple) and [rabbitmq](../../plugins/rabbitmq)
* `backend.NoSQLDatabase` an interface for NoSQL databases that uses MongoDB-style BSON queries; implementations for use in Wiring Specs include [simplenosqldb](../../plugins/simple) and [mongodb](../../plugins/mongodb)
* `backend.RelationalDB` an interface for SQL-based relational databases; implementations for use in Wiring Specs include [simplereldb](../../plugins/simple) and [mysql](../../plugins/mysql)

## Background Tasks

Some services might want to run additional background goroutines.  For example, a service that polls a queue will need to have a goroutine to do so.

The recommended way to implement background goroutines is by implementing a method `Run(context.Context) error`.  This will be automatically invoked in the generated code.

```
func (s *multiEchoerImpl) Run(ctx context.Context) error {
    fmt.Println("I'm running from a different goroutine!")
}
```

## Next Steps

The above is sufficient for defining an application's workflow.

Typically, a workflow will reside in its own Go module, and it will import the Blueprint runtime module.  A workflow can import and make use of any 3rd party libraries it desires.

### Tests

White-box unit tests can be written alongside the workflow spec using typical Golang conventions.

Black-box unit tests we recommend are written in a separate `tests` module alongside the `workflow` as this allows Blueprint to auto-generate tests for the compiled application.

See [Workflow Tests](workflow_tests.md) for more details about tests.

### Compiling the application

See [Wiring Spec](wiring.md) for the next steps to compile the application.