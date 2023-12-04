package mongodb

import (
	"context"
	"errors"

	"gitlab.mpi-sws.org/cld/blueprint/runtime/core/backend"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// * constructor
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

type MongoDB struct {
	client *mongo.Client
}

func (md *MongoDB) GetCollection(ctx context.Context, db_name string, collectionName string) (backend.NoSQLCollection, error) {
	db := md.client.Database(db_name)
	coll := db.Collection(collectionName)
	return &MongoCollection{
		collection: coll,
	}, nil
}

type MongoCollection struct {
	collection *mongo.Collection
}

func (mc *MongoCollection) DeleteOne(ctx context.Context, filter bson.D) error {

	_, err := mc.collection.DeleteOne(ctx, filter)

	return err

}
func (mc *MongoCollection) DeleteMany(ctx context.Context, filter bson.D) error {
	_, err := mc.collection.DeleteMany(ctx, filter)
	return err
}
func (mc *MongoCollection) InsertOne(ctx context.Context, document interface{}) error {
	_, err := mc.collection.InsertOne(ctx, document)

	return err
}
func (mc *MongoCollection) InsertMany(ctx context.Context, documents []interface{}) error {
	_, err := mc.collection.InsertMany(ctx, documents)

	return err
}

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

// * not sure about the `update` parameter and its conversion
func (mc *MongoCollection) UpdateOne(ctx context.Context, filter bson.D, update bson.D) (int, error) {
	result, err := mc.collection.UpdateOne(ctx, filter, update)
	return int(result.ModifiedCount), err
}

func (mc *MongoCollection) UpdateMany(ctx context.Context, filter bson.D, update bson.D) (int, error) {
	result, err := mc.collection.UpdateMany(ctx, filter, update)
	if result == nil || err != nil {
		return 0, err
	} else {
		return int(result.ModifiedCount), nil
	}
}

func (mc *MongoCollection) Upsert(ctx context.Context, filter bson.D, document interface{}) (bool, error) {
	opts := options.Replace().SetUpsert(true)
	result, err := mc.collection.ReplaceOne(ctx, filter, document, opts)
	if result == nil || err != nil {
		return false, err
	} else {
		return result.MatchedCount == 1, nil
	}
}

func (mc *MongoCollection) UpsertID(ctx context.Context, id primitive.ObjectID, document interface{}) (bool, error) {
	filter := bson.D{{"_id", id}}
	return mc.Upsert(ctx, filter, document)
}

func (mc *MongoCollection) ReplaceOne(ctx context.Context, filter bson.D, replacement interface{}) (int, error) {
	result, err := mc.collection.ReplaceOne(ctx, filter, replacement)
	return int(result.MatchedCount), err
}

func (mc *MongoCollection) ReplaceMany(ctx context.Context, filter bson.D, replacements ...interface{}) (int, error) {
	return 0, errors.New("ReplaceMany not implemented")
}

type MongoCursor struct {
	underlyingResult interface{}
}

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

func (mr *MongoCursor) All(ctx context.Context, objs interface{}) error {
	//add other types of results from mongo that are Cursors here
	switch v := mr.underlyingResult.(type) {
	case *mongo.Cursor:
		return v.All(context.TODO(), objs)
	default:
		return errors.New("result does not return a Cursor")
	}
}
