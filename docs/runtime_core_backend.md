---
title: runtime/core/backend
---
# runtime/core/backend
```go
package backend // import "gitlab.mpi-sws.org/cld/blueprint/runtime/core/backend"
```

## FUNCTIONS

## func CopyResult
```go
func CopyResult(src any, dst any) error
```
Lots of APIs want to copy results into interfaces. This is a helper method
to do so.

src can be anything; dst must be a pointer to the same type as src

## func GetPointerValue
```go
func GetPointerValue(val any) (any, error)
```
## func GetSpanContext
```go
func GetSpanContext(encoded_string string) (trace.SpanContextConfig, error)
```
Utility function to convert an encoded string into a Span Context

## func SetZero
```go
func SetZero(dst any) error
```
Sets the zero value of a pointer


## TYPES

```go
type Cache interface {
	Put(ctx context.Context, key string, value interface{}) error
	// val is the pointer to which the value will be stored
	Get(ctx context.Context, key string, val interface{}) error
	Mset(ctx context.Context, keys []string, values []interface{}) error
	// values is the array of pointers to which the value will be stored
	Mget(ctx context.Context, keys []string, values []interface{}) error
	Delete(ctx context.Context, key string) error
	Incr(ctx context.Context, key string) (int64, error)
}
```
```go
type NoSQLCollection interface {
	DeleteOne(ctx context.Context, filter bson.D) error
	DeleteMany(ctx context.Context, filter bson.D) error
	InsertOne(ctx context.Context, document interface{}) error
	InsertMany(ctx context.Context, documents []interface{}) error
	FindOne(ctx context.Context, filter bson.D, projection ...bson.D) (NoSQLCursor, error)  //projections should be optional,just like they are in go-mongo and py-mongo. In go-mongo they use an explicit SetProjection method.
	FindMany(ctx context.Context, filter bson.D, projection ...bson.D) (NoSQLCursor, error) // Result is not a slice -> it is an object we can use to retrieve documents using res.All().
	UpdateOne(ctx context.Context, filter bson.D, update bson.D) error
	UpdateMany(ctx context.Context, filter bson.D, update bson.D) error
	ReplaceOne(ctx context.Context, filter bson.D, replacement interface{}) error
	ReplaceMany(ctx context.Context, filter bson.D, replacements ...interface{}) error
}
```
```go
type NoSQLCursor interface {
	One(ctx context.Context, obj interface{}) error
	All(ctx context.Context, obj interface{}) error //similar logic to Decode, but for multiple documents
}
```
```go
type NoSQLDatabase interface {
	//		A NoSQLDatabse implementation might distinguish between databases and collections,
	//		or might not have those concepts.
```
```go
	GetCollection(ctx context.Context, db_name string, collection_name string) (NoSQLCollection, error)
}
```
```go
type Tracer interface {
	GetTracerProvider(ctx context.Context) (trace.TracerProvider, error)
}
```
```go
type XTracer interface {
	Log(ctx context.Context, msg string) (context.Context, error)
	LogWithTags(ctx context.Context, msg string, tags ...string) (context.Context, error)
	StartTask(ctx context.Context, tags ...string) (context.Context, error)
	StopTask(ctx context.Context) (context.Context, error)
	Merge(ctx context.Context, other tracingplane.BaggageContext) (context.Context, error)
	Set(ctx context.Context, baggage tracingplane.BaggageContext) (context.Context, error)
	Get(ctx context.Context) (tracingplane.BaggageContext, error)
	IsTracing(ctx context.Context) (bool, error)
}
```

