package backend

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type NoSQLDatabase interface {
	/*
		A NoSQLDatabse implementation might distinguish between databases and collections,
		or might not have those concepts.
	*/
	GetCollection(ctx context.Context, db_name string, collection_name string) (NoSQLCollection, error)
}

type NoSQLCursor interface {
	// Copies one result into the target pointer.
	// If there are no results, returns false; otherwise returns true.
	// Returns an error if obj is not a compatible type.
	One(ctx context.Context, obj interface{}) (bool, error)

	// Copies all results into the target pointer.
	// obj must be a pointer to a slice type.
	// Returns the number of results copied.
	// Returns an error if obj is not a compatible type.
	All(ctx context.Context, obj interface{}) error //similar logic to Decode, but for multiple documents
}

type NoSQLCollection interface {
	// Deletes the first document that matches filter
	//
	// We use the same filter semantics as mongodb
	// https://www.mongodb.com/docs/manual/tutorial/query-documents/
	DeleteOne(ctx context.Context, filter bson.D) error

	// Deletes all documents that match filter
	//
	// We use the same filter semantics as mongodb
	// https://www.mongodb.com/docs/manual/tutorial/query-documents/
	DeleteMany(ctx context.Context, filter bson.D) error

	// Inserts the document into the collection.
	InsertOne(ctx context.Context, document interface{}) error

	// Inserts all provided documents into the collection
	InsertMany(ctx context.Context, documents []interface{}) error

	// Finds a document that matches filter.
	//
	// We use the same filter semantics as mongodb
	// https://www.mongodb.com/docs/manual/tutorial/query-documents/
	//
	// Projections are optional and behave with mongodb semantics.
	FindOne(ctx context.Context, filter bson.D, projection ...bson.D) (NoSQLCursor, error)

	// Finds all documents that match the filter.
	//
	// We use the same filter semantics as mongodb
	// https://www.mongodb.com/docs/manual/tutorial/query-documents/
	//
	// Projections are optional and behave with mongodb semantics.
	FindMany(ctx context.Context, filter bson.D, projection ...bson.D) (NoSQLCursor, error) // Result is not a slice -> it is an object we can use to retrieve documents using res.All().

	// Applies the provided update to the first document that matches filter
	//
	// We use the same filter semantics as mongodb
	// https://www.mongodb.com/docs/manual/tutorial/query-documents/
	//
	// We use the same update operators as mongodb
	// https://www.mongodb.com/docs/manual/reference/method/db.collection.update/
	//
	// Returns the number of updated documents (0 or 1)
	UpdateOne(ctx context.Context, filter bson.D, update bson.D) (int, error)

	// Applies the provided update to all documents that match the filter
	//
	// We use the same filter semantics as mongodb
	// https://www.mongodb.com/docs/manual/tutorial/query-documents/
	//
	// We use the same update operators as mongodb
	// https://www.mongodb.com/docs/manual/reference/method/db.collection.update/
	//
	// Returns the number of updated documents (>= 0)
	UpdateMany(ctx context.Context, filter bson.D, update bson.D) (int, error)

	// Attempts to find a document in the collection that matches the filter.
	// If a match is found, replaces the existing document with the provided document.
	// If a match is not found, document is inserted into the collection.
	// Returns true if an existing document was updated; false otherwise
	Upsert(ctx context.Context, filter bson.D, document interface{}) (bool, error)

	// Attempts to match a document in the collection with "_id" = id.
	// If a match is found, replaces the existing document with the provided document.
	// If a match is not found, document is inserted into the collection.
	// Returns true if an existing document was updated; false otherwise
	//
	// This method requires that document has an "_id" field in its BSON representation.
	// If document is a golang struct, the standard way to do this is to tag a field as follows:
	//     ID   primitive.ObjectID `bson:"_id"`
	UpsertID(ctx context.Context, id primitive.ObjectID, document interface{}) (bool, error)

	// Replaces the first document that matches filter with the replacement document.
	//
	// We use the same filter semantics as mongodb
	// https://www.mongodb.com/docs/manual/tutorial/query-documents/
	//
	// Returns the number of replaced documents (0 or 1)
	ReplaceOne(ctx context.Context, filter bson.D, replacement interface{}) (int, error)

	// Replaces all documents that match filter with the replacement documents.
	//
	// We use the same filter semantics as mongodb
	// https://www.mongodb.com/docs/manual/tutorial/query-documents/
	//
	// Returns the number of replaced documents.
	ReplaceMany(ctx context.Context, filter bson.D, replacements ...interface{}) (int, error)
}
