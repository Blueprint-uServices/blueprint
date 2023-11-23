package simplenosqldb

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
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

type TeaCollection struct {
	Name string
	Teas []Tea
}

var teacollections = []TeaCollection{
	{Name: "cool collection", Teas: []Tea{teas[0], teas[1]}},
	{Name: "fun collection", Teas: []Tea{teas[3], teas[4]}},
}

func MakeTestDB(t *testing.T) (context.Context, *SimpleCollection) {
	verbose = true
	ctx := context.Background()
	db, err := NewSimpleNoSQLDB(ctx)
	assert.NoError(t, err)

	coll, err := db.GetCollection(ctx, "testdb", "testcollection")
	assert.NoError(t, err)

	var docs []interface{}
	for _, t := range teas {
		docs = append(docs, t)
	}
	err = coll.InsertMany(ctx, docs)
	assert.NoError(t, err)
	return ctx, coll.(*SimpleCollection)
}

func MakeTestDB2(t *testing.T) (context.Context, *SimpleCollection) {
	verbose = true
	ctx := context.Background()
	db, err := NewSimpleNoSQLDB(ctx)
	assert.NoError(t, err)

	coll, err := db.GetCollection(ctx, "testdb", "testcollection2")
	assert.NoError(t, err)

	var docs []interface{}
	for _, t := range teacollections {
		docs = append(docs, t)
	}
	err = coll.InsertMany(ctx, docs)
	assert.NoError(t, err)
	return ctx, coll.(*SimpleCollection)
}

func TestGetAll(t *testing.T) {
	ctx, db := MakeTestDB(t)

	filter := bson.D{}
	cursor, err := db.FindMany(ctx, filter)
	assert.NoError(t, err)

	getteas := []Tea{}
	err = cursor.All(ctx, &getteas)
	assert.NoError(t, err)
	assert.Len(t, getteas, 5)
	assert.ElementsMatch(t, teas, getteas)
}

func TestGetType(t *testing.T) {
	ctx, db := MakeTestDB(t)

	for _, tea := range teas {
		filter := bson.D{{"type", tea.Type}}
		cursor, err := db.FindMany(ctx, filter)
		assert.NoError(t, err)

		getteas := []Tea{}
		err = cursor.All(ctx, &getteas)
		assert.NoError(t, err)

		assert.Equal(t, 1, len(getteas))
		assert.Equal(t, tea.Type, getteas[0].Type)

		filter = bson.D{{"rating", tea.Rating}}
		cursor, err = db.FindMany(ctx, filter)
		assert.NoError(t, err)

		getteas = []Tea{}
		err = cursor.All(ctx, &getteas)
		assert.NoError(t, err)

		assert.Equal(t, 1, len(getteas))
		assert.Equal(t, tea.Rating, getteas[0].Rating)
	}

	cursor, err := db.FindMany(ctx, bson.D{{"type", "Masala"}})
	assert.NoError(t, err)
	gotteas := []Tea{}
	err = cursor.All(ctx, &gotteas)
	assert.NoError(t, err)
	assert.Equal(t, gotteas[0], teas[0])
}

func TestGetMissingField(t *testing.T) {
	ctx, db := MakeTestDB(t)

	filter := bson.D{{"type", "Oolong"}, {"rating", 7}, {"type.something", "selecta"}}
	cursor, err := db.FindMany(ctx, filter)
	assert.NoError(t, err)

	getteas := []Tea{}
	err = cursor.All(ctx, &getteas)
	assert.NoError(t, err)

	assert.Empty(t, getteas)
}

func TestCmp(t *testing.T) {
	ctx, db := MakeTestDB(t)

	{
		filter := bson.D{{"rating", bson.D{{"$gt", 5}}}}
		cursor, err := db.FindMany(ctx, filter)
		assert.NoError(t, err)

		getteas := []Tea{}
		err = cursor.All(ctx, &getteas)
		assert.NoError(t, err)

		assert.Len(t, getteas, 4)
		assert.Equal(t, teas[0], getteas[0])
		assert.Equal(t, teas[1], getteas[1])
		assert.Equal(t, teas[2], getteas[2])
		assert.Equal(t, teas[4], getteas[3])
	}
	{
		filter := bson.D{{"rating", bson.D{{"$gte", 5}}}}
		cursor, err := db.FindMany(ctx, filter)
		assert.NoError(t, err)

		getteas := []Tea{}
		err = cursor.All(ctx, &getteas)
		assert.NoError(t, err)

		assert.Len(t, getteas, 5)
		assert.ElementsMatch(t, teas, getteas)
	}
	{
		filter := bson.D{{"rating", bson.D{{"$lt", 5}}}}
		cursor, err := db.FindMany(ctx, filter)
		assert.NoError(t, err)

		getteas := []Tea{}
		err = cursor.All(ctx, &getteas)
		assert.NoError(t, err)

		assert.Len(t, getteas, 0)
	}
	{
		filter := bson.D{{"rating", bson.D{{"$lte", 5}}}}
		cursor, err := db.FindMany(ctx, filter)
		assert.NoError(t, err)

		getteas := []Tea{}
		err = cursor.All(ctx, &getteas)
		assert.NoError(t, err)

		assert.Len(t, getteas, 1)
		assert.Equal(t, teas[3], getteas[0])
	}
	{
		filter := bson.D{{"rating", bson.D{{"$gt", 5}, {"$lt", 7}}}}
		cursor, err := db.FindMany(ctx, filter)
		assert.NoError(t, err)

		getteas := []Tea{}
		err = cursor.All(ctx, &getteas)
		assert.NoError(t, err)

		assert.Len(t, getteas, 1)
		assert.Equal(t, teas[1], getteas[0])
	}
}

