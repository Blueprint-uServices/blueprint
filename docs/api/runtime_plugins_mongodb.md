---
title: runtime/plugins/mongodb
---
# runtime/plugins/mongodb
```go
package mongodb // import "gitlab.mpi-sws.org/cld/blueprint/runtime/plugins/mongodb"
```

## TYPES

```go
type MongoCollection struct {
	// Has unexported fields.
}
```
## func 
```go
func (mc *MongoCollection) DeleteMany(ctx context.Context, filter bson.D) error
```

## func 
```go
func (mc *MongoCollection) DeleteOne(ctx context.Context, filter bson.D) error
```

## func 
```go
func (mc *MongoCollection) FindMany(ctx context.Context, filter bson.D, projection ...bson.D) (backend.NoSQLCursor, error)
```

## func 
```go
func (mc *MongoCollection) FindOne(ctx context.Context, filter bson.D, projection ...bson.D) (backend.NoSQLCursor, error)
```

## func 
```go
func (mc *MongoCollection) InsertMany(ctx context.Context, documents []interface{}) error
```

## func 
```go
func (mc *MongoCollection) InsertOne(ctx context.Context, document interface{}) error
```

## func 
```go
func (mc *MongoCollection) ReplaceMany(ctx context.Context, filter bson.D, replacements ...interface{}) error
```

## func 
```go
func (mc *MongoCollection) ReplaceOne(ctx context.Context, filter bson.D, replacement interface{}) error
```

## func 
```go
func (mc *MongoCollection) UpdateMany(ctx context.Context, filter bson.D, update bson.D) error
```

## func 
```go
func (mc *MongoCollection) UpdateOne(ctx context.Context, filter bson.D, update bson.D) error
```
* not sure about the `update` parameter and its conversion

```go
type MongoCursor struct {
	// Has unexported fields.
}
```
## func 
```go
func (mr *MongoCursor) All(ctx context.Context, objs interface{}) error
```

## func 
```go
func (mr *MongoCursor) One(ctx context.Context, obj interface{}) error
```

```go
type MongoDB struct {
	// Has unexported fields.
}
```
## func NewMongoDB
```go
func NewMongoDB(ctx context.Context, addr string) (*MongoDB, error)
```
* constructor

## func 
```go
func (md *MongoDB) GetCollection(ctx context.Context, db_name string, collectionName string) (backend.NoSQLCollection, error)
```


