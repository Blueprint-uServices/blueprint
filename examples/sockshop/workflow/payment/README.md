<!-- Code generated by gomarkdoc. DO NOT EDIT -->

# payment

```go
import "github.com/blueprint-uservices/blueprint/examples/sockshop/workflow/payment"
```

Package payment implements the SockShop payment microservice.

The service fakes payments, implementing simple logic whereby payments are authorized when they're below a predefined threshold, and rejected when they are above that threshold.

## Index

- [Variables](<#variables>)
- [type Authorisation](<#Authorisation>)
- [type PaymentService](<#PaymentService>)
  - [func NewPaymentService\(ctx context.Context, declineOverAmount string\) \(PaymentService, error\)](<#NewPaymentService>)


## Variables

<a name="ErrInvalidPaymentAmount"></a>

```go
var ErrInvalidPaymentAmount = errors.New("invalid payment amount")
```

<a name="Authorisation"></a>
## type [Authorisation](<https://github.com/blueprint-uservices/blueprint/blob/main/examples/sockshop/workflow/payment/paymentservice.go#L20-L23>)



```go
type Authorisation struct {
    Authorised bool   `json:"authorised"`
    Message    string `json:"message"`
}
```

<a name="PaymentService"></a>
## type [PaymentService](<https://github.com/blueprint-uservices/blueprint/blob/main/examples/sockshop/workflow/payment/paymentservice.go#L16-L18>)

PaymentService provides payment services

```go
type PaymentService interface {
    Authorise(ctx context.Context, amount float32) (Authorisation, error)
}
```

<a name="NewPaymentService"></a>
### func [NewPaymentService](<https://github.com/blueprint-uservices/blueprint/blob/main/examples/sockshop/workflow/payment/paymentservice.go#L27>)

```go
func NewPaymentService(ctx context.Context, declineOverAmount string) (PaymentService, error)
```

Returns a payment service where any transaction above the preconfigured threshold will return an invalid payment amount

Generated by [gomarkdoc](<https://github.com/princjef/gomarkdoc>)