func TestGetNotMissingField(t *testing.T) {
	ctx, db := MakeTestDB(t)

	filter := bson.D{{"blah.something", bson.D{{"$exists", false}}}}
	cursor, err := db.FindMany(ctx, filter)
	assert.NoError(t, err)

	getteas := []Tea{}
	err = cursor.All(ctx, &getteas)

	assert.Len(t, getteas, 5)
	assert.ElementsMatch(t, teas, getteas)
}

func TestGetNotMissingField2(t *testing.T) {
	ctx, db := MakeTestDB(t)

	filter := bson.D{{"type.something", bson.D{{"$exists", false}}}}
	cursor, err := db.FindMany(ctx, filter)
	assert.NoError(t, err)

	getteas := []Tea{}
	err = cursor.All(ctx, &getteas)

	assert.Len(t, getteas, 5)
	assert.ElementsMatch(t, teas, getteas)
}

func TestGetNotMissingField3(t *testing.T) {
	ctx, db := MakeTestDB(t)

	filter := bson.D{{"blank", bson.D{{"$exists", false}}}}
	cursor, err := db.FindMany(ctx, filter)
	assert.NoError(t, err)

	getteas := []Tea{}
	err = cursor.All(ctx, &getteas)

	assert.Len(t, getteas, 5)
	assert.ElementsMatch(t, teas, getteas)
}

func TestGetNotMissingFieldExtra(t *testing.T) {
	ctx, db := MakeTestDB(t)

	filter := bson.D{{"type", "Oolong"}, {"rating", 7}, {"blah.something.cool", bson.D{{"$exists", false}}}}
	cursor, err := db.FindMany(ctx, filter)
	assert.NoError(t, err)

	getteas := []Tea{}
	err = cursor.All(ctx, &getteas)
	assert.NoError(t, err)

	assert.Len(t, getteas, 1)
}

func TestCompositeFieldKey(t *testing.T) {
	ctx, db := MakeTestDB(t)

	filter := bson.D{{"packaging.kind", "Cardboard"}}
	cursor, err := db.FindMany(ctx, filter)
	assert.NoError(t, err)

	getteas := []Tea{}
	err = cursor.All(ctx, &getteas)
	assert.NoError(t, err)

	assert.Len(t, getteas, 1)
	assert.Equal(t, getteas[0].Type, "Assam")
}

func TestExactSubstructMatch(t *testing.T) {
	ctx, db := MakeTestDB(t)
	// Packaging: Packaging{Length: 5, Width: 10, Kind: "Paper"}

	{
		filter := bson.D{{"packaging", bson.D{{"length", 5}, {"width", 10}, {"kind", "Paper"}}}}
		cursor, err := db.FindMany(ctx, filter)
		assert.NoError(t, err)

		getteas := []Tea{}
		err = cursor.All(ctx, &getteas)
		assert.NoError(t, err)

		assert.Len(t, getteas, 1)
		assert.Equal(t, getteas[0].Type, "Masala")
	}
	{
		// Substruct must be exact, no missing fields
		filter := bson.D{{"packaging", bson.D{{"length", 5}, {"kind", "Paper"}}}}
		cursor, err := db.FindMany(ctx, filter)
		assert.NoError(t, err)

		getteas := []Tea{}
		err = cursor.All(ctx, &getteas)
		assert.NoError(t, err)

		assert.Len(t, getteas, 0)
	}
	{
		filter := bson.D{{"packaging.length", 5}, {"packaging.width", 10}, {"packaging.kind", "Paper"}}
		cursor, err := db.FindMany(ctx, filter)
		assert.NoError(t, err)

		getteas := []Tea{}
		err = cursor.All(ctx, &getteas)
		assert.NoError(t, err)

		assert.Len(t, getteas, 1)
		assert.Equal(t, getteas[0].Type, "Masala")
	}
	{
		filter := bson.D{{"packaging.length", 5}, {"packaging.width", 10}}
		cursor, err := db.FindMany(ctx, filter)
		assert.NoError(t, err)

		getteas := []Tea{}
		err = cursor.All(ctx, &getteas)
		assert.NoError(t, err)

		assert.Len(t, getteas, 1)
		assert.Equal(t, getteas[0].Type, "Masala")
	}
}

