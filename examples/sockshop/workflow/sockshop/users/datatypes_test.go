package users

import (
	"reflect"
	"testing"
)

func TestAddLinksAdd(t *testing.T) {
	a := Address{ID: "test"}
	a.AddLinks()
	h := Href{"userservice/addresses/test"}
	if !reflect.DeepEqual(a.Links["address"], h) {
		t.Error("expected equal address links")
	}

}
func TestAddLinksCard(t *testing.T) {
	c := Card{ID: "test"}
	c.AddLinks()
	h := Href{"userservice/cards/test"}
	if !reflect.DeepEqual(c.Links["card"], h) {
		t.Error("expected equal address links")
	}

}

func TestMaskCC(t *testing.T) {
	test1 := "1234567890"
	c := Card{LongNum: test1}
	c.MaskCC()
	test1comp := "******7890"
	if c.LongNum != test1comp {
		t.Errorf("Expected matching CC number %v received %v", test1comp, test1)
	}
}
