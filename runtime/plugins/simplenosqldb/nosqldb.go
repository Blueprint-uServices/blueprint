package simplenosqldb

import (
	"context"
	"fmt"
	"reflect"

	"gitlab.mpi-sws.org/cld/blueprint/runtime/core/backend"
	"gitlab.mpi-sws.org/cld/blueprint/runtime/plugins/simplenosqldb/query"
	"go.mongodb.org/mongo-driver/bson"
)

/*
Simple implementations of the NoSQLDB Interfaces from runtime/core/backend
*/
type (
	SimpleNoSQLDB struct {
		collections map[string]map[string]*SimpleCollection
	}

	SimpleCollection struct {
		items []bson.D
	}

	SimpleCursor struct {
		results []bson.D
	}
)

func NewSimpleNoSQLDB() (*SimpleNoSQLDB, error) {
	db := &SimpleNoSQLDB{}
	db.collections = make(map[string]map[string]*SimpleCollection)
	return db, nil
}

func (impl *SimpleNoSQLDB) GetCollection(ctx context.Context, db_name string, collection_name string) (backend.NoSQLCollection, error) {
	db, dbExists := impl.collections[db_name]
	if !dbExists {
		db = make(map[string]*SimpleCollection)
		impl.collections[db_name] = db
	}

	collection, collectionExists := db[collection_name]
	if !collectionExists {
		collection = &SimpleCollection{}
		db[collection_name] = collection
	}

	return collection, nil
}

func (c *SimpleCursor) Decode(ctx context.Context, obj interface{}) error {
	if len(c.results) == 0 {
		return backend.SetZero(obj)
	} else {
		return fromBson(c.results[0], obj)
	}
}

func CopyResult(src []bson.D, dst any) error {
	dst_ptr := reflect.ValueOf(dst)
	if dst_ptr.Kind() != reflect.Pointer || dst_ptr.IsNil() {
		return fmt.Errorf("unable to copy result to type %v", reflect.TypeOf(dst))
	}
	dst_val := reflect.Indirect(dst_ptr)

	if dst_val.Kind() == reflect.Slice {
		new_dst := reflect.MakeSlice(dst_val.Type(), len(src), len(src))
		for i, src_elem := range src {
			dst_elem := new_dst.Index(i).Addr().Interface()
			err := fromBson(src_elem, dst_elem)
			if err != nil {
				return err
			}
		}
		dst_val.Set(new_dst)
		return nil
	}

	return fmt.Errorf("cannot copy slice results to non-slice %v", dst)
}

func (c *SimpleCursor) All(ctx context.Context, obj interface{}) error {
	return CopyResult(c.results, obj)
}

func (db *SimpleCollection) InsertOne(ctx context.Context, document interface{}) error {
	if d, isD := document.(bson.D); isD {
		db.items = append(db.items, d)
	} else {
		d, err := toBson(document)
		if err != nil {
			return err
		}
		db.items = append(db.items, d)
	}
	return nil
}

func (db *SimpleCollection) InsertMany(ctx context.Context, documents []interface{}) error {
	for _, d := range documents {
		err := db.InsertOne(ctx, d)
		if err != nil {
			return err
		}
	}
	return nil
}

func (db *SimpleCollection) FindOne(ctx context.Context, filter bson.D, projection ...bson.D) (backend.NoSQLCursor, error) {
	query, err := query.ParseFilter(filter)
	if err != nil {
		return nil, err
	}
	cursor := &SimpleCursor{}
	for _, item := range db.items {
		if query.Apply(item) {
			cursor.results = append(cursor.results, item)
			break
		}
	}
	return cursor, nil
}

var verbose = false

func (db *SimpleCollection) FindMany(ctx context.Context, filter bson.D, projection ...bson.D) (backend.NoSQLCursor, error) {
	query, err := query.ParseFilter(filter)
	if err != nil {
		return nil, err
	}
	if verbose {
		fmt.Printf("---- FindMany\n%v\n", query)
	}
	cursor := &SimpleCursor{}
	for _, item := range db.items {
		if query.Apply(item) {
			cursor.results = append(cursor.results, item)
			if verbose {
				fmt.Printf("MATCH: %v\n", item)
			}
		} else {
			if verbose {
				fmt.Printf("       %v\n", item)
			}
		}
	}
	return cursor, nil
}

