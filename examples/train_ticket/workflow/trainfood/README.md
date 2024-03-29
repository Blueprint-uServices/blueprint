<!-- Code generated by gomarkdoc. DO NOT EDIT -->

# trainfood

```go
import "github.com/blueprint-uservices/blueprint/examples/train_ticket/workflow/trainfood"
```

package trainfood implements ts\-train\-food\-service from the original train ticket application

## Index

- [type TrainFood](<#TrainFood>)
- [type TrainFoodService](<#TrainFoodService>)
- [type TrainFoodServiceImpl](<#TrainFoodServiceImpl>)
  - [func NewTrainFoodServiceImpl\(ctx context.Context, db backend.NoSQLDatabase\) \(\*TrainFoodServiceImpl, error\)](<#NewTrainFoodServiceImpl>)
  - [func \(t \*TrainFoodServiceImpl\) Cleanup\(ctx context.Context\) error](<#TrainFoodServiceImpl.Cleanup>)
  - [func \(t \*TrainFoodServiceImpl\) CreateTrainFood\(ctx context.Context, tf TrainFood\) \(TrainFood, error\)](<#TrainFoodServiceImpl.CreateTrainFood>)
  - [func \(t \*TrainFoodServiceImpl\) ListTrainFood\(ctx context.Context\) \(\[\]TrainFood, error\)](<#TrainFoodServiceImpl.ListTrainFood>)
  - [func \(t \*TrainFoodServiceImpl\) ListTrainFoodByTripID\(ctx context.Context, tripid string\) \(\[\]food.Food, error\)](<#TrainFoodServiceImpl.ListTrainFoodByTripID>)


<a name="TrainFood"></a>
## type [TrainFood](<https://github.com/blueprint-uservices/blueprint/blob/main/examples/train_ticket/workflow/trainfood/data.go#L5-L9>)



```go
type TrainFood struct {
    ID     string
    TripID string
    Foods  []food.Food
}
```

<a name="TrainFoodService"></a>
## type [TrainFoodService](<https://github.com/blueprint-uservices/blueprint/blob/main/examples/train_ticket/workflow/trainfood/trainFoodService.go#L14-L23>)

TrainFoodService manages food items in Trains

```go
type TrainFoodService interface {
    // Creates a new Train Food item
    CreateTrainFood(ctx context.Context, tf TrainFood) (TrainFood, error)
    // Get all train food items
    ListTrainFood(ctx context.Context) ([]TrainFood, error)
    // List all food items based on `tripid`
    ListTrainFoodByTripID(ctx context.Context, tripid string) ([]food.Food, error)
    // Remove all train food items; Only used during testing
    Cleanup(ctx context.Context) error
}
```

<a name="TrainFoodServiceImpl"></a>
## type [TrainFoodServiceImpl](<https://github.com/blueprint-uservices/blueprint/blob/main/examples/train_ticket/workflow/trainfood/trainFoodService.go#L26-L28>)

Implementation of TrainFoodService

```go
type TrainFoodServiceImpl struct {
    // contains filtered or unexported fields
}
```

<a name="NewTrainFoodServiceImpl"></a>
### func [NewTrainFoodServiceImpl](<https://github.com/blueprint-uservices/blueprint/blob/main/examples/train_ticket/workflow/trainfood/trainFoodService.go#L31>)

```go
func NewTrainFoodServiceImpl(ctx context.Context, db backend.NoSQLDatabase) (*TrainFoodServiceImpl, error)
```

Creates a NewTrainFoodService

<a name="TrainFoodServiceImpl.Cleanup"></a>
### func \(\*TrainFoodServiceImpl\) [Cleanup](<https://github.com/blueprint-uservices/blueprint/blob/main/examples/train_ticket/workflow/trainfood/trainFoodService.go#L100>)

```go
func (t *TrainFoodServiceImpl) Cleanup(ctx context.Context) error
```



<a name="TrainFoodServiceImpl.CreateTrainFood"></a>
### func \(\*TrainFoodServiceImpl\) [CreateTrainFood](<https://github.com/blueprint-uservices/blueprint/blob/main/examples/train_ticket/workflow/trainfood/trainFoodService.go#L72>)

```go
func (t *TrainFoodServiceImpl) CreateTrainFood(ctx context.Context, tf TrainFood) (TrainFood, error)
```



<a name="TrainFoodServiceImpl.ListTrainFood"></a>
### func \(\*TrainFoodServiceImpl\) [ListTrainFood](<https://github.com/blueprint-uservices/blueprint/blob/main/examples/train_ticket/workflow/trainfood/trainFoodService.go#L35>)

```go
func (t *TrainFoodServiceImpl) ListTrainFood(ctx context.Context) ([]TrainFood, error)
```



<a name="TrainFoodServiceImpl.ListTrainFoodByTripID"></a>
### func \(\*TrainFoodServiceImpl\) [ListTrainFoodByTripID](<https://github.com/blueprint-uservices/blueprint/blob/main/examples/train_ticket/workflow/trainfood/trainFoodService.go#L52>)

```go
func (t *TrainFoodServiceImpl) ListTrainFoodByTripID(ctx context.Context, tripid string) ([]food.Food, error)
```



Generated by [gomarkdoc](<https://github.com/princjef/gomarkdoc>)
