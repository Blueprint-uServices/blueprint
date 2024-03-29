<!-- Code generated by gomarkdoc. DO NOT EDIT -->

# rabbitmq

```go
import "github.com/blueprint-uservices/blueprint/runtime/plugins/rabbitmq"
```

Package rabbitmq provides a client\-wrapper implementation of the \[backend.Queue\] interface for a rabbitmq server.

## Index

- [type RabbitMQ](<#RabbitMQ>)
  - [func NewRabbitMQ\(ctx context.Context, addr string, queue\_name string\) \(\*RabbitMQ, error\)](<#NewRabbitMQ>)
  - [func \(q \*RabbitMQ\) Pop\(ctx context.Context, dst interface\{\}\) \(bool, error\)](<#RabbitMQ.Pop>)
  - [func \(q \*RabbitMQ\) Push\(ctx context.Context, item interface\{\}\) \(bool, error\)](<#RabbitMQ.Push>)


<a name="RabbitMQ"></a>
## type [RabbitMQ](<https://github.com/blueprint-uservices/blueprint/blob/main/runtime/plugins/rabbitmq/queue.go#L13-L19>)

Implements a Queue that uses the rabbitmq package

```go
type RabbitMQ struct {
    // contains filtered or unexported fields
}
```

<a name="NewRabbitMQ"></a>
### func [NewRabbitMQ](<https://github.com/blueprint-uservices/blueprint/blob/main/runtime/plugins/rabbitmq/queue.go#L22>)

```go
func NewRabbitMQ(ctx context.Context, addr string, queue_name string) (*RabbitMQ, error)
```

Instantiates a new \[Queue\] instances that provides a queue interface via a RabbitMQ instance

<a name="RabbitMQ.Pop"></a>
### func \(\*RabbitMQ\) [Pop](<https://github.com/blueprint-uservices/blueprint/blob/main/runtime/plugins/rabbitmq/queue.go#L63>)

```go
func (q *RabbitMQ) Pop(ctx context.Context, dst interface{}) (bool, error)
```

Pop implements backend.Queue

<a name="RabbitMQ.Push"></a>
### func \(\*RabbitMQ\) [Push](<https://github.com/blueprint-uservices/blueprint/blob/main/runtime/plugins/rabbitmq/queue.go#L53>)

```go
func (q *RabbitMQ) Push(ctx context.Context, item interface{}) (bool, error)
```

Push implements backend.Queue

Generated by [gomarkdoc](<https://github.com/princjef/gomarkdoc>)
