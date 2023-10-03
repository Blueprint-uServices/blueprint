package backend

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
)

type NoSQLDatabase interface {
	/*
		A NoSQLDatabse implementation might distinguish between databases and collections,
		or might not have those concepts.
	*/
	GetCollection(ctx context.Context, db_name string, collection_name string) (NoSQLCollection, error)
}

type NoSQLCursor interface {
	Decode(ctx context.Context, obj interface{}) error
	All(ctx context.Context, obj interface{}) error //similar logic to Decode, but for multiple documents
}

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
