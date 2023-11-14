---
title: runtime/plugins/simplenosqldb
---
# runtime/plugins/simplenosqldb
```go
package simplenosqldb // import "gitlab.mpi-sws.org/cld/blueprint/runtime/plugins/simplenosqldb"
```

## FUNCTIONS

## func CopyResult
```go
func CopyResult(src []bson.D, dst any) error
```

## TYPES

Simple implementations of the NoSQLDB Interfaces from runtime/core/backend
```go
type SimpleCollection struct {
	// Has unexported fields.
}
```
## func 
```go
func (db *SimpleCollection) DeleteMany(ctx context.Context, filter bson.D) error
```

## func 
```go
func (db *SimpleCollection) DeleteOne(ctx context.Context, filter bson.D) error
```

## func 
```go
func (db *SimpleCollection) FindMany(ctx context.Context, filter bson.D, projection ...bson.D) (backend.NoSQLCursor, error)
```

## func 
```go
func (db *SimpleCollection) FindOne(ctx context.Context, filter bson.D, projection ...bson.D) (backend.NoSQLCursor, error)
```

## func 
```go
func (db *SimpleCollection) InsertMany(ctx context.Context, documents []interface{}) error
```

## func 
```go
func (db *SimpleCollection) InsertOne(ctx context.Context, document interface{}) error
```

## func 
```go
func (db *SimpleCollection) ReplaceMany(ctx context.Context, filter bson.D, replacements ...interface{}) error
```

## func 
```go
func (db *SimpleCollection) ReplaceOne(ctx context.Context, filter bson.D, replacement interface{}) error
```

## func 
```go
func (db *SimpleCollection) UpdateMany(ctx context.Context, filter bson.D, update bson.D) error
```

## func 
```go
func (db *SimpleCollection) UpdateOne(ctx context.Context, filter bson.D, update bson.D) error
```

Simple implementations of the NoSQLDB Interfaces from runtime/core/backend
```go
type SimpleCursor struct {
	// Has unexported fields.
}
```
## func 
```go
func (c *SimpleCursor) All(ctx context.Context, obj interface{}) error
```

## func 
```go
func (c *SimpleCursor) One(ctx context.Context, obj interface{}) error
```

Simple implementations of the NoSQLDB Interfaces from runtime/core/backend
```go
type SimpleNoSQLDB struct {
	// Has unexported fields.
}
```
## func NewSimpleNoSQLDB
```go
func NewSimpleNoSQLDB(ctx context.Context) (*SimpleNoSQLDB, error)
```

## func 
```go
func (impl *SimpleNoSQLDB) GetCollection(ctx context.Context, db_name string, collection_name string) (backend.NoSQLCollection, error)
```