func TestArrayEquals(t *testing.T) {
	ctx, db := MakeTestDB(t)

	{
		// Array match is not partial
		filter := bson.D{{"vendor", bson.A{"C"}}}
		cursor, err := db.FindMany(ctx, filter)
		assert.NoError(t, err)

		getteas := []Tea{}
		err = cursor.All(ctx, &getteas)
		assert.NoError(t, err)

		assert.Len(t, getteas, 1)
		assert.Equal(t, getteas[0].Type, "Oolong")
	}
	{
		filter := bson.D{{"vendor", bson.A{"A", "C"}}}
		cursor, err := db.FindMany(ctx, filter)
		assert.NoError(t, err)

		getteas := []Tea{}
		err = cursor.All(ctx, &getteas)
		assert.NoError(t, err)

		assert.Len(t, getteas, 1)
		assert.Equal(t, getteas[0].Type, "Masala")
	}
	{
		// Array match requires exact order
		filter := bson.D{{"vendor", bson.A{"C", "A"}}}
		cursor, err := db.FindMany(ctx, filter)
		assert.NoError(t, err)

		getteas := []Tea{}
		err = cursor.All(ctx, &getteas)
		assert.NoError(t, err)

		assert.Len(t, getteas, 0)
	}
	{
		// Array match requires exact order
		filter := bson.D{{"vendor.1", "C"}}
		cursor, err := db.FindMany(ctx, filter)
		assert.NoError(t, err)

		getteas := []Tea{}
		err = cursor.All(ctx, &getteas)
		assert.NoError(t, err)

		assert.Len(t, getteas, 1)
		assert.Equal(t, getteas[0].Type, "Masala")
	}
}

func TestArrayContains(t *testing.T) {
	ctx, db := MakeTestDB(t)

	{
		// Should match Teas with any size == 4
		filter := bson.D{{"sizes", bson.D{{"$eq", 4}}}}
		cursor, err := db.FindMany(ctx, filter)
		assert.NoError(t, err)

		getteas := []Tea{}
		err = cursor.All(ctx, &getteas)
		assert.NoError(t, err)

		assert.Len(t, getteas, 2)
		assert.ElementsMatch(t, getteas, teas[0:2])
	}

	{
		// Should match Teas with any size == 4
		filter := bson.D{{"sizes", 4}}
		cursor, err := db.FindMany(ctx, filter)
		assert.NoError(t, err)

		getteas := []Tea{}
		err = cursor.All(ctx, &getteas)
		assert.NoError(t, err)

		assert.Len(t, getteas, 2)
		assert.ElementsMatch(t, getteas, teas[0:2])
	}

	{
		// Should match Teas with any size <= 8
		filter := bson.D{{"sizes", bson.D{{"$lt", 8}}}}
		cursor, err := db.FindMany(ctx, filter)
		assert.NoError(t, err)

		getteas := []Tea{}
		err = cursor.All(ctx, &getteas)
		assert.NoError(t, err)

		assert.Len(t, getteas, 2)
		assert.Equal(t, getteas[0].Type, "Masala")
		assert.Equal(t, getteas[1].Type, "English Breakfast")
	}

	{
		filter := bson.D{{"sizes", bson.D{{"$gt", 8}}}}
		cursor, err := db.FindMany(ctx, filter)
		assert.NoError(t, err)

		getteas := []Tea{}
		err = cursor.All(ctx, &getteas)
		assert.NoError(t, err)

		assert.Len(t, getteas, 4)
		assert.ElementsMatch(t, getteas, teas[1:])
	}

	{
		filter := bson.D{{"sizes", bson.D{{"$gte", 8}, {"$lte", 8}}}}
		cursor, err := db.FindMany(ctx, filter)
		assert.NoError(t, err)

		getteas := []Tea{}
		err = cursor.All(ctx, &getteas)
		assert.NoError(t, err)

		assert.Len(t, getteas, 2)
		assert.ElementsMatch(t, getteas, teas[1:3])
	}

	{
		filter := bson.D{{"sizes", bson.D{{"$in", bson.A{16, 32}}}}}
		cursor, err := db.FindMany(ctx, filter)
		assert.NoError(t, err)

		getteas := []Tea{}
		err = cursor.All(ctx, &getteas)
		assert.NoError(t, err)

		assert.Len(t, getteas, 4)
		assert.ElementsMatch(t, getteas, teas[1:5])
	}

	{
		filter := bson.D{{"sizes", bson.D{{"$in", []int{16, 32}}}}}
		cursor, err := db.FindMany(ctx, filter)
		assert.NoError(t, err)

		getteas := []Tea{}
		err = cursor.All(ctx, &getteas)
		assert.NoError(t, err)

		assert.Len(t, getteas, 4)
		assert.ElementsMatch(t, getteas, teas[1:5])
	}

	{
		filter := bson.D{{"sizes", bson.D{{"$nin", bson.A{16, 32}}}}}
		cursor, err := db.FindMany(ctx, filter)
		assert.NoError(t, err)

		getteas := []Tea{}
		err = cursor.All(ctx, &getteas)
		assert.NoError(t, err)

		assert.Len(t, getteas, 3)
		assert.ElementsMatch(t, getteas, teas[0:3])
	}

	{
		filter := bson.D{{"sizes", bson.D{{"$in", bson.A{16, 32}}, {"$gt", 20}}}}
		cursor, err := db.FindMany(ctx, filter)
		assert.NoError(t, err)

		getteas := []Tea{}
		err = cursor.All(ctx, &getteas)
		assert.NoError(t, err)

		assert.Len(t, getteas, 1)
		assert.Equal(t, getteas[0], teas[4])
	}

	{
		filter := bson.D{{"sizes", bson.D{{"$nin", bson.A{16, 32}}, {"$gt", 20}}}}
		cursor, err := db.FindMany(ctx, filter)
		assert.NoError(t, err)

		getteas := []Tea{}
		err = cursor.All(ctx, &getteas)
		assert.NoError(t, err)

		assert.Len(t, getteas, 0)
	}
}

