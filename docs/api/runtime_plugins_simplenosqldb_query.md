---
title: runtime/plugins/simplenosqldb/query
---
# runtime/plugins/simplenosqldb/query
```go
package query // import "gitlab.mpi-sws.org/cld/blueprint/runtime/plugins/simplenosqldb/query"
```

## TYPES

```go
type CmpType int
```
```go
const (
	Eq CmpType = iota
	Gt
	Gte
	Lt
	Lte
)
type Filter interface {
	Apply(item any) bool
	String() string
}
```
## func All
```go
func All(filters ...Filter) Filter
```

## func And
```go
func And(filters ...Filter) Filter
```

## func Broadcast
```go
func Broadcast(next Filter) Filter
```

## func CmpFloat
```go
func CmpFloat(value float64, cmp CmpType) Filter
```

## func CmpInt
```go
func CmpInt(value int64, cmp CmpType) Filter
```

## func ElemMatch
```go
func ElemMatch(queries ...Filter) Filter
```

## func Equals
```go
func Equals(value any) Filter
```

## func ExactFieldMatch
```go
func ExactFieldMatch(fieldsSeen ...string) Filter
```

## func Exists
```go
func Exists() Filter
```

## func Index
```go
func Index(i int, next Filter) Filter
```

## func Lookup
```go
func Lookup(selector string, next Filter) Filter
```

## func Not
```go
func Not(filter Filter) Filter
```

## func Or
```go
func Or(filters ...Filter) Filter
```

## func ParseFilter
```go
func ParseFilter(filter bson.D) (Filter, error)
```

## func Regex
```go
func Regex(regex_string string) (Filter, error)
```

## func Select
```go
func Select(fieldName string, next Filter) Filter
```

```go
type Update interface {
	Apply(itemRef any) error
	String() string
}
```
## func BroadcastUpdate
```go
func BroadcastUpdate(update Update) Update
```

## func IncInt
```go
func IncInt(amount int64) Update
```

## func ParseUpdate
```go
func ParseUpdate(update bson.D) (Update, error)
```

## func PushValue
```go
func PushValue(value any) (Update, error)
```

## func SetValue
```go
func SetValue(value any) (Update, error)
```

## func UnsetElement
```go
func UnsetElement(index int) Update
```

## func UnsetField
```go
func UnsetField(fieldName string) Update
```

## func UnsetPath
```go
func UnsetPath(selector string) Update
```

## func UpdateAll
```go
func UpdateAll(updates []Update) Update
```

## func UpdateField
```go
func UpdateField(fieldName string, update Update, createIfAbsent bool) Update
```

## func UpdateIndex
```go
func UpdateIndex(index int, update Update, createIfAbsent bool) Update
```

## func UpdatePath
```go
func UpdatePath(selector string, update Update) Update
```


