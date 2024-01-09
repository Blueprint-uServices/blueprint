package simplenosqldb_test

import (
	"context"
	"flag"
	"fmt"
	"strings"
	"testing"

	"github.com/blueprint-uservices/blueprint/runtime/core/backend"
	"github.com/blueprint-uservices/blueprint/runtime/plugins/mongodb"
	"github.com/blueprint-uservices/blueprint/runtime/plugins/simplenosqldb"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Packaging struct {
	Length int64
	Width  int64
	Kind   string
}

type Tea struct {
	Type      string
	Rating    int
	Vendor    []string `bson:"vendor,omitempty" json:"vendor,omitempty"`
	Packaging Packaging
	Sizes     []int32
}

var teas = []Tea{
	{Type: "Masala", Rating: 10, Vendor: []string{"A", "C"}, Packaging: Packaging{Length: 5, Width: 10, Kind: "Paper"}, Sizes: []int32{4}},
	{Type: "English Breakfast", Rating: 6, Sizes: []int32{4, 8, 16}},
	{Type: "Oolong", Rating: 7, Vendor: []string{"C"}, Sizes: []int32{8, 16}},
	{Type: "Assam", Rating: 5, Packaging: Packaging{Length: 8, Width: 5, Kind: "Cardboard"}, Sizes: []int32{16}},
	{Type: "Earl Grey", Rating: 8, Vendor: []string{"A", "B"}, Sizes: []int32{32}},
}

var newtea = Tea{Type: "Scottish Breakfast", Rating: 11, Sizes: []int32{4, 5, 6, 7, 8}}

type TeaCollection struct {
	Name string
	Teas []Tea
}

var teacollections = []TeaCollection{
	{Name: "cool collection", Teas: []Tea{teas[0], teas[1]}},
	{Name: "fun collection", Teas: []Tea{teas[3], teas[4]}},
}

var dbtype = flag.String("db", "simple", "Type of nosqldb to use.  simple or mongodb.  default to simple")
var collectionid = 0

func getDB(t *testing.T) (context.Context, backend.NoSQLDatabase) {
	ctx := context.Background()

	var db backend.NoSQLDatabase
	var err error
	if *dbtype == "mongodb" || *dbtype == "mongo" {
		db, err = mongodb.NewMongoDB(ctx, "localhost:27017")
	} else if *dbtype == "" || *dbtype == "simple" {
		simplenosqldb.SetVerbose(true)
		db, err = simplenosqldb.NewSimpleNoSQLDB(ctx)
	} else {
		t.Fatalf("Error: unknown db %v; expected simple or mongodb\n", *dbtype)
	}
	require.NoError(t, err)

	return ctx, db
}

func MakeTestDB(t *testing.T) (context.Context, backend.NoSQLCollection) {
	ctx, db := getDB(t)
	coll, err := db.GetCollection(ctx, "testdb", fmt.Sprintf("testcollection%v", collectionid))
	collectionid += 1
	require.NoError(t, err)

	var docs []interface{}
	for _, t := range teas {
		docs = append(docs, t)
	}
	err = coll.InsertMany(ctx, docs)
	require.NoError(t, err)
	return ctx, coll
}

func MakeTestDB2(t *testing.T) (context.Context, backend.NoSQLCollection) {
	ctx, db := getDB(t)
	coll, err := db.GetCollection(ctx, "testdb", fmt.Sprintf("testcollection%v", collectionid))
	collectionid += 1
	require.NoError(t, err)

	var docs []interface{}
	for _, t := range teacollections {
		docs = append(docs, t)
	}
	err = coll.InsertMany(ctx, docs)
	require.NoError(t, err)
	return ctx, coll
}

func TestPullNested(t *testing.T) {
	ctx, db := MakeTestDB2(t)

	{
		update := bson.D{{"$pull", bson.D{{"teas", bson.D{{"packaging", bson.D{{"length", 5}, {"width", 10}, {"kind", "Paper"}}}}}}}}
		_, err := db.UpdateMany(ctx, bson.D{}, update)
		require.NoError(t, err)

		filter := bson.D{{"name", "cool collection"}}
		cursor, err := db.FindOne(ctx, filter)
		require.NoError(t, err)

		var newcollection TeaCollection
		exists, err := cursor.One(ctx, &newcollection)
		require.NoError(t, err)
		require.True(t, exists)
		require.Len(t, newcollection.Teas, 1)
		require.Equal(t, teas[1], newcollection.Teas[0])
	}
}

func TestPullNested2(t *testing.T) {
	ctx, db := MakeTestDB2(t)

	{
		update := bson.D{{"$pull", bson.D{{"teas", bson.D{{"packaging.width", 10}}}}}}
		_, err := db.UpdateMany(ctx, bson.D{}, update)
		require.NoError(t, err)

		filter := bson.D{{"name", "cool collection"}}
		cursor, err := db.FindOne(ctx, filter)
		require.NoError(t, err)

		var newcollection TeaCollection
		exists, err := cursor.One(ctx, &newcollection)
		require.NoError(t, err)
		require.True(t, exists)
		require.Len(t, newcollection.Teas, 1)
		require.Equal(t, teas[1], newcollection.Teas[0])
	}
}

func TestGetAll(t *testing.T) {
	ctx, db := MakeTestDB(t)

	filter := bson.D{}
	cursor, err := db.FindMany(ctx, filter)
	require.NoError(t, err)

	getteas := []Tea{}
	err = cursor.All(ctx, &getteas)
	require.NoError(t, err)
	require.Len(t, getteas, 5)
	require.ElementsMatch(t, teas, getteas)
}

func TestGetType(t *testing.T) {
	ctx, db := MakeTestDB(t)

	for _, tea := range teas {
		filter := bson.D{{"type", tea.Type}}
		cursor, err := db.FindMany(ctx, filter)
		require.NoError(t, err)

		getteas := []Tea{}
		err = cursor.All(ctx, &getteas)
		require.NoError(t, err)

		require.Equal(t, 1, len(getteas))
		require.Equal(t, tea.Type, getteas[0].Type)

		filter = bson.D{{"rating", tea.Rating}}
		cursor, err = db.FindMany(ctx, filter)
		require.NoError(t, err)

		getteas = []Tea{}
		err = cursor.All(ctx, &getteas)
		require.NoError(t, err)

		require.Equal(t, 1, len(getteas))
		require.Equal(t, tea.Rating, getteas[0].Rating)
	}

	cursor, err := db.FindMany(ctx, bson.D{{"type", "Masala"}})
	require.NoError(t, err)
	gotteas := []Tea{}
	err = cursor.All(ctx, &gotteas)
	require.NoError(t, err)
	require.Equal(t, gotteas[0], teas[0])
}

func TestGetMissingField(t *testing.T) {
	ctx, db := MakeTestDB(t)

	filter := bson.D{{"type", "Oolong"}, {"rating", 7}, {"type.something", "selecta"}}
	cursor, err := db.FindMany(ctx, filter)
	require.NoError(t, err)

	getteas := []Tea{}
	err = cursor.All(ctx, &getteas)
	require.NoError(t, err)

	require.Empty(t, getteas)
}

func TestCmp(t *testing.T) {
	ctx, db := MakeTestDB(t)

	{
		filter := bson.D{{"rating", bson.D{{"$gt", 5}}}}
		cursor, err := db.FindMany(ctx, filter)
		require.NoError(t, err)

		getteas := []Tea{}
		err = cursor.All(ctx, &getteas)
		require.NoError(t, err)

		require.Len(t, getteas, 4)
		require.Equal(t, teas[0], getteas[0])
		require.Equal(t, teas[1], getteas[1])
		require.Equal(t, teas[2], getteas[2])
		require.Equal(t, teas[4], getteas[3])
	}
	{
		filter := bson.D{{"rating", bson.D{{"$gte", 5}}}}
		cursor, err := db.FindMany(ctx, filter)
		require.NoError(t, err)

		getteas := []Tea{}
		err = cursor.All(ctx, &getteas)
		require.NoError(t, err)

		require.Len(t, getteas, 5)
		require.ElementsMatch(t, teas, getteas)
	}
	{
		filter := bson.D{{"rating", bson.D{{"$lt", 5}}}}
		cursor, err := db.FindMany(ctx, filter)
		require.NoError(t, err)

		getteas := []Tea{}
		err = cursor.All(ctx, &getteas)
		require.NoError(t, err)

		require.Len(t, getteas, 0)
	}
	{
		filter := bson.D{{"rating", bson.D{{"$lte", 5}}}}
		cursor, err := db.FindMany(ctx, filter)
		require.NoError(t, err)

		getteas := []Tea{}
		err = cursor.All(ctx, &getteas)
		require.NoError(t, err)

		require.Len(t, getteas, 1)
		require.Equal(t, teas[3], getteas[0])
	}
	{
		filter := bson.D{{"rating", bson.D{{"$gt", 5}, {"$lt", 7}}}}
		cursor, err := db.FindMany(ctx, filter)
		require.NoError(t, err)

		getteas := []Tea{}
		err = cursor.All(ctx, &getteas)
		require.NoError(t, err)

		require.Len(t, getteas, 1)
		require.Equal(t, teas[1], getteas[0])
	}
}

func TestGetNotMissingField(t *testing.T) {
	ctx, db := MakeTestDB(t)

	filter := bson.D{{"blah.something", bson.D{{"$exists", false}}}}
	cursor, err := db.FindMany(ctx, filter)
	require.NoError(t, err)

	getteas := []Tea{}
	err = cursor.All(ctx, &getteas)

	require.Len(t, getteas, 5)
	require.ElementsMatch(t, teas, getteas)
}

func TestGetNotMissingField2(t *testing.T) {
	ctx, db := MakeTestDB(t)

	filter := bson.D{{"type.something", bson.D{{"$exists", false}}}}
	cursor, err := db.FindMany(ctx, filter)
	require.NoError(t, err)

	getteas := []Tea{}
	err = cursor.All(ctx, &getteas)

	require.Len(t, getteas, 5)
	require.ElementsMatch(t, teas, getteas)
}

func TestGetNotMissingField3(t *testing.T) {
	ctx, db := MakeTestDB(t)

	filter := bson.D{{"blank", bson.D{{"$exists", false}}}}
	cursor, err := db.FindMany(ctx, filter)
	require.NoError(t, err)

	getteas := []Tea{}
	err = cursor.All(ctx, &getteas)

	require.Len(t, getteas, 5)
	require.ElementsMatch(t, teas, getteas)
}

func TestGetNotMissingFieldExtra(t *testing.T) {
	ctx, db := MakeTestDB(t)

	filter := bson.D{{"type", "Oolong"}, {"rating", 7}, {"blah.something.cool", bson.D{{"$exists", false}}}}
	cursor, err := db.FindMany(ctx, filter)
	require.NoError(t, err)

	getteas := []Tea{}
	err = cursor.All(ctx, &getteas)
	require.NoError(t, err)

	require.Len(t, getteas, 1)
}

func TestCompositeFieldKey(t *testing.T) {
	ctx, db := MakeTestDB(t)

	filter := bson.D{{"packaging.kind", "Cardboard"}}
	cursor, err := db.FindMany(ctx, filter)
	require.NoError(t, err)

	getteas := []Tea{}
	err = cursor.All(ctx, &getteas)
	require.NoError(t, err)

	require.Len(t, getteas, 1)
	require.Equal(t, getteas[0].Type, "Assam")
}

func TestExactSubstructMatch(t *testing.T) {
	ctx, db := MakeTestDB(t)
	// Packaging: Packaging{Length: 5, Width: 10, Kind: "Paper"}

	{
		filter := bson.D{{"packaging", bson.D{{"length", 5}, {"width", 10}, {"kind", "Paper"}}}}
		cursor, err := db.FindMany(ctx, filter)
		require.NoError(t, err)

		getteas := []Tea{}
		err = cursor.All(ctx, &getteas)
		require.NoError(t, err)

		require.Len(t, getteas, 1)
		require.Equal(t, getteas[0].Type, "Masala")
	}
	{
		// Substruct must be exact, no missing fields
		filter := bson.D{{"packaging", bson.D{{"length", 5}, {"kind", "Paper"}}}}
		cursor, err := db.FindMany(ctx, filter)
		require.NoError(t, err)

		getteas := []Tea{}
		err = cursor.All(ctx, &getteas)
		require.NoError(t, err)

		require.Len(t, getteas, 0)
	}
	{
		filter := bson.D{{"packaging.length", 5}, {"packaging.width", 10}, {"packaging.kind", "Paper"}}
		cursor, err := db.FindMany(ctx, filter)
		require.NoError(t, err)

		getteas := []Tea{}
		err = cursor.All(ctx, &getteas)
		require.NoError(t, err)

		require.Len(t, getteas, 1)
		require.Equal(t, getteas[0].Type, "Masala")
	}
	{
		filter := bson.D{{"packaging.length", 5}, {"packaging.width", 10}}
		cursor, err := db.FindMany(ctx, filter)
		require.NoError(t, err)

		getteas := []Tea{}
		err = cursor.All(ctx, &getteas)
		require.NoError(t, err)

		require.Len(t, getteas, 1)
		require.Equal(t, getteas[0].Type, "Masala")
	}
}

func TestArrayEquals(t *testing.T) {
	ctx, db := MakeTestDB(t)

	{
		// Array match is not partial
		filter := bson.D{{"vendor", bson.A{"C"}}}
		cursor, err := db.FindMany(ctx, filter)
		require.NoError(t, err)

		getteas := []Tea{}
		err = cursor.All(ctx, &getteas)
		require.NoError(t, err)

		require.Len(t, getteas, 1)
		require.Equal(t, getteas[0].Type, "Oolong")
	}
	{
		filter := bson.D{{"vendor", bson.A{"A", "C"}}}
		cursor, err := db.FindMany(ctx, filter)
		require.NoError(t, err)

		getteas := []Tea{}
		err = cursor.All(ctx, &getteas)
		require.NoError(t, err)

		require.Len(t, getteas, 1)
		require.Equal(t, getteas[0].Type, "Masala")
	}
	{
		// Array match requires exact order
		filter := bson.D{{"vendor", bson.A{"C", "A"}}}
		cursor, err := db.FindMany(ctx, filter)
		require.NoError(t, err)

		getteas := []Tea{}
		err = cursor.All(ctx, &getteas)
		require.NoError(t, err)

		require.Len(t, getteas, 0)
	}
	{
		// Array match requires exact order
		filter := bson.D{{"vendor.1", "C"}}
		cursor, err := db.FindMany(ctx, filter)
		require.NoError(t, err)

		getteas := []Tea{}
		err = cursor.All(ctx, &getteas)
		require.NoError(t, err)

		require.Len(t, getteas, 1)
		require.Equal(t, getteas[0].Type, "Masala")
	}
}

func TestArrayContains(t *testing.T) {
	ctx, db := MakeTestDB(t)

	{
		// Should match Teas with any size == 4
		filter := bson.D{{"sizes", bson.D{{"$eq", 4}}}}
		cursor, err := db.FindMany(ctx, filter)
		require.NoError(t, err)

		getteas := []Tea{}
		err = cursor.All(ctx, &getteas)
		require.NoError(t, err)

		require.Len(t, getteas, 2)
		require.ElementsMatch(t, getteas, teas[0:2])
	}

	{
		// Should match Teas with any size == 4
		filter := bson.D{{"sizes", 4}}
		cursor, err := db.FindMany(ctx, filter)
		require.NoError(t, err)

		getteas := []Tea{}
		err = cursor.All(ctx, &getteas)
		require.NoError(t, err)

		require.Len(t, getteas, 2)
		require.ElementsMatch(t, getteas, teas[0:2])
	}

	{
		// Should match Teas with any size <= 8
		filter := bson.D{{"sizes", bson.D{{"$lt", 8}}}}
		cursor, err := db.FindMany(ctx, filter)
		require.NoError(t, err)

		getteas := []Tea{}
		err = cursor.All(ctx, &getteas)
		require.NoError(t, err)

		require.Len(t, getteas, 2)
		require.Equal(t, getteas[0].Type, "Masala")
		require.Equal(t, getteas[1].Type, "English Breakfast")
	}

	{
		filter := bson.D{{"sizes", bson.D{{"$gt", 8}}}}
		cursor, err := db.FindMany(ctx, filter)
		require.NoError(t, err)

		getteas := []Tea{}
		err = cursor.All(ctx, &getteas)
		require.NoError(t, err)

		require.Len(t, getteas, 4)
		require.ElementsMatch(t, getteas, teas[1:])
	}

	{
		filter := bson.D{{"sizes", bson.D{{"$gte", 8}, {"$lte", 8}}}}
		cursor, err := db.FindMany(ctx, filter)
		require.NoError(t, err)

		getteas := []Tea{}
		err = cursor.All(ctx, &getteas)
		require.NoError(t, err)

		require.Len(t, getteas, 2)
		require.ElementsMatch(t, teas[1:3], getteas)
	}

	{
		filter := bson.D{{"sizes", bson.D{{"$in", bson.A{16, 32}}}}}
		cursor, err := db.FindMany(ctx, filter)
		require.NoError(t, err)

		getteas := []Tea{}
		err = cursor.All(ctx, &getteas)
		require.NoError(t, err)

		require.Len(t, getteas, 4)
		require.ElementsMatch(t, teas[1:5], getteas)
	}

	{
		filter := bson.D{{"sizes", bson.D{{"$in", []int{16, 32}}}}}
		cursor, err := db.FindMany(ctx, filter)
		require.NoError(t, err)

		getteas := []Tea{}
		err = cursor.All(ctx, &getteas)
		require.NoError(t, err)

		require.Len(t, getteas, 4)
		require.ElementsMatch(t, teas[1:5], getteas)
	}

	fmt.Println(dbstring(ctx, db, t))

	{
		filter := bson.D{{"$nor", bson.A{bson.D{{"sizes", bson.D{{"$in", bson.A{16, 32}}}}}}}}
		cursor, err := db.FindMany(ctx, filter)
		require.NoError(t, err)

		getteas := []Tea{}
		err = cursor.All(ctx, &getteas)
		require.NoError(t, err)

		require.Len(t, getteas, 1)
		require.ElementsMatch(t, teas[0:1], getteas)
	}

	{
		filter := bson.D{{"sizes", bson.D{{"$in", bson.A{16, 32}}, {"$gt", 20}}}}
		cursor, err := db.FindMany(ctx, filter)
		require.NoError(t, err)

		getteas := []Tea{}
		err = cursor.All(ctx, &getteas)
		require.NoError(t, err)

		require.Len(t, getteas, 1)
		require.Equal(t, teas[4], getteas[0])
	}

	{
		filter := bson.D{{"$and", bson.A{
			bson.D{{"sizes", bson.D{{"$gt", 20}}}},
			bson.D{{"$nor", bson.A{bson.D{{"sizes", bson.D{{"$in", bson.A{16, 32}}}}}}}},
		}}}
		// filter := bson.D{{"sizes", bson.D{{"$nin", bson.A{16, 32}}, {"$gt", 20}}}}
		cursor, err := db.FindMany(ctx, filter)
		require.NoError(t, err)

		getteas := []Tea{}
		err = cursor.All(ctx, &getteas)
		require.NoError(t, err)

		require.Len(t, getteas, 0)
	}
}

func TestAll(t *testing.T) {
	ctx, db := MakeTestDB(t)

	{
		filter := bson.D{{"sizes", bson.D{{"$all", bson.A{4, 8}}}}}
		cursor, err := db.FindMany(ctx, filter)
		require.NoError(t, err)

		getteas := []Tea{}
		err = cursor.All(ctx, &getteas)
		require.NoError(t, err)

		require.Len(t, getteas, 1)
		require.Equal(t, getteas[0].Type, "English Breakfast")
	}

	{
		filter := bson.D{{"sizes", bson.D{{"$all", bson.A{4, 8, 16, 32}}}}}
		cursor, err := db.FindMany(ctx, filter)
		require.NoError(t, err)

		getteas := []Tea{}
		err = cursor.All(ctx, &getteas)
		require.NoError(t, err)

		require.Len(t, getteas, 0)
	}

	{
		filter := bson.D{{"sizes", bson.D{{"$all", bson.A{4}}}}}
		cursor, err := db.FindMany(ctx, filter)
		require.NoError(t, err)

		getteas := []Tea{}
		err = cursor.All(ctx, &getteas)
		require.NoError(t, err)

		require.Len(t, getteas, 2)
		require.ElementsMatch(t, getteas, teas[0:2])
	}
}

func TestElemMatch(t *testing.T) {
	ctx, db := MakeTestDB(t)

	{
		filter := bson.D{{"sizes", bson.D{{"$elemMatch", bson.D{{"nothing", bson.D{{"$gt", 4}}}}}}}}
		cursor, err := db.FindMany(ctx, filter)
		require.NoError(t, err)

		getteas := []Tea{}
		err = cursor.All(ctx, &getteas)
		require.NoError(t, err)

		require.Len(t, getteas, 0)
	}

	ctx, db = MakeTestDB2(t)

	{
		filter := bson.D{{"teas.type", "Masala"}}
		cursor, err := db.FindMany(ctx, filter)
		require.NoError(t, err)

		getteas := []TeaCollection{}
		err = cursor.All(ctx, &getteas)
		require.NoError(t, err)

		require.Len(t, getteas, 1)
	}

	{
		filter := bson.D{{"teas", bson.D{{"$elemMatch", bson.D{{"type", "Masala"}}}}}}
		cursor, err := db.FindMany(ctx, filter)
		require.NoError(t, err)

		getteas := []TeaCollection{}
		err = cursor.All(ctx, &getteas)
		require.NoError(t, err)

		require.Len(t, getteas, 1)
	}

	{
		filter := bson.D{{"teas", bson.D{{"$elemMatch", bson.D{{"$or", bson.A{bson.D{{"type", "Masala"}}}}}}}}}
		cursor, err := db.FindMany(ctx, filter)
		require.NoError(t, err)

		getteas := []TeaCollection{}
		err = cursor.All(ctx, &getteas)
		require.NoError(t, err)

		require.Len(t, getteas, 1)
		require.Equal(t, getteas[0], teacollections[0])
	}

	{
		filter := bson.D{{"teas", bson.D{{"$elemMatch", bson.D{{"$or", bson.A{bson.D{{"type", "Assam"}}}}}}}}}
		cursor, err := db.FindMany(ctx, filter)
		require.NoError(t, err)

		getteas := []TeaCollection{}
		err = cursor.All(ctx, &getteas)
		require.NoError(t, err)

		require.Len(t, getteas, 1)
		require.Equal(t, getteas[0], teacollections[1])
	}

	{
		filter := bson.D{{"teas", bson.D{{"$elemMatch", bson.D{{"$or", bson.A{bson.D{{"type", "Masala"}}, bson.D{{"type", "English Breakfast"}}}}}}}}}
		cursor, err := db.FindMany(ctx, filter)
		require.NoError(t, err)

		getteas := []TeaCollection{}
		err = cursor.All(ctx, &getteas)
		require.NoError(t, err)

		require.Len(t, getteas, 1)
	}

	{
		filter := bson.D{{"teas", bson.D{{"$elemMatch", bson.D{{"sizes", bson.D{{"$gt", 20}}}}}}}}
		cursor, err := db.FindMany(ctx, filter)
		require.NoError(t, err)

		getteas := []TeaCollection{}
		err = cursor.All(ctx, &getteas)
		require.NoError(t, err)

		require.Len(t, getteas, 1)
		require.Equal(t, getteas[0], teacollections[1])
	}

	{
		filter := bson.D{{"teas", bson.D{{"$elemMatch", bson.D{{"$or", bson.A{bson.D{{"type", "Masala"}}, bson.D{{"sizes", bson.D{{"$gt", 20}}}}}}}}}}}
		cursor, err := db.FindMany(ctx, filter)
		require.NoError(t, err)

		getteas := []TeaCollection{}
		err = cursor.All(ctx, &getteas)
		require.NoError(t, err)

		require.Len(t, getteas, 2)
		require.Equal(t, getteas[0], teacollections[0])
		require.Equal(t, getteas[1], teacollections[1])
	}
}

func TestDeleteOne(t *testing.T) {
	ctx, db := MakeTestDB(t)

	{
		filter := bson.D{{"rating", bson.D{{"$gt", 5}}}}

		for i := 4; i >= 0; i-- {

			{
				cursor, err := db.FindMany(ctx, filter)
				require.NoError(t, err)

				getteas := []Tea{}
				err = cursor.All(ctx, &getteas)
				require.NoError(t, err)
				require.Len(t, getteas, i)
			}

			{
				err := db.DeleteOne(ctx, filter)
				require.NoError(t, err)
			}
		}

		{
			getteas := []Tea{}
			filter := bson.D{}
			cursor, err := db.FindMany(ctx, filter)
			require.NoError(t, err)

			err = cursor.All(ctx, &getteas)
			require.NoError(t, err)
			require.Len(t, getteas, 1)
			require.Equal(t, getteas[0], teas[3])
		}
	}
}

func TestDeleteMany(t *testing.T) {
	ctx, db := MakeTestDB(t)

	{
		filter := bson.D{{"rating", bson.D{{"$gt", 5}}}}

		{
			cursor, err := db.FindMany(ctx, filter)
			require.NoError(t, err)

			getteas := []Tea{}
			err = cursor.All(ctx, &getteas)
			require.NoError(t, err)
			require.Len(t, getteas, 4)
		}

		{
			err := db.DeleteMany(ctx, filter)
			require.NoError(t, err)
		}

		{
			getteas := []Tea{}
			cursor, err := db.FindMany(ctx, filter)
			require.NoError(t, err)

			err = cursor.All(ctx, &getteas)
			require.NoError(t, err)
			require.Len(t, getteas, 0)
		}

		{
			getteas := []Tea{}
			filter := bson.D{}
			cursor, err := db.FindMany(ctx, filter)
			require.NoError(t, err)

			err = cursor.All(ctx, &getteas)
			require.NoError(t, err)
			require.Len(t, getteas, 1)
			require.Equal(t, getteas[0], teas[3])
		}
	}
}

type SimpleTea struct {
	Type string
}

func TestConvertToOtherStruct(t *testing.T) {

	ctx, db := MakeTestDB(t)

	filter := bson.D{{"type", "English Breakfast"}}
	cursor, err := db.FindMany(ctx, filter)
	require.NoError(t, err)

	getteas := []SimpleTea{}
	err = cursor.All(ctx, &getteas)
	require.NoError(t, err)

	require.Equal(t, 1, len(getteas))
	require.Equal(t, "English Breakfast", getteas[0].Type)
}

func TestSimpleUpdate(t *testing.T) {
	ctx, db := MakeTestDB(t)

	{
		filter := bson.D{{"type", "English Breakfast"}}
		cursor, err := db.FindMany(ctx, filter)
		require.NoError(t, err)

		getteas := []Tea{}
		err = cursor.All(ctx, &getteas)
		require.NoError(t, err)

		require.Equal(t, 1, len(getteas))
		require.Equal(t, "English Breakfast", getteas[0].Type)
		require.Equal(t, 6, getteas[0].Rating)
	}

	{
		filter := bson.D{{"type", "English Breakfast"}}
		update := bson.D{{"$set", bson.D{{"rating", 9}}}}

		updated, err := db.UpdateMany(ctx, filter, update)
		require.Equal(t, 1, updated)
		require.NoError(t, err)

		cursor, err := db.FindMany(ctx, filter)
		require.NoError(t, err)

		getteas := []Tea{}
		err = cursor.All(ctx, &getteas)
		require.NoError(t, err)

		require.Equal(t, 1, len(getteas))
		require.Equal(t, "English Breakfast", getteas[0].Type)
		require.Equal(t, 9, getteas[0].Rating)
	}

}

type TeaFact struct {
	Info   string
	Source string
}

type ExtendedTea struct {
	Type    string
	Message string
	Fact    TeaFact
	Facts   []TeaFact
}

func TestUpdateAddNewField(t *testing.T) {
	ctx, db := MakeTestDB(t)

	{
		filter := bson.D{{"type", "English Breakfast"}}
		cursor, err := db.FindMany(ctx, filter)
		require.NoError(t, err)

		getteas := []ExtendedTea{}
		err = cursor.All(ctx, &getteas)
		require.NoError(t, err)

		require.Equal(t, 1, len(getteas))
		require.Equal(t, "English Breakfast", getteas[0].Type)
		require.Equal(t, "", getteas[0].Message)
	}

	{
		filter := bson.D{{"type", "English Breakfast"}}
		update := bson.D{{"$set", bson.D{{"message", "The best tea."}}}}

		updated, err := db.UpdateMany(ctx, filter, update)
		require.Equal(t, 1, updated)
		require.NoError(t, err)

		cursor, err := db.FindMany(ctx, filter)
		require.NoError(t, err)

		getteas := []ExtendedTea{}
		err = cursor.All(ctx, &getteas)
		require.NoError(t, err)

		require.Equal(t, 1, len(getteas))
		require.Equal(t, "English Breakfast", getteas[0].Type)
		require.Equal(t, "The best tea.", getteas[0].Message)
	}

	{
		filter := bson.D{{"type", "English Breakfast"}}
		update := bson.D{{"$set", bson.D{{"fact.info", "Further information."}, {"fact.source", "jon"}}}}

		updated, err := db.UpdateMany(ctx, filter, update)
		require.Equal(t, 1, updated)
		require.NoError(t, err)

		cursor, err := db.FindMany(ctx, filter)
		require.NoError(t, err)

		getteas := []ExtendedTea{}
		err = cursor.All(ctx, &getteas)
		require.NoError(t, err)

		require.Equal(t, 1, len(getteas))
		require.Equal(t, "English Breakfast", getteas[0].Type)
		require.Equal(t, "Further information.", getteas[0].Fact.Info)
		require.Equal(t, "jon", getteas[0].Fact.Source)
	}
}

// func TestUpdateArray(t *testing.T) {
// 	ctx, db := MakeTestDB(t)

// 	{
// 		filter := bson.D{{"type", "English Breakfast"}}
// 		cursor, err := db.FindMany(ctx, filter)
// 		require.NoError(t, err)

// 		getteas := []ExtendedTea{}
// 		err = cursor.All(ctx, &getteas)
// 		require.NoError(t, err)

// 		require.Equal(t, 1, len(getteas))
// 		require.Equal(t, "English Breakfast", getteas[0].Type)
// 		require.Equal(t, 0, len(getteas[0].Facts))
// 	}

// 	{
// 		filter := bson.D{{"type", "English Breakfast"}}
// 		update := bson.D{{"$push", bson.D{{"facts.3.info", "The best tea."}}}}

// 		updated, err := db.UpdateMany(ctx, filter, update)
// 		require.Equal(t, 1, updated)
// 		require.NoError(t, err)

// 		fmt.Println("before error:")
// 		fmt.Println(dbstring(ctx, db, t))

// 		cursor, err := db.FindMany(ctx, filter)
// 		require.NoError(t, err)

// 		getteas := []ExtendedTea{}
// 		err = cursor.All(ctx, &getteas)
// 		require.NoError(t, err)

// 		require.Equal(t, 1, len(getteas))
// 		require.Equal(t, "English Breakfast", getteas[0].Type)
// 		require.Equal(t, 4, len(getteas[0].Facts))
// 		require.Equal(t, "The best tea.", getteas[0].Facts[3].Info)
// 	}

// 	{
// 		filter := bson.D{{"type", "English Breakfast"}}
// 		update := bson.D{{"$set", bson.D{{"facts.3.info", "no fact."}}}}

// 		updated, err := db.UpdateMany(ctx, filter, update)
// 		require.Equal(t, 1, updated)
// 		require.NoError(t, err)

// 		cursor, err := db.FindMany(ctx, filter)
// 		require.NoError(t, err)

// 		getteas := []ExtendedTea{}
// 		err = cursor.All(ctx, &getteas)
// 		require.NoError(t, err)

// 		require.Equal(t, 1, len(getteas))
// 		require.Equal(t, "English Breakfast", getteas[0].Type)
// 		require.Equal(t, 4, len(getteas[0].Facts))
// 		require.Equal(t, "no fact.", getteas[0].Facts[3].Info)
// 	}

// 	{
// 		filter := bson.D{{"type", "English Breakfast"}}
// 		update := bson.D{{"$set", bson.D{{"facts", bson.A{}}}}}

// 		updated, err := db.UpdateMany(ctx, filter, update)
// 		require.Equal(t, 1, updated)
// 		require.NoError(t, err)

// 		cursor, err := db.FindMany(ctx, filter)
// 		require.NoError(t, err)

// 		getteas := []ExtendedTea{}
// 		err = cursor.All(ctx, &getteas)
// 		require.NoError(t, err)

// 		require.Equal(t, 1, len(getteas))
// 		require.Equal(t, "English Breakfast", getteas[0].Type)
// 		require.Equal(t, 0, len(getteas[0].Facts))
// 	}
// }

func TestUnset(t *testing.T) {
	ctx, db := MakeTestDB(t)

	{
		filter := bson.D{{"rating", 6}}
		cursor, err := db.FindMany(ctx, filter)
		require.NoError(t, err)

		getteas := []ExtendedTea{}
		err = cursor.All(ctx, &getteas)
		require.NoError(t, err)

		require.Equal(t, 1, len(getteas))
		require.Equal(t, "English Breakfast", getteas[0].Type)
		require.Equal(t, 0, len(getteas[0].Facts))
	}

	{
		filter := bson.D{{"rating", 6}, {"sizes", 8}}
		update := bson.D{{"$unset", bson.D{{"type", ""}}}}

		updated, err := db.UpdateMany(ctx, filter, update)
		require.Equal(t, 1, updated)
		require.NoError(t, err)

		cursor, err := db.FindMany(ctx, filter)
		require.NoError(t, err)

		getteas := []ExtendedTea{}
		err = cursor.All(ctx, &getteas)
		require.NoError(t, err)

		require.Equal(t, 1, len(getteas))
		require.Equal(t, "", getteas[0].Type)
	}

	{
		filter := bson.D{{"rating", 6}, {"sizes", 8}}
		update := bson.D{{"$unset", bson.D{{"sizes.1", ""}}}}

		updated, err := db.UpdateMany(ctx, filter, update)
		require.Equal(t, 1, updated)
		require.NoError(t, err)

		cursor, err := db.FindMany(ctx, filter)
		require.NoError(t, err)

		getteas := []ExtendedTea{}
		err = cursor.All(ctx, &getteas)
		require.NoError(t, err)

		require.Equal(t, 0, len(getteas))
	}
}

func TestInc(t *testing.T) {
	ctx, db := MakeTestDB(t)

	{
		filter := bson.D{{"type", "English Breakfast"}}
		cursor, err := db.FindMany(ctx, filter)
		require.NoError(t, err)

		getteas := []Tea{}
		err = cursor.All(ctx, &getteas)
		require.NoError(t, err)

		require.Equal(t, 1, len(getteas))
		require.Equal(t, "English Breakfast", getteas[0].Type)
		require.Equal(t, 6, getteas[0].Rating)
	}

	{
		filter := bson.D{{"type", "English Breakfast"}}

		for i := 1; i < 5; i++ {
			update := bson.D{{"$inc", bson.D{{"rating", 1}}}}

			updated, err := db.UpdateMany(ctx, filter, update)
			require.Equal(t, 1, updated)
			require.NoError(t, err)

			cursor, err := db.FindMany(ctx, filter)
			require.NoError(t, err)

			getteas := []Tea{}
			err = cursor.All(ctx, &getteas)
			require.NoError(t, err)

			require.Equal(t, 1, len(getteas))
			require.Equal(t, "English Breakfast", getteas[0].Type)
			require.Equal(t, 6+i, getteas[0].Rating)
		}
	}
}

func TestIncArray(t *testing.T) {
	ctx, db := MakeTestDB(t)

	{
		filter := bson.D{{"type", "English Breakfast"}}
		cursor, err := db.FindMany(ctx, filter)
		require.NoError(t, err)

		getteas := []Tea{}
		err = cursor.All(ctx, &getteas)
		require.NoError(t, err)

		require.Equal(t, 1, len(getteas))
		require.Equal(t, "English Breakfast", getteas[0].Type)
		require.Equal(t, 3, len(getteas[0].Sizes))
	}

	{
		filter := bson.D{{"type", "English Breakfast"}}

		for i := 1; i < 5; i++ {
			update := bson.D{{"$inc", bson.D{{"sizes.5", 2}}}}

			updated, err := db.UpdateMany(ctx, filter, update)
			require.Equal(t, 1, updated)
			require.NoError(t, err)

			cursor, err := db.FindMany(ctx, filter)
			require.NoError(t, err)

			getteas := []Tea{}
			err = cursor.All(ctx, &getteas)
			require.NoError(t, err)

			require.Equal(t, 1, len(getteas))
			require.Equal(t, "English Breakfast", getteas[0].Type)
			require.Equal(t, 6, len(getteas[0].Sizes))
			require.Equal(t, int32(2*i), getteas[0].Sizes[5])
		}
	}
}

func TestPush(t *testing.T) {
	ctx, db := MakeTestDB(t)

	{
		filter := bson.D{{"type", "Masala"}}
		cursor, err := db.FindMany(ctx, filter)
		require.NoError(t, err)

		getteas := []Tea{}
		err = cursor.All(ctx, &getteas)
		require.NoError(t, err)

		require.Equal(t, 1, len(getteas))
		require.Equal(t, "Masala", getteas[0].Type)
		require.ElementsMatch(t, getteas[0].Sizes, []int32{4})
	}

	{
		filter := bson.D{{"type", "Masala"}}
		update := bson.D{{"$push", bson.D{{"sizes", 6}}}}

		updated, err := db.UpdateMany(ctx, filter, update)
		require.Equal(t, 1, updated)
		require.NoError(t, err)

		cursor, err := db.FindMany(ctx, filter)
		require.NoError(t, err)

		getteas := []Tea{}
		err = cursor.All(ctx, &getteas)
		require.NoError(t, err)

		require.Equal(t, 1, len(getteas))
		require.Equal(t, "Masala", getteas[0].Type)
		require.ElementsMatch(t, getteas[0].Sizes, []int32{4, 6})
	}

	{
		filter := bson.D{{"type", "English Breakfast"}}
		update := bson.D{{"$push", bson.D{{"sizes", 12}, {"Vendor", "newvendor"}}}}

		updated, err := db.UpdateMany(ctx, filter, update)
		require.Equal(t, 1, updated)
		require.NoError(t, err)

		cursor, err := db.FindMany(ctx, filter)
		require.NoError(t, err)

		getteas := []Tea{}
		err = cursor.All(ctx, &getteas)
		require.NoError(t, err)

		require.Equal(t, 1, len(getteas))
		require.Equal(t, "English Breakfast", getteas[0].Type)
		require.ElementsMatch(t, getteas[0].Sizes, []int32{4, 8, 16, 12})
		require.ElementsMatch(t, getteas[0].Vendor, []string{"newvendor"})
	}
}

func TestPull(t *testing.T) {
	ctx, db := MakeTestDB(t)

	{
		update := bson.D{{"$pull", bson.D{{"sizes", 8}}}}
		_, err := db.UpdateMany(ctx, bson.D{}, update)
		require.NoError(t, err)

		filter := bson.D{{"type", "English Breakfast"}}
		cursor, err := db.FindOne(ctx, filter)
		require.NoError(t, err)

		var tea Tea
		_, err = cursor.One(ctx, &tea)
		require.NoError(t, err)
		require.Equal(t, "English Breakfast", tea.Type)
		require.Len(t, tea.Sizes, 2)
		require.ElementsMatch(t, tea.Sizes, []int32{4, 16})
	}
}

type ObjectIDTest struct {
	ID   primitive.ObjectID `bson:"_id,omitempty"`
	Type string
}

func TestObjectIDGetsSet(t *testing.T) {
	ctx, db := MakeTestDB(t)

	{
		filter := bson.D{{"type", "English Breakfast"}}
		cursor, err := db.FindOne(ctx, filter)
		require.NoError(t, err)

		var id ObjectIDTest
		require.Equal(t, id.ID, primitive.NilObjectID)

		_, err = cursor.One(ctx, &id)
		require.NoError(t, err)
		require.NotEqual(t, id.ID, primitive.NilObjectID)
		require.Equal(t, id.Type, "English Breakfast")
	}
}

func TestUpdateByObjectID(t *testing.T) {
	ctx, db := MakeTestDB(t)

	{
		typeFilter := bson.D{{"type", "English Breakfast"}}
		cursor, err := db.FindOne(ctx, typeFilter)
		require.NoError(t, err)

		var id ObjectIDTest
		_, err = cursor.One(ctx, &id)
		require.NoError(t, err)
		require.Equal(t, id.Type, "English Breakfast")

		idFilter := bson.D{{"_id", id.ID}}
		cursor, err = db.FindOne(ctx, idFilter)
		require.NoError(t, err)

		_, err = cursor.One(ctx, &id)
		require.NoError(t, err)
		require.Equal(t, id.Type, "English Breakfast")

		updateTo := bson.D{{"$set", bson.D{{"type", "Scottish Breakfast"}}}}
		updated, err := db.UpdateOne(ctx, idFilter, updateTo)
		require.Equal(t, 1, updated)
		require.NoError(t, err)

		cursor, err = db.FindOne(ctx, idFilter)
		require.NoError(t, err)

		_, err = cursor.One(ctx, &id)
		require.NoError(t, err)
		require.Equal(t, id.Type, "Scottish Breakfast")

		cursor, err = db.FindOne(ctx, typeFilter)
		require.NoError(t, err)

		hasResult, err := cursor.One(ctx, &id)
		require.NoError(t, err)
		require.False(t, hasResult)
	}

}

func TestUpsertExisting(t *testing.T) {
	ctx, db := MakeTestDB(t)

	filterold := bson.D{{"type", teas[0].Type}}
	filternew := bson.D{{"type", newtea.Type}}

	{
		// Check the old tea exists using FindOne
		var tea Tea
		cursor, err := db.FindOne(ctx, filterold)
		require.NoError(t, err)

		hasResult, err := cursor.One(ctx, &tea)
		require.NoError(t, err)
		require.True(t, hasResult)
		require.Equal(t, teas[0], tea)
	}

	{
		// Check the new tea doesn't exist using FindOne
		var tea Tea
		cursor, err := db.FindOne(ctx, filternew)
		require.NoError(t, err)

		hasResult, err := cursor.One(ctx, &tea)
		require.NoError(t, err)
		require.False(t, hasResult)
	}

	{
		// Check the old tea exists using FindMany
		var newteas []Tea
		cursor, err := db.FindMany(ctx, filterold)
		require.NoError(t, err)

		err = cursor.All(ctx, &newteas)
		require.NoError(t, err)
		require.Len(t, newteas, 1)
		require.Equal(t, newteas[0], teas[0])
	}

	{
		// Check the new tea doesn't exist using FindMany
		var newteas []Tea
		cursor, err := db.FindMany(ctx, filternew)
		require.NoError(t, err)

		err = cursor.All(ctx, &newteas)
		require.NoError(t, err)
		require.Len(t, newteas, 0)
	}

	{
		// Upsert the existing tea
		updated, err := db.Upsert(ctx, filterold, newtea)
		require.NoError(t, err)
		require.True(t, updated)
	}

	{
		// Check the old tea no longer exists using FindOne
		var tea Tea
		cursor, err := db.FindOne(ctx, filterold)
		require.NoError(t, err)

		hasResult, err := cursor.One(ctx, &tea)
		require.NoError(t, err)
		require.False(t, hasResult)
	}

	{
		// Check the new tea now exists using FindOne
		var tea Tea
		cursor, err := db.FindOne(ctx, filternew)
		require.NoError(t, err)

		hasResult, err := cursor.One(ctx, &tea)
		require.NoError(t, err)
		require.True(t, hasResult)
		require.Equal(t, newtea, tea)
	}

	{
		// Check the old tea no longer exists using FindMany
		var newteas []Tea
		cursor, err := db.FindMany(ctx, filterold)
		require.NoError(t, err)

		err = cursor.All(ctx, &newteas)
		require.NoError(t, err)
		require.Len(t, newteas, 0)
	}

	{
		// Check the new tea now exists using FindMany
		var newteas []Tea
		cursor, err := db.FindMany(ctx, filternew)
		require.NoError(t, err)

		err = cursor.All(ctx, &newteas)
		require.NoError(t, err)
		require.Len(t, newteas, 1)
		require.Equal(t, newteas[0], newtea)
	}
}

func TestUpsertNew(t *testing.T) {
	ctx, db := MakeTestDB(t)

	filterold := bson.D{{"type", teas[0].Type}}
	filternew := bson.D{{"type", newtea.Type}}

	{
		// Check the old tea exists using FindOne
		var tea Tea
		cursor, err := db.FindOne(ctx, filterold)
		require.NoError(t, err)

		hasResult, err := cursor.One(ctx, &tea)
		require.NoError(t, err)
		require.True(t, hasResult)
		require.Equal(t, teas[0], tea)
	}

	{
		// Check the new tea doesn't exist using FindOne
		var tea Tea
		cursor, err := db.FindOne(ctx, filternew)
		require.NoError(t, err)

		hasResult, err := cursor.One(ctx, &tea)
		require.NoError(t, err)
		require.False(t, hasResult)
	}

	{
		// Check the old tea exists using FindMany
		var newteas []Tea
		cursor, err := db.FindMany(ctx, filterold)
		require.NoError(t, err)

		err = cursor.All(ctx, &newteas)
		require.NoError(t, err)
		require.Len(t, newteas, 1)
		require.Equal(t, newteas[0], teas[0])
	}

	{
		// Check the new tea doesn't exist using FindMany
		var newteas []Tea
		cursor, err := db.FindMany(ctx, filternew)
		require.NoError(t, err)

		err = cursor.All(ctx, &newteas)
		require.NoError(t, err)
		require.Len(t, newteas, 0)
	}

	{
		// Upsert the new tea
		updated, err := db.Upsert(ctx, filternew, newtea)
		require.NoError(t, err)
		require.False(t, updated)
	}

	{
		// Check the old tea still exists using FindOne
		var tea Tea
		cursor, err := db.FindOne(ctx, filterold)
		require.NoError(t, err)

		hasResult, err := cursor.One(ctx, &tea)
		require.NoError(t, err)
		require.True(t, hasResult)
		require.Equal(t, teas[0], tea)
	}

	{
		// Check the new tea now exists using FindOne
		var tea Tea
		cursor, err := db.FindOne(ctx, filternew)
		require.NoError(t, err)

		hasResult, err := cursor.One(ctx, &tea)
		require.NoError(t, err)
		require.True(t, hasResult)
		require.Equal(t, newtea, tea)
	}

	{
		// Check the old tea still exists using FindMany
		var newteas []Tea
		cursor, err := db.FindMany(ctx, filterold)
		require.NoError(t, err)

		err = cursor.All(ctx, &newteas)
		require.NoError(t, err)
		require.Len(t, newteas, 1)
		require.Equal(t, newteas[0], teas[0])
	}

	{
		// Check the new tea now exists using FindMany
		var newteas []Tea
		cursor, err := db.FindMany(ctx, filternew)
		require.NoError(t, err)

		err = cursor.All(ctx, &newteas)
		require.NoError(t, err)
		require.Len(t, newteas, 1)
		require.Equal(t, newteas[0], newtea)
	}
}

func TestAddToSet(t *testing.T) {
	ctx, db := MakeTestDB(t)

	masala := teas[0]
	filter := bson.D{{"type", masala.Type}}

	{
		// Check the tea exists using FindOne
		var tea Tea
		cursor, err := db.FindOne(ctx, filter)
		require.NoError(t, err)

		hasResult, err := cursor.One(ctx, &tea)
		require.NoError(t, err)
		require.True(t, hasResult)
		require.Equal(t, teas[0], tea)
	}

	{
		// Addtoset a new vendor
		update := bson.D{{"$addToSet", bson.D{{"vendor", "D"}}}}
		updated, err := db.UpdateOne(ctx, filter, update)
		require.NoError(t, err)
		require.Equal(t, 1, updated)

		// Check the vendors
		var tea Tea
		cursor, err := db.FindOne(ctx, filter)
		require.NoError(t, err)

		hasResult, err := cursor.One(ctx, &tea)
		require.NoError(t, err)
		require.True(t, hasResult)
		require.Equal(t, masala.Type, tea.Type)
		require.ElementsMatch(t, []string{"A", "C", "D"}, tea.Vendor)
	}

	{
		// Addtoset an existing vendor
		update := bson.D{{"$addToSet", bson.D{{"vendor", "A"}}}}
		updated, err := db.UpdateOne(ctx, filter, update)
		require.NoError(t, err)
		require.Equal(t, 1, updated)

		// Check the vendors
		var tea Tea
		cursor, err := db.FindOne(ctx, filter)
		require.NoError(t, err)

		hasResult, err := cursor.One(ctx, &tea)
		require.NoError(t, err)
		require.True(t, hasResult)
		require.Equal(t, masala.Type, tea.Type)
		require.ElementsMatch(t, []string{"A", "C", "D"}, tea.Vendor)
	}

	{
		// Addtoset an existing vendor
		update := bson.D{{"$addToSet", bson.D{{"vendor", "D"}}}}
		updated, err := db.UpdateOne(ctx, filter, update)
		require.NoError(t, err)
		require.Equal(t, 1, updated)

		// Check the vendors
		var tea Tea
		cursor, err := db.FindOne(ctx, filter)
		require.NoError(t, err)

		hasResult, err := cursor.One(ctx, &tea)
		require.NoError(t, err)
		require.True(t, hasResult)
		require.Equal(t, masala.Type, tea.Type)
		require.ElementsMatch(t, []string{"A", "C", "D"}, tea.Vendor)
	}

	{
		// Addtoset a new vendor
		update := bson.D{{"$addToSet", bson.D{{"vendor", "F"}}}}
		updated, err := db.UpdateOne(ctx, filter, update)
		require.NoError(t, err)
		require.Equal(t, 1, updated)

		// Check the vendors
		var tea Tea
		cursor, err := db.FindOne(ctx, filter)
		require.NoError(t, err)

		hasResult, err := cursor.One(ctx, &tea)
		require.NoError(t, err)
		require.True(t, hasResult)
		require.Equal(t, masala.Type, tea.Type)
		require.ElementsMatch(t, []string{"A", "C", "D", "F"}, tea.Vendor)
	}

	{
		// Addtoset an existing vendor
		update := bson.D{{"$addToSet", bson.D{{"vendor", "F"}}}}
		updated, err := db.UpdateOne(ctx, filter, update)
		require.NoError(t, err)
		require.Equal(t, 1, updated)

		// Check the vendors
		var tea Tea
		cursor, err := db.FindOne(ctx, filter)
		require.NoError(t, err)

		hasResult, err := cursor.One(ctx, &tea)
		require.NoError(t, err)
		require.True(t, hasResult)
		require.Equal(t, masala.Type, tea.Type)
		require.ElementsMatch(t, []string{"A", "C", "D", "F"}, tea.Vendor)
	}

	{
		// Addtoset an existing vendor
		update := bson.D{{"$addToSet", bson.D{{"vendor", "A"}}}}
		updated, err := db.UpdateOne(ctx, filter, update)
		require.NoError(t, err)
		require.Equal(t, 1, updated)

		// Check the vendors
		var tea Tea
		cursor, err := db.FindOne(ctx, filter)
		require.NoError(t, err)

		hasResult, err := cursor.One(ctx, &tea)
		require.NoError(t, err)
		require.True(t, hasResult)
		require.Equal(t, masala.Type, tea.Type)
		require.ElementsMatch(t, []string{"A", "C", "D", "F"}, tea.Vendor)
	}
}

func dbstring(ctx context.Context, db backend.NoSQLCollection, t *testing.T) string {
	before := simplenosqldb.SetVerbose(false)
	defer simplenosqldb.SetVerbose(before)
	cursor, err := db.FindMany(ctx, bson.D{})
	if err != nil {
		return fmt.Sprintf("unable to print db due to %v", err.Error())
	}
	var documents []bson.D
	err = cursor.All(ctx, &documents)
	if err != nil {
		return fmt.Sprintf("unable to print db due to %v", err.Error())
	}
	var docstrings []string
	for i := range documents {
		docstrings = append(docstrings, fmt.Sprintf("%v=%v", i, documents[i]))
	}
	return fmt.Sprintf("\nDocuments in DB:\n%v\n\n", strings.Join(docstrings, "\n"))
}