func TestAll(t *testing.T) {
	ctx, db := MakeTestDB(t)

	{
		filter := bson.D{{"sizes", bson.D{{"$all", bson.A{4, 8}}}}}
		cursor, err := db.FindMany(ctx, filter)
		assert.NoError(t, err)

		getteas := []Tea{}
		err = cursor.All(ctx, &getteas)
		assert.NoError(t, err)

		assert.Len(t, getteas, 1)
		assert.Equal(t, getteas[0].Type, "English Breakfast")
	}

	{
		filter := bson.D{{"sizes", bson.D{{"$all", bson.A{4, 8, 16, 32}}}}}
		cursor, err := db.FindMany(ctx, filter)
		assert.NoError(t, err)

		getteas := []Tea{}
		err = cursor.All(ctx, &getteas)
		assert.NoError(t, err)

		assert.Len(t, getteas, 0)
	}

	{
		filter := bson.D{{"sizes", bson.D{{"$all", bson.A{4}}}}}
		cursor, err := db.FindMany(ctx, filter)
		assert.NoError(t, err)

		getteas := []Tea{}
		err = cursor.All(ctx, &getteas)
		assert.NoError(t, err)

		assert.Len(t, getteas, 2)
		assert.ElementsMatch(t, getteas, teas[0:2])
	}
}

