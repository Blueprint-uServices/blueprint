package leaf

import (
	ctxx "context"
	"fmt"

	"github.com/blueprint-uservices/blueprint/runtime/core/backend"
	"go.mongodb.org/mongo-driver/bson"
	"go.opentelemetry.io/otel/metric"
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
	Counter    metric.Int64Counter
	logger     backend.Logger
}

func NewLeafServiceImpl(ctx ctxx.Context, cache backend.Cache, db backend.NoSQLDatabase) (*LeafServiceImpl, error) {
	collection, err := db.GetCollection(ctx, "leafdb", "leafcollection")
	if err != nil {
		return nil, err
	}
	meter, err := backend.Meter(ctx, "leafService")
	if err != nil {
		return nil, err
	}
	counter, err := meter.Int64Counter("num_requests")
	if err != nil {
		return nil, err
	}
	logger := backend.GetLogger()
	return &LeafServiceImpl{Cache: cache, Collection: collection, Counter: counter, logger: logger}, nil
}

func (l *LeafServiceImpl) HelloNothing(ctx ctxx.Context) error {
	l.Counter.Add(ctx, 1)
	ctx, _ = l.logger.Info(ctx, "hello nothing!")
	return nil
}

func (l *LeafServiceImpl) HelloInt(ctx ctxx.Context, a int64) (int64, error) {
	l.Counter.Add(ctx, 1)
	ctx, _ = l.logger.Info(ctx, "hello")
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
	l.Counter.Add(ctx, 1)
	return obj, nil
}

func (l *LeafServiceImpl) HelloMate(ctx ctxx.Context, a int, b int32, c string, d map[string]LeafObject, elems []string, elems2 []NestedLeafObject) (string, []string, int32, int, map[string]LeafObject, error) {
	l.Counter.Add(ctx, 1)
	return c, elems, b, a, d, nil
}

func (l *LeafServiceImpl) NonServiceFunction() int64 {
	return 3
}
