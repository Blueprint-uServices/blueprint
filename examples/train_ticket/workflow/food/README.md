<!-- Code generated by gomarkdoc. DO NOT EDIT -->

# food

```go
import "gitlab.mpi-sws.org/cld/blueprint/examples/train_ticket/workflow/food"
```

## Index

- [type Food](<#Food>)
- [type FoodOrder](<#FoodOrder>)


<a name="Food"></a>
## type [Food](<https://gitlab.mpi-sws.org/cld/blueprint2/blueprint/blob/main/examples/train_ticket/workflow/food/data.go#L3-L6>)



```go
type Food struct {
    Name  string
    Price float64
}
```

<a name="FoodOrder"></a>
## type [FoodOrder](<https://gitlab.mpi-sws.org/cld/blueprint2/blueprint/blob/main/examples/train_ticket/workflow/food/data.go#L8-L16>)



```go
type FoodOrder struct {
    ID          string
    OrderID     string
    FoodType    int64
    StationName string
    StoreName   string
    FoodName    string
    Price       float64
}
```

Generated by [gomarkdoc](<https://github.com/princjef/gomarkdoc>)