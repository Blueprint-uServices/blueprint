// Package simplenosqldb implements an in-memory NoSQLDB that supports a subset of MongoDB's query and update API.
//
// Only a small set of common basic filter and update operators are supported, but typically this is sufficient
// for most applications and enables writing service-level unit tests.
package simplenosqldb

import (
	"context"
	"fmt"
	"reflect"
	"strings"

	"github.com/blueprint-uservices/blueprint/runtime/core/backend"
	"github.com/blueprint-uservices/blueprint/runtime/plugins/simplenosqldb/query"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

/*
Simple implementations of the NoSQLDB Interfaces from runtime/core/backend
*/
type (
	// Implements the [backend.NoSQLDatabase] interface for a subset of MongoDB's query and update operators.
	//
	// Only a small set of common basic filter and update operators are supported, but typically this is sufficient
	// for most applications and enables writing service-level unit tests.
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

// Instantiate a new in-memory NoSQLDB
func NewSimpleNoSQLDB(ctx context.Context) (*SimpleNoSQLDB, error) {
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

func (c *SimpleCursor) One(ctx context.Context, obj interface{}) (bool, error) {
	if len(c.results) == 0 {
		return false, nil
	} else {
		return true, fromBson(c.results[0], obj)
	}
}

func copyResult(src []bson.D, dst any) error {
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
	return copyResult(c.results, obj)
}

func (db *SimpleCollection) InsertOne(ctx context.Context, document interface{}) error {
	d, isD := document.(bson.D)
	if !isD {
		var err error
		d, err = toBson(document)
		if err != nil {
			return err
		}
	}
	hasId := false
	for _, e := range d {
		if e.Key == "_id" {
			hasId = true
			break
		}
	}
	if !hasId {
		d = append(bson.D{{"_id", primitive.NewObjectID()}}, d...)
	}

	db.items = append(db.items, d)
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
	if verbose {
		fmt.Printf("---- FindOne\n%v\n", query)
	}
	cursor := &SimpleCursor{}
	for _, item := range db.items {
		if query.Apply(item) {
			cursor.results = append(cursor.results, item)
			if verbose {
				fmt.Printf("MATCH: %v\n", item)
			}
			break
		}
	}
	return cursor, nil
}

var verbose = false

// Enable or disable verbose logging; used for testing
func SetVerbose(enabled bool) bool {
	before := verbose
	verbose = enabled
	return before
}

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

func (db *SimpleCollection) UpdateOne(ctx context.Context, filter bson.D, update bson.D) (int, error) {
	filterOp, err := query.ParseFilter(filter)
	if err != nil {
		return 0, err
	}
	updateOp, err := query.ParseUpdate(update)
	if err != nil {
		return 0, err
	}

	if verbose {
		fmt.Printf("---- UpdateOne\n%v\n%v\n", filter, update)
	}

	for i := range db.items {
		if filterOp.Apply(db.items[i]) {
			if verbose {
				fmt.Printf("MATCH: %v\n", db.items[i])
			}
			return 1, updateOp.Apply(&db.items[i])
		} else {
			if verbose {
				fmt.Printf("      %v\n", db.items[i])
			}
		}
	}
	return 0, nil
}

func (db *SimpleCollection) UpdateMany(ctx context.Context, filter bson.D, update bson.D) (int, error) {
	filterOp, err := query.ParseFilter(filter)
	if err != nil {
		return 0, err
	}
	updateOp, err := query.ParseUpdate(update)
	if err != nil {
		return 0, err
	}
	if verbose {
		fmt.Printf("---- UpdateMany\n")
		fmt.Printf(" MATCH:  %v\n", filterOp)
		fmt.Printf(" UPDATE: %v\n", updateOp)
	}

	updated := 0
	for i := range db.items {
		if filterOp.Apply(db.items[i]) {
			if verbose {
				fmt.Printf("UPDATING: %v\n", db.items[i])
			}
			err := updateOp.Apply(&db.items[i])
			if err != nil {
				return updated, err
			}
			if verbose {
				fmt.Printf("      --> %v\n", db.items[i])
			}
			updated += 1
		} else {
			if verbose {
				fmt.Printf("          %v\n", db.items[i])
			}
		}
	}
	return updated, nil
}

func (db *SimpleCollection) Upsert(ctx context.Context, filter bson.D, document interface{}) (bool, error) {
	updatedCount, err := db.ReplaceOne(ctx, filter, document)
	if updatedCount == 1 || err != nil {
		return true, err
	}
	return false, db.InsertOne(ctx, document)
}

func (db *SimpleCollection) UpsertID(ctx context.Context, id primitive.ObjectID, document interface{}) (bool, error) {
	filter := bson.D{{"_id", id}}
	updated, err := db.Upsert(ctx, filter, document)
	if updated && verbose && err == nil {
		fmt.Printf("Upsert replaced existing %v\n", id)
	}
	return updated, err
}

func (db *SimpleCollection) ReplaceOne(ctx context.Context, filter bson.D, replacement interface{}) (int, error) {
	query, err := query.ParseFilter(filter)
	if err != nil {
		return 0, err
	}
	for i, item := range db.items {
		if query.Apply(item) {
			db.items[i], err = toBson(replacement)
			return 1, err
		}
	}
	return 0, nil
}

func (db *SimpleCollection) ReplaceMany(ctx context.Context, filter bson.D, replacements ...interface{}) (int, error) {
	query, err := query.ParseFilter(filter)
	if err != nil {
		return 0, nil
	}
	updateCount := 0
	for i := 0; updateCount < len(replacements) && i < len(db.items); i++ {
		if query.Apply(db.items[i]) {
			db.items[i], err = toBson(replacements[updateCount])
			if err != nil {
				return updateCount, err
			}
			updateCount++
		}
	}
	return updateCount, nil
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

func (db *SimpleCollection) String() string {
	var strs []string
	for i := range db.items {
		strs = append(strs, fmt.Sprintf("%v", db.items[i]))
	}
	return strings.Join(strs, "\n")
}
