// Package mongodb implements a cleint interface to a mongodb server that supports MongoDB's query and update API.
package mongodb

import (
	"context"
	"errors"

	"github.com/blueprint-uservices/blueprint/runtime/core/backend"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Implements the [backend.NoSQLDatabase] interface as a client-wrapper to a mongodb server.
type MongoDB struct {
	client *mongo.Client
}

// Implements the [backend.NoSQLCollection] interface as a client-wrapper to a mongodb server
type MongoCollection struct {
	collection *mongo.Collection
}

// Instantiates a new MongoDB client-wrapper instance which connects to a mongodb server running at `addr`.
// REQUIRED: A mongodb server should be running at `addr`
func NewMongoDB(ctx context.Context, addr string) (*MongoDB, error) {
	clientOptions := options.Client().ApplyURI("mongodb://" + addr)
	client, err := mongo.Connect(ctx, clientOptions)

	if err != nil {
		return nil, err
	}
	return &MongoDB{
		client: client,
	}, nil
}

// Implements the [backend.NoSQLDatabase] interface
func (md *MongoDB) GetCollection(ctx context.Context, db_name string, collectionName string) (backend.NoSQLCollection, error) {
	db := md.client.Database(db_name)
	coll := db.Collection(collectionName)
	return &MongoCollection{
		collection: coll,
	}, nil
}

// Implements the [backend.NoSQLCollection] interface
func (mc *MongoCollection) DeleteOne(ctx context.Context, filter bson.D) error {

	_, err := mc.collection.DeleteOne(ctx, filter)

	return err

}

// Implements the [backend.NoSQLCollection] interface
func (mc *MongoCollection) DeleteMany(ctx context.Context, filter bson.D) error {
	_, err := mc.collection.DeleteMany(ctx, filter)
	return err
}

// Implements the [backend.NoSQLCollection] interface
func (mc *MongoCollection) InsertOne(ctx context.Context, document interface{}) error {
	_, err := mc.collection.InsertOne(ctx, document)

	return err
}

// Implements the [backend.NoSQLCollection] interface
func (mc *MongoCollection) InsertMany(ctx context.Context, documents []interface{}) error {
	_, err := mc.collection.InsertMany(ctx, documents)

	return err
}

// Implements the [backend.NoSQLCollection] interface
func (mc *MongoCollection) FindOne(ctx context.Context, filter bson.D, projection ...bson.D) (backend.NoSQLCursor, error) {

	withProjection := false

	if len(projection) > 1 {
		return nil, errors.New("Invalid projection parameter!")
	} else if len(projection) == 1 {
		withProjection = true
	}

	var singleResult *mongo.SingleResult
	if withProjection {
		opts := options.FindOne().SetProjection(projection[0])
		singleResult = mc.collection.FindOne(ctx, filter, opts)
	} else {
		singleResult = mc.collection.FindOne(ctx, filter)
	}

	err := singleResult.Err()
	if err == nil || err == mongo.ErrNoDocuments {
		return &MongoCursor{underlyingResult: singleResult}, nil
	} else {
		return nil, err
	}
}

// Implements the [backend.NoSQLCollection] interface
func (mc *MongoCollection) FindMany(ctx context.Context, filter bson.D, projection ...bson.D) (backend.NoSQLCursor, error) {

	withProjection := false

	if len(projection) > 1 {
		return nil, errors.New("Invalid projection parameter!")
	} else if len(projection) == 1 {
		withProjection = true
	}

	var cursor *mongo.Cursor
	var err error
	if withProjection {
		opts := options.Find().SetProjection(projection[0])
		cursor, err = mc.collection.Find(ctx, filter, opts)
	} else {
		cursor, err = mc.collection.Find(ctx, filter)
	}

	if err != nil {
		return nil, err
	}
	if cursor.Err() != nil {
		return nil, cursor.Err()
	}

	return &MongoCursor{underlyingResult: cursor}, nil
}

// Implements the [backend.NoSQLCollection] interface
func (mc *MongoCollection) UpdateOne(ctx context.Context, filter bson.D, update bson.D) (int, error) {
	result, err := mc.collection.UpdateOne(ctx, filter, update)
	if result == nil {
		return 0, err
	} else {
		return int(result.ModifiedCount), err
	}
}

// Implements the [backend.NoSQLCollection] interface
func (mc *MongoCollection) UpdateMany(ctx context.Context, filter bson.D, update bson.D) (int, error) {
	result, err := mc.collection.UpdateMany(ctx, filter, update)
	if result == nil {
		return 0, err
	} else {
		return int(result.ModifiedCount), err
	}
}

// Implements the [backend.NoSQLCollection] interface
func (mc *MongoCollection) Upsert(ctx context.Context, filter bson.D, document interface{}) (bool, error) {
	update := bson.D{{"$set", document}}
	opts := options.Update().SetUpsert(true)
	result, err := mc.collection.UpdateOne(ctx, filter, update, opts)
	if result == nil {
		return false, err
	} else {
		return result.MatchedCount == 1, err
	}
}

// Implements the [backend.NoSQLCollection] interface
func (mc *MongoCollection) UpsertID(ctx context.Context, id primitive.ObjectID, document interface{}) (bool, error) {
	filter := bson.D{{"_id", id}}
	return mc.Upsert(ctx, filter, document)
}

// Implements the [backend.NoSQLCollection] interface
func (mc *MongoCollection) ReplaceOne(ctx context.Context, filter bson.D, replacement interface{}) (int, error) {
	result, err := mc.collection.ReplaceOne(ctx, filter, replacement)
	if result == nil {
		return 0, err
	} else {
		return int(result.MatchedCount), err
	}
}

// Implements the [backend.NoSQLCollection] interface
func (mc *MongoCollection) ReplaceMany(ctx context.Context, filter bson.D, replacements ...interface{}) (int, error) {
	return 0, errors.New("ReplaceMany not implemented")
}

// Implements the [backend.NoSQLCursor] interface as a client-wrapper to the Cursor returned by a mongodb server
type MongoCursor struct {
	underlyingResult interface{}
}

// Implements the [backend.NoSQLCursor] interface
func (mr *MongoCursor) One(ctx context.Context, obj interface{}) (bool, error) {
	//add other types of results from mongo that have a Decode method here
	switch v := mr.underlyingResult.(type) {
	case *mongo.SingleResult:
		if v.Err() == nil {
			return true, v.Decode(obj)
		} else if v.Err() == mongo.ErrNoDocuments {
			return false, nil
		} else {
			return false, v.Err()
		}
	default:
		return false, errors.New("result has no decode method")
	}
}

// Implements the [backend.NoSQLCursor] interface
func (mr *MongoCursor) All(ctx context.Context, objs interface{}) error {
	//add other types of results from mongo that are Cursors here
	switch v := mr.underlyingResult.(type) {
	case *mongo.Cursor:
		return v.All(ctx, objs)
	default:
		return errors.New("result does not return a Cursor")
	}
}
