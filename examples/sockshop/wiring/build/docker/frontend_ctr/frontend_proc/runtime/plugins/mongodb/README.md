<!-- Code generated by gomarkdoc. DO NOT EDIT -->

# mongodb

```go
import "github.com/blueprint-uservices/blueprint/runtime/plugins/mongodb"
```

Package mongodb implements a cleint interface to a mongodb server that supports MongoDB's query and update API.

## Index

- [type MongoCollection](<#MongoCollection>)
  - [func \(mc \*MongoCollection\) DeleteMany\(ctx context.Context, filter bson.D\) error](<#MongoCollection.DeleteMany>)
  - [func \(mc \*MongoCollection\) DeleteOne\(ctx context.Context, filter bson.D\) error](<#MongoCollection.DeleteOne>)
  - [func \(mc \*MongoCollection\) FindMany\(ctx context.Context, filter bson.D, projection ...bson.D\) \(backend.NoSQLCursor, error\)](<#MongoCollection.FindMany>)
  - [func \(mc \*MongoCollection\) FindOne\(ctx context.Context, filter bson.D, projection ...bson.D\) \(backend.NoSQLCursor, error\)](<#MongoCollection.FindOne>)
  - [func \(mc \*MongoCollection\) InsertMany\(ctx context.Context, documents \[\]interface\{\}\) error](<#MongoCollection.InsertMany>)
  - [func \(mc \*MongoCollection\) InsertOne\(ctx context.Context, document interface\{\}\) error](<#MongoCollection.InsertOne>)
  - [func \(mc \*MongoCollection\) ReplaceMany\(ctx context.Context, filter bson.D, replacements ...interface\{\}\) \(int, error\)](<#MongoCollection.ReplaceMany>)
  - [func \(mc \*MongoCollection\) ReplaceOne\(ctx context.Context, filter bson.D, replacement interface\{\}\) \(int, error\)](<#MongoCollection.ReplaceOne>)
  - [func \(mc \*MongoCollection\) UpdateMany\(ctx context.Context, filter bson.D, update bson.D\) \(int, error\)](<#MongoCollection.UpdateMany>)
  - [func \(mc \*MongoCollection\) UpdateOne\(ctx context.Context, filter bson.D, update bson.D\) \(int, error\)](<#MongoCollection.UpdateOne>)
  - [func \(mc \*MongoCollection\) Upsert\(ctx context.Context, filter bson.D, document interface\{\}\) \(bool, error\)](<#MongoCollection.Upsert>)
  - [func \(mc \*MongoCollection\) UpsertID\(ctx context.Context, id primitive.ObjectID, document interface\{\}\) \(bool, error\)](<#MongoCollection.UpsertID>)
- [type MongoCursor](<#MongoCursor>)
  - [func \(mr \*MongoCursor\) All\(ctx context.Context, objs interface\{\}\) error](<#MongoCursor.All>)
  - [func \(mr \*MongoCursor\) One\(ctx context.Context, obj interface\{\}\) \(bool, error\)](<#MongoCursor.One>)
- [type MongoDB](<#MongoDB>)
  - [func NewMongoDB\(ctx context.Context, addr string\) \(\*MongoDB, error\)](<#NewMongoDB>)
  - [func \(md \*MongoDB\) GetCollection\(ctx context.Context, db\_name string, collectionName string\) \(backend.NoSQLCollection, error\)](<#MongoDB.GetCollection>)


<a name="MongoCollection"></a>
## type [MongoCollection](<https://github.com/blueprint-uservices/blueprint/blob/main/runtime/plugins/mongodb/nosqldb.go#L21-L23>)

Implements the \[backend.NoSQLCollection\] interface as a client\-wrapper to a mongodb server

```go
type MongoCollection struct {
    // contains filtered or unexported fields
}
```

<a name="MongoCollection.DeleteMany"></a>
### func \(\*MongoCollection\) [DeleteMany](<https://github.com/blueprint-uservices/blueprint/blob/main/runtime/plugins/mongodb/nosqldb.go#L58>)

```go
func (mc *MongoCollection) DeleteMany(ctx context.Context, filter bson.D) error
```

Implements the \[backend.NoSQLCollection\] interface

<a name="MongoCollection.DeleteOne"></a>
### func \(\*MongoCollection\) [DeleteOne](<https://github.com/blueprint-uservices/blueprint/blob/main/runtime/plugins/mongodb/nosqldb.go#L49>)

```go
func (mc *MongoCollection) DeleteOne(ctx context.Context, filter bson.D) error
```

Implements the \[backend.NoSQLCollection\] interface

<a name="MongoCollection.FindMany"></a>
### func \(\*MongoCollection\) [FindMany](<https://github.com/blueprint-uservices/blueprint/blob/main/runtime/plugins/mongodb/nosqldb.go#L105>)

```go
func (mc *MongoCollection) FindMany(ctx context.Context, filter bson.D, projection ...bson.D) (backend.NoSQLCursor, error)
```

Implements the \[backend.NoSQLCollection\] interface

<a name="MongoCollection.FindOne"></a>
### func \(\*MongoCollection\) [FindOne](<https://github.com/blueprint-uservices/blueprint/blob/main/runtime/plugins/mongodb/nosqldb.go#L78>)

```go
func (mc *MongoCollection) FindOne(ctx context.Context, filter bson.D, projection ...bson.D) (backend.NoSQLCursor, error)
```

Implements the \[backend.NoSQLCollection\] interface

<a name="MongoCollection.InsertMany"></a>
### func \(\*MongoCollection\) [InsertMany](<https://github.com/blueprint-uservices/blueprint/blob/main/runtime/plugins/mongodb/nosqldb.go#L71>)

```go
func (mc *MongoCollection) InsertMany(ctx context.Context, documents []interface{}) error
```

Implements the \[backend.NoSQLCollection\] interface

<a name="MongoCollection.InsertOne"></a>
### func \(\*MongoCollection\) [InsertOne](<https://github.com/blueprint-uservices/blueprint/blob/main/runtime/plugins/mongodb/nosqldb.go#L64>)

```go
func (mc *MongoCollection) InsertOne(ctx context.Context, document interface{}) error
```

Implements the \[backend.NoSQLCollection\] interface

<a name="MongoCollection.ReplaceMany"></a>
### func \(\*MongoCollection\) [ReplaceMany](<https://github.com/blueprint-uservices/blueprint/blob/main/runtime/plugins/mongodb/nosqldb.go#L183>)

```go
func (mc *MongoCollection) ReplaceMany(ctx context.Context, filter bson.D, replacements ...interface{}) (int, error)
```

Implements the \[backend.NoSQLCollection\] interface

<a name="MongoCollection.ReplaceOne"></a>
### func \(\*MongoCollection\) [ReplaceOne](<https://github.com/blueprint-uservices/blueprint/blob/main/runtime/plugins/mongodb/nosqldb.go#L173>)

```go
func (mc *MongoCollection) ReplaceOne(ctx context.Context, filter bson.D, replacement interface{}) (int, error)
```

Implements the \[backend.NoSQLCollection\] interface

<a name="MongoCollection.UpdateMany"></a>
### func \(\*MongoCollection\) [UpdateMany](<https://github.com/blueprint-uservices/blueprint/blob/main/runtime/plugins/mongodb/nosqldb.go#L145>)

```go
func (mc *MongoCollection) UpdateMany(ctx context.Context, filter bson.D, update bson.D) (int, error)
```

Implements the \[backend.NoSQLCollection\] interface

<a name="MongoCollection.UpdateOne"></a>
### func \(\*MongoCollection\) [UpdateOne](<https://github.com/blueprint-uservices/blueprint/blob/main/runtime/plugins/mongodb/nosqldb.go#L135>)

```go
func (mc *MongoCollection) UpdateOne(ctx context.Context, filter bson.D, update bson.D) (int, error)
```

Implements the \[backend.NoSQLCollection\] interface

<a name="MongoCollection.Upsert"></a>
### func \(\*MongoCollection\) [Upsert](<https://github.com/blueprint-uservices/blueprint/blob/main/runtime/plugins/mongodb/nosqldb.go#L155>)

```go
func (mc *MongoCollection) Upsert(ctx context.Context, filter bson.D, document interface{}) (bool, error)
```

Implements the \[backend.NoSQLCollection\] interface

<a name="MongoCollection.UpsertID"></a>
### func \(\*MongoCollection\) [UpsertID](<https://github.com/blueprint-uservices/blueprint/blob/main/runtime/plugins/mongodb/nosqldb.go#L167>)

```go
func (mc *MongoCollection) UpsertID(ctx context.Context, id primitive.ObjectID, document interface{}) (bool, error)
```

Implements the \[backend.NoSQLCollection\] interface

<a name="MongoCursor"></a>
## type [MongoCursor](<https://github.com/blueprint-uservices/blueprint/blob/main/runtime/plugins/mongodb/nosqldb.go#L188-L190>)

Implements the \[backend.NoSQLCursor\] interface as a client\-wrapper to the Cursor returned by a mongodb server

```go
type MongoCursor struct {
    // contains filtered or unexported fields
}
```

<a name="MongoCursor.All"></a>
### func \(\*MongoCursor\) [All](<https://github.com/blueprint-uservices/blueprint/blob/main/runtime/plugins/mongodb/nosqldb.go#L210>)

```go
func (mr *MongoCursor) All(ctx context.Context, objs interface{}) error
```

Implements the \[backend.NoSQLCursor\] interface

<a name="MongoCursor.One"></a>
### func \(\*MongoCursor\) [One](<https://github.com/blueprint-uservices/blueprint/blob/main/runtime/plugins/mongodb/nosqldb.go#L193>)

```go
func (mr *MongoCursor) One(ctx context.Context, obj interface{}) (bool, error)
```

Implements the \[backend.NoSQLCursor\] interface

<a name="MongoDB"></a>
## type [MongoDB](<https://github.com/blueprint-uservices/blueprint/blob/main/runtime/plugins/mongodb/nosqldb.go#L16-L18>)

Implements the \[backend.NoSQLDatabase\] interface as a client\-wrapper to a mongodb server.

```go
type MongoDB struct {
    // contains filtered or unexported fields
}
```

<a name="NewMongoDB"></a>
### func [NewMongoDB](<https://github.com/blueprint-uservices/blueprint/blob/main/runtime/plugins/mongodb/nosqldb.go#L27>)

```go
func NewMongoDB(ctx context.Context, addr string) (*MongoDB, error)
```

Instantiates a new MongoDB client\-wrapper instance which connects to a mongodb server running at \`addr\`. REQUIRED: A mongodb server should be running at \`addr\`

<a name="MongoDB.GetCollection"></a>
### func \(\*MongoDB\) [GetCollection](<https://github.com/blueprint-uservices/blueprint/blob/main/runtime/plugins/mongodb/nosqldb.go#L40>)

```go
func (md *MongoDB) GetCollection(ctx context.Context, db_name string, collectionName string) (backend.NoSQLCollection, error)
```

Implements the \[backend.NoSQLDatabase\] interface

Generated by [gomarkdoc](<https://github.com/princjef/gomarkdoc>)