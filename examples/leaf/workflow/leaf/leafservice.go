package leaf

import (
	ctxx "context"
	"fmt"

	"gitlab.mpi-sws.org/cld/blueprint/runtime/core/backend"
	"go.mongodb.org/mongo-driver/bson"
)

type MyInt int64

type NestedLeafObject struct {
	Key   string
	Value string
	Props []string
}

type LeafObject struct {
	ID    int64
	Name  string
	Count int
	Props map[string]NestedLeafObject
}

type LeafService interface {
	HelloNothing(ctx ctxx.Context) error
	HelloInt(ctx ctxx.Context, a int64) (int64, error)
	HelloObject(ctx ctxx.Context, obj *LeafObject) (*LeafObject, error)
	HelloMate(ctx ctxx.Context, a int, b int32, c string, d map[string]LeafObject, elems []string, elems2 []NestedLeafObject) (string, []string, int32, int, map[string]LeafObject, error)
}

type LeafServiceImpl struct {
	LeafService
	Cache      backend.Cache
	Collection backend.NoSQLCollection
}

func NewLeafServiceImpl(ctx ctxx.Context, cache backend.Cache, db backend.NoSQLDatabase) (*LeafServiceImpl, error) {
	collection, err := db.GetCollection(ctx, "leafdb", "leafcollection")
	if err != nil {
		return nil, err
	}
	return &LeafServiceImpl{Cache: cache, Collection: collection}, nil
}

func (l *LeafServiceImpl) HelloNothing(ctx ctxx.Context) error {
	fmt.Println("hello nothing!")
	return nil
}

func (l *LeafServiceImpl) HelloInt(ctx ctxx.Context, a int64) (int64, error) {
	fmt.Println("hello")
	l.Cache.Put(ctx, "helloint", a)
	var b int64
	l.Cache.Get(ctx, "helloint", &b)

	filter := bson.D{{"id", 7}}
	cursor, err := l.Collection.FindMany(ctx, filter)
	if err != nil {
		return 0, err
	}

	var objs []LeafObject
	err = cursor.All(ctx, &objs)
	if err != nil {
		return 0, err
	}

	if len(objs) == 0 {
		obj := LeafObject{ID: 7, Name: "MyObject"}
		err = l.Collection.InsertOne(ctx, obj)
		if err != nil {
			return 0, err
		}
		objs = append(objs, obj)
	}
	fmt.Printf("%v\n", objs[0])

	update := bson.D{{"$inc", bson.D{{"count", 1}}}}
	l.Collection.UpdateOne(ctx, filter, update)

	return int64(objs[0].Count), nil
}

func (l *LeafServiceImpl) HelloObject(ctx ctxx.Context, obj *LeafObject) (*LeafObject, error) {
	return obj, nil
}

func (l *LeafServiceImpl) HelloMate(ctx ctxx.Context, a int, b int32, c string, d map[string]LeafObject, elems []string, elems2 []NestedLeafObject) (string, []string, int32, int, map[string]LeafObject, error) {
	return c, elems, b, a, d, nil
}

func (l *LeafServiceImpl) NonServiceFunction() int64 {
	return 3
}