func (db *SimpleCollection) DeleteOne(ctx context.Context, filter bson.D) error {
	query, err := query.ParseFilter(filter)
	if err != nil {
		return err
	}
	for i, item := range db.items {
		if query.Apply(item) {
			db.items = append(db.items[:i], db.items[i+1:]...)
			return nil
		}
	}
	return nil
}

func (db *SimpleCollection) DeleteMany(ctx context.Context, filter bson.D) error {
	query, err := query.ParseFilter(filter)
	if err != nil {
		return err
	}
	copyrangebegin := 0
	newitems := make([]bson.D, 0, len(db.items))
	for i, item := range db.items {
		if query.Apply(item) {
			if i > copyrangebegin {
				newitems = append(newitems, db.items[copyrangebegin:i]...)
			}
			copyrangebegin = i + 1
		}
	}
	if copyrangebegin < len(db.items) {
		newitems = append(newitems, db.items[copyrangebegin:len(db.items)]...)
	}
	db.items = newitems
	return nil

}

func (db *SimpleCollection) UpdateOne(ctx context.Context, filter bson.D, update bson.D) error {
	filterOp, err := query.ParseFilter(filter)
	if err != nil {
		return err
	}
	updateOp, err := query.ParseUpdate(update)
	if err != nil {
		return err
	}

	for i := range db.items {
		if filterOp.Apply(db.items[i]) {
			return updateOp.Apply(&db.items[i])
		}
	}
	return nil
}

func (db *SimpleCollection) UpdateMany(ctx context.Context, filter bson.D, update bson.D) error {
	filterOp, err := query.ParseFilter(filter)
	if err != nil {
		return err
	}
	updateOp, err := query.ParseUpdate(update)
	if err != nil {
		return err
	}
	if verbose {
		fmt.Printf("---- UpdateMany\n")
		fmt.Printf(" MATCH:  %v\n", filterOp)
		fmt.Printf(" UPDATE: %v\n", updateOp)
	}

	for i := range db.items {
		if filterOp.Apply(db.items[i]) {
			if verbose {
				fmt.Printf("UPDATING: %v\n", db.items[i])
			}
			err := updateOp.Apply(&db.items[i])
			if err != nil {
				return err
			}
			if verbose {
				fmt.Printf("      --> %v\n", db.items[i])
			}
		} else {
			if verbose {
				fmt.Printf("          %v\n", db.items[i])
			}
		}
	}
	return nil
}

func (db *SimpleCollection) ReplaceOne(ctx context.Context, filter bson.D, replacement interface{}) error {
	query, err := query.ParseFilter(filter)
	if err != nil {
		return err
	}
	for i, item := range db.items {
		if query.Apply(item) {
			db.items[i], err = toBson(replacement)
			return err
		}
	}
	return nil
}

func (db *SimpleCollection) ReplaceMany(ctx context.Context, filter bson.D, replacements ...interface{}) error {
	query, err := query.ParseFilter(filter)
	if err != nil {
		return nil
	}
	for i, j := 0, 0; i < len(replacements) && j < len(db.items); j++ {
		if query.Apply(db.items[j]) {
			db.items[j], err = toBson(replacements[i])
			if err != nil {
				return err
			}
			i++
		}
	}
	return nil
}

func toBson(document any) (bson.D, error) {
	bytes, err := bson.Marshal(document)
	if err != nil {
		return nil, err
	}
	var d bson.D
	err = bson.Unmarshal(bytes, &d)
	return d, err
}

func fromBson(d bson.D, dst any) error {
	bytes, err := bson.Marshal(d)
	if err != nil {
		return err
	}
	return bson.Unmarshal(bytes, dst)
}