func TestElemMatch(t *testing.T) {
	ctx, db := MakeTestDB(t)

	{
		filter := bson.D{{"sizes", bson.D{{"$elemMatch", bson.D{{"nothing", bson.D{{"$gt", 4}}}}}}}}
		cursor, err := db.FindMany(ctx, filter)
		assert.NoError(t, err)

		getteas := []Tea{}
		err = cursor.All(ctx, &getteas)
		assert.NoError(t, err)

		assert.Len(t, getteas, 0)
	}

	ctx, db = MakeTestDB2(t)

	{
		filter := bson.D{{"teas.type", "Masala"}}
		cursor, err := db.FindMany(ctx, filter)
		assert.NoError(t, err)

		getteas := []TeaCollection{}
		err = cursor.All(ctx, &getteas)
		assert.NoError(t, err)

		assert.Len(t, getteas, 1)
	}

	{
		filter := bson.D{{"teas", bson.D{{"$elemMatch", bson.D{{"type", "Masala"}}}}}}
		cursor, err := db.FindMany(ctx, filter)
		assert.NoError(t, err)

		getteas := []TeaCollection{}
		err = cursor.All(ctx, &getteas)
		assert.NoError(t, err)

		assert.Len(t, getteas, 1)
	}

	{
		filter := bson.D{{"teas", bson.D{{"$elemMatch", bson.D{{"$or", bson.A{bson.D{{"type", "Masala"}}}}}}}}}
		cursor, err := db.FindMany(ctx, filter)
		assert.NoError(t, err)

		getteas := []TeaCollection{}
		err = cursor.All(ctx, &getteas)
		assert.NoError(t, err)

		assert.Len(t, getteas, 1)
		assert.Equal(t, getteas[0], teacollections[0])
	}

	{
		filter := bson.D{{"teas", bson.D{{"$elemMatch", bson.D{{"$or", bson.A{bson.D{{"type", "Assam"}}}}}}}}}
		cursor, err := db.FindMany(ctx, filter)
		assert.NoError(t, err)

		getteas := []TeaCollection{}
		err = cursor.All(ctx, &getteas)
		assert.NoError(t, err)

		assert.Len(t, getteas, 1)
		assert.Equal(t, getteas[0], teacollections[1])
	}

	{
		filter := bson.D{{"teas", bson.D{{"$elemMatch", bson.D{{"$or", bson.A{bson.D{{"type", "Masala"}}, bson.D{{"type", "English Breakfast"}}}}}}}}}
		cursor, err := db.FindMany(ctx, filter)
		assert.NoError(t, err)

		getteas := []TeaCollection{}
		err = cursor.All(ctx, &getteas)
		assert.NoError(t, err)

		assert.Len(t, getteas, 1)
	}

	{
		filter := bson.D{{"teas", bson.D{{"$elemMatch", bson.D{{"sizes", bson.D{{"$gt", 20}}}}}}}}
		cursor, err := db.FindMany(ctx, filter)
		assert.NoError(t, err)

		getteas := []TeaCollection{}
		err = cursor.All(ctx, &getteas)
		assert.NoError(t, err)

		assert.Len(t, getteas, 1)
		assert.Equal(t, getteas[0], teacollections[1])
	}

	{
		filter := bson.D{{"teas", bson.D{{"$elemMatch", bson.D{{"$or", bson.A{bson.D{{"type", "Masala"}}, bson.D{{"sizes", bson.D{{"$gt", 20}}}}}}}}}}}
		cursor, err := db.FindMany(ctx, filter)
		assert.NoError(t, err)

		getteas := []TeaCollection{}
		err = cursor.All(ctx, &getteas)
		assert.NoError(t, err)

		assert.Len(t, getteas, 2)
		assert.Equal(t, getteas[0], teacollections[0])
		assert.Equal(t, getteas[1], teacollections[1])
	}
}

func TestDeleteOne(t *testing.T) {
	ctx, db := MakeTestDB(t)

	{
		filter := bson.D{{"rating", bson.D{{"$gt", 5}}}}

		for i := 4; i >= 0; i-- {

			{
				cursor, err := db.FindMany(ctx, filter)
				assert.NoError(t, err)

				getteas := []Tea{}
				err = cursor.All(ctx, &getteas)
				assert.NoError(t, err)
				assert.Len(t, getteas, i)
			}

			{
				err := db.DeleteOne(ctx, filter)
				assert.NoError(t, err)
			}
		}

		{
			getteas := []Tea{}
			filter := bson.D{}
			cursor, err := db.FindMany(ctx, filter)
			assert.NoError(t, err)

			err = cursor.All(ctx, &getteas)
			assert.NoError(t, err)
			assert.Len(t, getteas, 1)
			assert.Equal(t, getteas[0], teas[3])
		}
	}
}

