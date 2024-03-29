<!-- Code generated by gomarkdoc. DO NOT EDIT -->

# route

```go
import "github.com/blueprint-uservices/blueprint/examples/train_ticket/workflow/route"
```

package route implements ts\-route\-service from the original train ticket application

## Index

- [type Route](<#Route>)
- [type RouteInfo](<#RouteInfo>)
- [type RouteService](<#RouteService>)
- [type RouteServiceImpl](<#RouteServiceImpl>)
  - [func NewRouteServiceImpl\(ctx context.Context, db backend.NoSQLDatabase\) \(\*RouteServiceImpl, error\)](<#NewRouteServiceImpl>)
  - [func \(r \*RouteServiceImpl\) CreateAndModify\(ctx context.Context, info RouteInfo\) \(Route, error\)](<#RouteServiceImpl.CreateAndModify>)
  - [func \(r \*RouteServiceImpl\) DeleteRoute\(ctx context.Context, id string\) error](<#RouteServiceImpl.DeleteRoute>)
  - [func \(r \*RouteServiceImpl\) GetAllRoutes\(ctx context.Context\) \(\[\]Route, error\)](<#RouteServiceImpl.GetAllRoutes>)
  - [func \(r \*RouteServiceImpl\) GetRouteById\(ctx context.Context, id string\) \(Route, error\)](<#RouteServiceImpl.GetRouteById>)
  - [func \(r \*RouteServiceImpl\) GetRouteByIds\(ctx context.Context, ids \[\]string\) \(\[\]Route, error\)](<#RouteServiceImpl.GetRouteByIds>)
  - [func \(r \*RouteServiceImpl\) GetRouteByStartAndEnd\(ctx context.Context, start string, end string\) \(Route, error\)](<#RouteServiceImpl.GetRouteByStartAndEnd>)


<a name="Route"></a>
## type [Route](<https://github.com/blueprint-uservices/blueprint/blob/main/examples/train_ticket/workflow/route/data.go#L3-L9>)



```go
type Route struct {
    ID           string
    Stations     []string
    Distances    []int64
    StartStation string
    EndStation   string
}
```

<a name="RouteInfo"></a>
## type [RouteInfo](<https://github.com/blueprint-uservices/blueprint/blob/main/examples/train_ticket/workflow/route/data.go#L11-L17>)



```go
type RouteInfo struct {
    ID           string
    StartStation string
    EndStation   string
    StationList  string
    DistanceList string
}
```

<a name="RouteService"></a>
## type [RouteService](<https://github.com/blueprint-uservices/blueprint/blob/main/examples/train_ticket/workflow/route/routeService.go#L16-L29>)

RouteService manages all the routes in the application

```go
type RouteService interface {
    // Get a route based on the `start` point and `end` point
    GetRouteByStartAndEnd(ctx context.Context, start string, end string) (Route, error)
    // Gets all routes
    GetAllRoutes(ctx context.Context) ([]Route, error)
    // Get a route by ID
    GetRouteById(ctx context.Context, id string) (Route, error)
    // Get multiple routes based on ids
    GetRouteByIds(ctx context.Context, ids []string) ([]Route, error)
    // Delete a route by `id`
    DeleteRoute(ctx context.Context, id string) error
    // Create a new route or modify an existing route based on provided `info` for the route
    CreateAndModify(ctx context.Context, info RouteInfo) (Route, error)
}
```

<a name="RouteServiceImpl"></a>
## type [RouteServiceImpl](<https://github.com/blueprint-uservices/blueprint/blob/main/examples/train_ticket/workflow/route/routeService.go#L31-L33>)



```go
type RouteServiceImpl struct {
    // contains filtered or unexported fields
}
```

<a name="NewRouteServiceImpl"></a>
### func [NewRouteServiceImpl](<https://github.com/blueprint-uservices/blueprint/blob/main/examples/train_ticket/workflow/route/routeService.go#L35>)

```go
func NewRouteServiceImpl(ctx context.Context, db backend.NoSQLDatabase) (*RouteServiceImpl, error)
```



<a name="RouteServiceImpl.CreateAndModify"></a>
### func \(\*RouteServiceImpl\) [CreateAndModify](<https://github.com/blueprint-uservices/blueprint/blob/main/examples/train_ticket/workflow/route/routeService.go#L122>)

```go
func (r *RouteServiceImpl) CreateAndModify(ctx context.Context, info RouteInfo) (Route, error)
```



<a name="RouteServiceImpl.DeleteRoute"></a>
### func \(\*RouteServiceImpl\) [DeleteRoute](<https://github.com/blueprint-uservices/blueprint/blob/main/examples/train_ticket/workflow/route/routeService.go#L39>)

```go
func (r *RouteServiceImpl) DeleteRoute(ctx context.Context, id string) error
```



<a name="RouteServiceImpl.GetAllRoutes"></a>
### func \(\*RouteServiceImpl\) [GetAllRoutes](<https://github.com/blueprint-uservices/blueprint/blob/main/examples/train_ticket/workflow/route/routeService.go#L47>)

```go
func (r *RouteServiceImpl) GetAllRoutes(ctx context.Context) ([]Route, error)
```



<a name="RouteServiceImpl.GetRouteById"></a>
### func \(\*RouteServiceImpl\) [GetRouteById](<https://github.com/blueprint-uservices/blueprint/blob/main/examples/train_ticket/workflow/route/routeService.go#L65>)

```go
func (r *RouteServiceImpl) GetRouteById(ctx context.Context, id string) (Route, error)
```



<a name="RouteServiceImpl.GetRouteByIds"></a>
### func \(\*RouteServiceImpl\) [GetRouteByIds](<https://github.com/blueprint-uservices/blueprint/blob/main/examples/train_ticket/workflow/route/routeService.go#L85>)

```go
func (r *RouteServiceImpl) GetRouteByIds(ctx context.Context, ids []string) ([]Route, error)
```



<a name="RouteServiceImpl.GetRouteByStartAndEnd"></a>
### func \(\*RouteServiceImpl\) [GetRouteByStartAndEnd](<https://github.com/blueprint-uservices/blueprint/blob/main/examples/train_ticket/workflow/route/routeService.go#L98>)

```go
func (r *RouteServiceImpl) GetRouteByStartAndEnd(ctx context.Context, start string, end string) (Route, error)
```



Generated by [gomarkdoc](<https://github.com/princjef/gomarkdoc>)
