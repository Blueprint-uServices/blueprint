<!-- Code generated by gomarkdoc. DO NOT EDIT -->

# queuemaster

```go
import "gitlab.mpi-sws.org/cld/blueprint/examples/sockshop/workflow/queuemaster"
```

Package queuemaster implements the queue\-master SockShop service, responsible for pulling and "processing" shipments from the shipment queue.

## Index

- [type QueueMaster](<#QueueMaster>)
  - [func NewQueueMaster\(ctx context.Context, queue backend.Queue, shipping shipping.ShippingService\) \(QueueMaster, error\)](<#NewQueueMaster>)


<a name="QueueMaster"></a>
## type QueueMaster

QueueMaster implements the SockShop queue\-master microservice.

It is not a service that can be called; instead it pulls shipments from the shipments queue

```go
type QueueMaster interface {
    // Runs the background goroutine that continually pulls elements from
    // the queue.  Does not return until ctx is cancelled or an error is
    // encountered
    Run(ctx context.Context) error
}
```

<a name="NewQueueMaster"></a>
### func NewQueueMaster

```go
func NewQueueMaster(ctx context.Context, queue backend.Queue, shipping shipping.ShippingService) (QueueMaster, error)
```

Creates a new QueueMaster service.

New: once an order is shipped, it will update the order status in the orderservice.

Generated by [gomarkdoc](<https://github.com/princjef/gomarkdoc>)