func TestDeleteMany(t *testing.T) {
	ctx, db := MakeTestDB(t)

	{
		filter := bson.D{{"rating", bson.D{{"$gt", 5}}}}

		{
			cursor, err := db.FindMany(ctx, filter)
			assert.NoError(t, err)

			getteas := []Tea{}
			err = cursor.All(ctx, &getteas)
			assert.NoError(t, err)
			assert.Len(t, getteas, 4)
		}

		{
			err := db.DeleteMany(ctx, filter)
			assert.NoError(t, err)
		}

		{
			getteas := []Tea{}
			cursor, err := db.FindMany(ctx, filter)
			assert.NoError(t, err)

			err = cursor.All(ctx, &getteas)
			assert.NoError(t, err)
			assert.Len(t, getteas, 0)
		}

		{
			getteas := []Tea{}
			filter := bson.D{}
			cursor, err := db.FindMany(ctx, filter)
			assert.NoError(t, err)

			err = cursor.All(ctx, &getteas)
			assert.NoError(t, err)
			assert.Len(t, getteas, 1)
			assert.Equal(t, getteas[0], teas[3])
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
	assert.NoError(t, err)

	getteas := []SimpleTea{}
	err = cursor.All(ctx, &getteas)
	assert.NoError(t, err)

	assert.Equal(t, 1, len(getteas))
	assert.Equal(t, "English Breakfast", getteas[0].Type)
}

func TestSimpleUpdate(t *testing.T) {
	ctx, db := MakeTestDB(t)

	{
		filter := bson.D{{"type", "English Breakfast"}}
		cursor, err := db.FindMany(ctx, filter)
		assert.NoError(t, err)

		getteas := []Tea{}
		err = cursor.All(ctx, &getteas)
		assert.NoError(t, err)

		assert.Equal(t, 1, len(getteas))
		assert.Equal(t, "English Breakfast", getteas[0].Type)
		assert.Equal(t, 6, getteas[0].Rating)
	}

	{
		filter := bson.D{{"type", "English Breakfast"}}
		update := bson.D{{"$set", bson.D{{"rating", 9}}}}

		updated, err := db.UpdateMany(ctx, filter, update)
		assert.Equal(t, 1, updated)
		assert.NoError(t, err)

		cursor, err := db.FindMany(ctx, filter)
		assert.NoError(t, err)

		getteas := []Tea{}
		err = cursor.All(ctx, &getteas)
		assert.NoError(t, err)

		assert.Equal(t, 1, len(getteas))
		assert.Equal(t, "English Breakfast", getteas[0].Type)
		assert.Equal(t, 9, getteas[0].Rating)
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
		assert.NoError(t, err)

		getteas := []ExtendedTea{}
		err = cursor.All(ctx, &getteas)
		assert.NoError(t, err)

		assert.Equal(t, 1, len(getteas))
		assert.Equal(t, "English Breakfast", getteas[0].Type)
		assert.Equal(t, "", getteas[0].Message)
	}

	{
		filter := bson.D{{"type", "English Breakfast"}}
		update := bson.D{{"$set", bson.D{{"message", "The best tea."}}}}

		updated, err := db.UpdateMany(ctx, filter, update)
		assert.Equal(t, 1, updated)
		assert.NoError(t, err)

		cursor, err := db.FindMany(ctx, filter)
		assert.NoError(t, err)

		getteas := []ExtendedTea{}
		err = cursor.All(ctx, &getteas)
		assert.NoError(t, err)

		assert.Equal(t, 1, len(getteas))
		assert.Equal(t, "English Breakfast", getteas[0].Type)
		assert.Equal(t, "The best tea.", getteas[0].Message)
	}

	{
		filter := bson.D{{"type", "English Breakfast"}}
		update := bson.D{{"$set", bson.D{{"fact.info", "Further information."}, {"fact.source", "jon"}}}}

		updated, err := db.UpdateMany(ctx, filter, update)
		assert.Equal(t, 1, updated)
		assert.NoError(t, err)

		cursor, err := db.FindMany(ctx, filter)
		assert.NoError(t, err)

		getteas := []ExtendedTea{}
		err = cursor.All(ctx, &getteas)
		assert.NoError(t, err)

		assert.Equal(t, 1, len(getteas))
		assert.Equal(t, "English Breakfast", getteas[0].Type)
		assert.Equal(t, "Further information.", getteas[0].Fact.Info)
		assert.Equal(t, "jon", getteas[0].Fact.Source)
	}
}

func TestUpdateArray(t *testing.T) {
	ctx, db := MakeTestDB(t)

	{
		filter := bson.D{{"type", "English Breakfast"}}
		cursor, err := db.FindMany(ctx, filter)
		assert.NoError(t, err)

		getteas := []ExtendedTea{}
		err = cursor.All(ctx, &getteas)
		assert.NoError(t, err)

		assert.Equal(t, 1, len(getteas))
		assert.Equal(t, "English Breakfast", getteas[0].Type)
		assert.Equal(t, 0, len(getteas[0].Facts))
	}

	{
		filter := bson.D{{"type", "English Breakfast"}}
		update := bson.D{{"$set", bson.D{{"facts.3.info", "The best tea."}}}}

		updated, err := db.UpdateMany(ctx, filter, update)
		assert.Equal(t, 1, updated)
		assert.NoError(t, err)

		cursor, err := db.FindMany(ctx, filter)
		assert.NoError(t, err)

		getteas := []ExtendedTea{}
		err = cursor.All(ctx, &getteas)
		assert.NoError(t, err)

		assert.Equal(t, 1, len(getteas))
		assert.Equal(t, "English Breakfast", getteas[0].Type)
		assert.Equal(t, 4, len(getteas[0].Facts))
		assert.Equal(t, "The best tea.", getteas[0].Facts[3].Info)
	}

	{
		filter := bson.D{{"type", "English Breakfast"}}
		update := bson.D{{"$set", bson.D{{"facts.3.info", "no fact."}}}}

		updated, err := db.UpdateMany(ctx, filter, update)
		assert.Equal(t, 1, updated)
		assert.NoError(t, err)

		cursor, err := db.FindMany(ctx, filter)
		assert.NoError(t, err)

		getteas := []ExtendedTea{}
		err = cursor.All(ctx, &getteas)
		assert.NoError(t, err)

		assert.Equal(t, 1, len(getteas))
		assert.Equal(t, "English Breakfast", getteas[0].Type)
		assert.Equal(t, 4, len(getteas[0].Facts))
		assert.Equal(t, "no fact.", getteas[0].Facts[3].Info)
	}

	{
		filter := bson.D{{"type", "English Breakfast"}}
		update := bson.D{{"$set", bson.D{{"facts", bson.A{}}}}}

		updated, err := db.UpdateMany(ctx, filter, update)
		assert.Equal(t, 1, updated)
		assert.NoError(t, err)

		cursor, err := db.FindMany(ctx, filter)
		assert.NoError(t, err)

		getteas := []ExtendedTea{}
		err = cursor.All(ctx, &getteas)
		assert.NoError(t, err)

		assert.Equal(t, 1, len(getteas))
		assert.Equal(t, "English Breakfast", getteas[0].Type)
		assert.Equal(t, 0, len(getteas[0].Facts))
	}
}

func TestUnset(t *testing.T) {
	ctx, db := MakeTestDB(t)

	{
		filter := bson.D{{"rating", 6}}
		cursor, err := db.FindMany(ctx, filter)
		assert.NoError(t, err)

		getteas := []ExtendedTea{}
		err = cursor.All(ctx, &getteas)
		assert.NoError(t, err)

		assert.Equal(t, 1, len(getteas))
		assert.Equal(t, "English Breakfast", getteas[0].Type)
		assert.Equal(t, 0, len(getteas[0].Facts))
	}

	{
		filter := bson.D{{"rating", 6}, {"sizes", 8}}
		update := bson.D{{"$unset", bson.D{{"type", ""}}}}

		updated, err := db.UpdateMany(ctx, filter, update)
		assert.Equal(t, 1, updated)
		assert.NoError(t, err)

		cursor, err := db.FindMany(ctx, filter)
		assert.NoError(t, err)

		getteas := []ExtendedTea{}
		err = cursor.All(ctx, &getteas)
		assert.NoError(t, err)

		assert.Equal(t, 1, len(getteas))
		assert.Equal(t, "", getteas[0].Type)
	}

	{
		filter := bson.D{{"rating", 6}, {"sizes", 8}}
		update := bson.D{{"$unset", bson.D{{"sizes.1", ""}}}}

		updated, err := db.UpdateMany(ctx, filter, update)
		assert.Equal(t, 1, updated)
		assert.NoError(t, err)

		cursor, err := db.FindMany(ctx, filter)
		assert.NoError(t, err)

		getteas := []ExtendedTea{}
		err = cursor.All(ctx, &getteas)
		assert.NoError(t, err)

		assert.Equal(t, 0, len(getteas))
	}
}

func TestInc(t *testing.T) {
	ctx, db := MakeTestDB(t)

	{
		filter := bson.D{{"type", "English Breakfast"}}
		cursor, err := db.FindMany(ctx, filter)
		assert.NoError(t, err)

		getteas := []Tea{}
		err = cursor.All(ctx, &getteas)
		assert.NoError(t, err)

		assert.Equal(t, 1, len(getteas))
		assert.Equal(t, "English Breakfast", getteas[0].Type)
		assert.Equal(t, 6, getteas[0].Rating)
	}

	{
		filter := bson.D{{"type", "English Breakfast"}}

		for i := 1; i < 5; i++ {
			update := bson.D{{"$inc", bson.D{{"rating", 1}}}}

			updated, err := db.UpdateMany(ctx, filter, update)
			assert.Equal(t, 1, updated)
			assert.NoError(t, err)

			cursor, err := db.FindMany(ctx, filter)
			assert.NoError(t, err)

			getteas := []Tea{}
			err = cursor.All(ctx, &getteas)
			assert.NoError(t, err)

			assert.Equal(t, 1, len(getteas))
			assert.Equal(t, "English Breakfast", getteas[0].Type)
			assert.Equal(t, 6+i, getteas[0].Rating)
		}
	}
}

func TestIncArray(t *testing.T) {
	ctx, db := MakeTestDB(t)

	{
		filter := bson.D{{"type", "English Breakfast"}}
		cursor, err := db.FindMany(ctx, filter)
		assert.NoError(t, err)

		getteas := []Tea{}
		err = cursor.All(ctx, &getteas)
		assert.NoError(t, err)

		assert.Equal(t, 1, len(getteas))
		assert.Equal(t, "English Breakfast", getteas[0].Type)
		assert.Equal(t, 3, len(getteas[0].Sizes))
	}

	{
		filter := bson.D{{"type", "English Breakfast"}}

		for i := 1; i < 5; i++ {
			update := bson.D{{"$inc", bson.D{{"sizes.5", 2}}}}

			updated, err := db.UpdateMany(ctx, filter, update)
			assert.Equal(t, 1, updated)
			assert.NoError(t, err)

			cursor, err := db.FindMany(ctx, filter)
			assert.NoError(t, err)

			getteas := []Tea{}
			err = cursor.All(ctx, &getteas)
			assert.NoError(t, err)

			assert.Equal(t, 1, len(getteas))
			assert.Equal(t, "English Breakfast", getteas[0].Type)
			assert.Equal(t, 6, len(getteas[0].Sizes))
			assert.Equal(t, int32(2*i), getteas[0].Sizes[5])
		}
	}
}

func TestPush(t *testing.T) {
	ctx, db := MakeTestDB(t)

	{
		filter := bson.D{{"type", "Masala"}}
		cursor, err := db.FindMany(ctx, filter)
		assert.NoError(t, err)

		getteas := []Tea{}
		err = cursor.All(ctx, &getteas)
		assert.NoError(t, err)

		assert.Equal(t, 1, len(getteas))
		assert.Equal(t, "Masala", getteas[0].Type)
		assert.ElementsMatch(t, getteas[0].Sizes, []int32{4})
	}

	{
		filter := bson.D{{"type", "Masala"}}
		update := bson.D{{"$push", bson.D{{"sizes", 6}}}}

		updated, err := db.UpdateMany(ctx, filter, update)
		assert.Equal(t, 1, updated)
		assert.NoError(t, err)

		cursor, err := db.FindMany(ctx, filter)
		assert.NoError(t, err)

		getteas := []Tea{}
		err = cursor.All(ctx, &getteas)
		assert.NoError(t, err)

		assert.Equal(t, 1, len(getteas))
		assert.Equal(t, "Masala", getteas[0].Type)
		assert.ElementsMatch(t, getteas[0].Sizes, []int32{4, 6})
	}

	{
		filter := bson.D{{"type", "English Breakfast"}}
		update := bson.D{{"$push", bson.D{{"sizes", 12}, {"Vendor", "newvendor"}}}}

		updated, err := db.UpdateMany(ctx, filter, update)
		assert.Equal(t, 1, updated)
		assert.NoError(t, err)

		cursor, err := db.FindMany(ctx, filter)
		assert.NoError(t, err)

		getteas := []Tea{}
		err = cursor.All(ctx, &getteas)
		assert.NoError(t, err)

		assert.Equal(t, 1, len(getteas))
		assert.Equal(t, "English Breakfast", getteas[0].Type)
		assert.ElementsMatch(t, getteas[0].Sizes, []int32{4, 8, 16, 12})
		assert.ElementsMatch(t, getteas[0].Vendor, []string{"newvendor"})
	}
}

func TestPull(t *testing.T) {
	ctx, db := MakeTestDB(t)

	{
		update := bson.D{{"$pull", bson.D{{"sizes", 8}}}}
		_, err := db.UpdateMany(ctx, bson.D{}, update)
		assert.NoError(t, err)

		filter := bson.D{{"type", "English Breakfast"}}
		cursor, err := db.FindOne(ctx, filter)
		assert.NoError(t, err)

		var tea Tea
		_, err = cursor.One(ctx, &tea)
		assert.NoError(t, err)
		assert.Equal(t, "English Breakfast", tea.Type)
		assert.Len(t, tea.Sizes, 2)
		assert.ElementsMatch(t, tea.Sizes, []int32{4, 16})
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
		assert.NoError(t, err)

		var id ObjectIDTest
		assert.Equal(t, id.ID, primitive.NilObjectID)

		_, err = cursor.One(ctx, &id)
		assert.NoError(t, err)
		assert.NotEqual(t, id.ID, primitive.NilObjectID)
		assert.Equal(t, id.Type, "English Breakfast")
	}
}

func TestUpdateByObjectID(t *testing.T) {
	ctx, db := MakeTestDB(t)

	{
		typeFilter := bson.D{{"type", "English Breakfast"}}
		cursor, err := db.FindOne(ctx, typeFilter)
		assert.NoError(t, err)

		var id ObjectIDTest
		_, err = cursor.One(ctx, &id)
		assert.NoError(t, err)
		assert.Equal(t, id.Type, "English Breakfast")

		idFilter := bson.D{{"_id", id.ID}}
		cursor, err = db.FindOne(ctx, idFilter)
		assert.NoError(t, err)

		_, err = cursor.One(ctx, &id)
		assert.NoError(t, err)
		assert.Equal(t, id.Type, "English Breakfast")

		updateTo := bson.D{{"$set", bson.D{{"type", "Scottish Breakfast"}}}}
		updated, err := db.UpdateOne(ctx, idFilter, updateTo)
		assert.Equal(t, 1, updated)
		assert.NoError(t, err)

		cursor, err = db.FindOne(ctx, idFilter)
		assert.NoError(t, err)

		_, err = cursor.One(ctx, &id)
		assert.NoError(t, err)
		assert.Equal(t, id.Type, "Scottish Breakfast")

		cursor, err = db.FindOne(ctx, typeFilter)
		assert.NoError(t, err)

		_, err = cursor.One(ctx, &id)
		assert.NoError(t, err)
		assert.Equal(t, id.Type, "")
	}

}
