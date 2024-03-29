<!-- Code generated by gomarkdoc. DO NOT EDIT -->

# simplequeue

```go
import "github.com/blueprint-uservices/blueprint/runtime/plugins/simplequeue"
```

Package simplequeue implements an simple in\-memory \[backend.Queue\] that internally uses a golang channel of capacity 10 for passing items from producer to consumer.

Calls to \[backend.Queue.Push\] will block once the queue capacity reaches 10.

## Index

- [type SimpleQueue](<#SimpleQueue>)
  - [func NewSimpleQueue\(ctx context.Context\) \(q \*SimpleQueue, err error\)](<#NewSimpleQueue>)
  - [func \(q \*SimpleQueue\) Pop\(ctx context.Context, dst interface\{\}\) \(bool, error\)](<#SimpleQueue.Pop>)
  - [func \(q \*SimpleQueue\) Push\(ctx context.Context, item interface\{\}\) \(bool, error\)](<#SimpleQueue.Push>)


<a name="SimpleQueue"></a>
## type [SimpleQueue](<https://github.com/blueprint-uservices/blueprint/blob/main/runtime/plugins/simplequeue/queue.go#L14-L17>)

A simple chan\-based queue that implements the \[backend.Queue\] interface

```go
type SimpleQueue struct {
    backend.Queue
    // contains filtered or unexported fields
}
```

<a name="NewSimpleQueue"></a>
### func [NewSimpleQueue](<https://github.com/blueprint-uservices/blueprint/blob/main/runtime/plugins/simplequeue/queue.go#L22>)

```go
func NewSimpleQueue(ctx context.Context) (q *SimpleQueue, err error)
```

Instantiates a \[backend.Queue\] that internally uses a golang channel of capacity 10.

Calls to \[q.Push\] will block once the queue capacity reaches 10.

<a name="SimpleQueue.Pop"></a>
### func \(\*SimpleQueue\) [Pop](<https://github.com/blueprint-uservices/blueprint/blob/main/runtime/plugins/simplequeue/queue.go#L34>)

```go
func (q *SimpleQueue) Pop(ctx context.Context, dst interface{}) (bool, error)
```

Pop implements backend.Queue.

<a name="SimpleQueue.Push"></a>
### func \(\*SimpleQueue\) [Push](<https://github.com/blueprint-uservices/blueprint/blob/main/runtime/plugins/simplequeue/queue.go#L51>)

```go
func (q *SimpleQueue) Push(ctx context.Context, item interface{}) (bool, error)
```

Push implements backend.Queue.

Generated by [gomarkdoc](<https://github.com/princjef/gomarkdoc>)
