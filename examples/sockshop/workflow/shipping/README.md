<!-- Code generated by gomarkdoc. DO NOT EDIT -->

# shipping

```go
import "github.com/blueprint-uservices/blueprint/examples/sockshop/workflow/shipping"
```

Package shipping implements the SockShop shipping microservice.

All the shipping microservice does is push the shipment to a queue. The queue\-master service pulls shipments from the queue and "processes" them.

## Index

- [type Shipment](<#Shipment>)
- [type ShippingService](<#ShippingService>)
  - [func NewShippingService\(ctx context.Context, queue backend.Queue, db backend.NoSQLDatabase\) \(ShippingService, error\)](<#NewShippingService>)


<a name="Shipment"></a>
## type [Shipment](<https://github.com/blueprint-uservices/blueprint/blob/main/examples/sockshop/workflow/shipping/shippingservice.go#L32-L36>)

Represents a shipment for an order

```go
type Shipment struct {
    ID     string
    Name   string
    Status string
}
```

<a name="ShippingService"></a>
## type [ShippingService](<https://github.com/blueprint-uservices/blueprint/blob/main/examples/sockshop/workflow/shipping/shippingservice.go#L17-L29>)

ShippingService implements the SockShop shipping microservice

```go
type ShippingService interface {
    // Submit a shipment to be shipped.  The actual handling of the
    // shipment will happen asynchronously by the queue-master service.
    //
    // Returns the submitted shipment or an error
    PostShipping(ctx context.Context, shipment Shipment) (Shipment, error)

    // Get a shipment's status
    GetShipment(ctx context.Context, id string) (Shipment, error)

    // Update a shipment's status; called by the queue master
    UpdateStatus(ctx context.Context, id, status string) error
}
```

<a name="NewShippingService"></a>
### func [NewShippingService](<https://github.com/blueprint-uservices/blueprint/blob/main/examples/sockshop/workflow/shipping/shippingservice.go#L39>)

```go
func NewShippingService(ctx context.Context, queue backend.Queue, db backend.NoSQLDatabase) (ShippingService, error)
```

Instantiates a shipping service that submits all shipments to a queue for asynchronous background processing

Generated by [gomarkdoc](<https://github.com/princjef/gomarkdoc>)
