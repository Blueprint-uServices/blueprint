package query

import (
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"

	"go.mongodb.org/mongo-driver/bson"
)

/*
Simple query evaluation on golang structs
*/

type Filter interface {
	Apply(item any) bool
	String() string
}

type (
	selectFilter struct {
		fieldName string
		next      Filter
	}

	index struct {
		i    int
		next Filter
	}

	broadcast struct {
		next Filter
	}

	equals struct {
		value any
	}

	CmpType int

	cmpInt struct {
		value int64
		cmp   CmpType
	}

	cmpFloat struct {
		value float64
		cmp   CmpType
	}

	and struct {
		filters []Filter
	}

	or struct {
		filters []Filter
	}

	not struct {
		filter Filter
	}

	exists struct{}

	exactFieldMatch struct {
		fields map[string]struct{}
	}

	regex struct {
		re *regexp.Regexp
	}

	all struct {
		values []Filter
	}

	elemMatch struct {
		queries []Filter
	}
)

const (
	Eq CmpType = iota
	Gt
	Gte
	Lt
	Lte
)

func Lookup(selector string, next Filter) Filter {
	splits := strings.Split(selector, ".")
	for i := len(splits) - 1; i >= 0; i-- {
		if j, err := strconv.Atoi(splits[i]); err == nil {
			next = Index(j, next)
		} else {
			next = Select(splits[i], next)
			if i != 0 {
				next = Broadcast(next) // broadcast in case the field is an array type
			}
		}
	}
	return next
}

func Select(fieldName string, next Filter) Filter {
	return &selectFilter{fieldName: fieldName, next: next}
}

func Index(i int, next Filter) Filter {
	return &index{i: i, next: next}
}

func Broadcast(next Filter) Filter {
	return &broadcast{next: next}
}

func Equals(value any) Filter {
	if i, isInt := intValue(value); isInt {
		return CmpInt(i, Eq)
	} else if f, isFloat := floatValue(value); isFloat {
		return CmpFloat(f, Eq)
	} else {
		return &equals{value: value}
	}
}

func CmpInt(value int64, cmp CmpType) Filter {
	return &cmpInt{value: value, cmp: cmp}
}

func CmpFloat(value float64, cmp CmpType) Filter {
	return &cmpFloat{value: value, cmp: cmp}
}

func And(filters ...Filter) Filter {
	switch len(filters) {
	case 0:
		return Exists()
	case 1:
		return filters[0]
	default:
		return &and{filters: filters}
	}
}

func Or(filters ...Filter) Filter {
	switch len(filters) {
	case 0:
		return Exists()
	case 1:
		return filters[0]
	default:
		return &or{filters: filters}
	}
}

func Not(filter Filter) Filter {
	return &not{filter: filter}
}

func Exists() Filter {
	return &exists{}
}

func ExactFieldMatch(fieldsSeen ...string) Filter {
	lookup := make(map[string]struct{})
	for _, fieldName := range fieldsSeen {
		lookup[fieldName] = struct{}{}
	}
	return &exactFieldMatch{fields: lookup}
}

func Regex(regex_string string) (Filter, error) {
	re, err := regexp.Compile(regex_string)
	return &regex{re: re}, err
}

func All(filters ...Filter) Filter {
	return &all{values: filters}
}

func ElemMatch(queries ...Filter) Filter {
	return &elemMatch{queries: queries}
}

func (f *selectFilter) Apply(item any) bool {
	if d, isD := item.(bson.D); isD {
		for _, e := range d {
			if e.Key == f.fieldName {
				return f.next.Apply(e.Value)
			}
		}
	}
	return false
}

func (f *index) Apply(item any) bool {
	if a, isA := item.(bson.A); isA && f.i < len(a) {
		return f.next.Apply(a[f.i])
	}
	return false
}

func (f *broadcast) Apply(item any) bool {
	if a, isA := item.(bson.A); isA {
		for _, e := range a {
			if f.next.Apply(e) {
				return true
			}
		}
		return false
	} else {
		return f.next.Apply(item)
	}
}

func (f *equals) Apply(item any) bool {
	return reflect.DeepEqual(item, f.value)
}

func (f *cmpInt) Apply(item any) bool {
	if v, isInt := intValue(item); isInt {
		switch f.cmp {
		case Eq:
			return v == f.value
		case Gt:
			return v > f.value
		case Gte:
			return v >= f.value
		case Lt:
			return v < f.value
		case Lte:
			return v <= f.value
		}
	}
	return false
}

func (f *cmpFloat) Apply(item any) bool {
	if v, isFloat := floatValue(item); isFloat {
		switch f.cmp {
		case Eq:
			return v == f.value
		case Gt:
			return v > f.value
		case Gte:
			return v >= f.value
		case Lt:
			return v < f.value
		case Lte:
			return v <= f.value
		}
	}
	return false
}

func (f *and) Apply(item any) bool {
	for _, filter := range f.filters {
		if !filter.Apply(item) {
			return false
		}
	}
	return true
}

func (f *or) Apply(item any) bool {
	for _, filter := range f.filters {
		if filter.Apply(item) {
			return true
		}
	}
	return false
}

func (f *not) Apply(item any) bool {
	return !f.filter.Apply(item)
}

func (f *exists) Apply(item any) bool {
	return true
}

func (f *exactFieldMatch) Apply(item any) bool {
	d, isD := item.(bson.D)
	if !isD {
		return false
	}

	for _, e := range d {
		if _, seen := f.fields[e.Key]; !seen {
			return false
		}
	}
	return true
}

func (f *regex) Apply(item any) bool {
	strval, isStr := item.(string)
	return isStr && f.re.MatchString(strval)
}

func (f *all) Apply(item any) bool {
	a, isA := item.(bson.A)
	if !isA {
		return false
	}

outer:
	for _, valuefilter := range f.values {
		for _, e := range a {
			if valuefilter.Apply(e) {
				continue outer
			}
		}
		return false
	}
	return true
}

func (f *elemMatch) Apply(item any) bool {
	a, isA := item.(bson.A)
	if !isA {
		return false
	}

outer:
	for _, e := range a {
		for _, query := range f.queries {
			if !query.Apply(e) {
				continue outer
			}
		}
		return true
	}
	return false
}

func (f *selectFilter) String() string {
	return fmt.Sprintf(".%v %v", f.fieldName, f.next.String())
}

func (f *index) String() string {
	return fmt.Sprintf("[%v] %v", f.i, f.next.String())
}

func (f *broadcast) String() string {
	return fmt.Sprintf(" broadcast %v", f.next.String())
}

func (f *equals) String() string {
	return fmt.Sprintf("= %v", f.value)
}

func (f *cmpInt) String() string {
	switch f.cmp {
	case Eq:
		return fmt.Sprintf("= %v", f.value)
	case Gt:
		return fmt.Sprintf("> %v", f.value)
	case Gte:
		return fmt.Sprintf(">= %v", f.value)
	case Lt:
		return fmt.Sprintf("< %v", f.value)
	case Lte:
		return fmt.Sprintf("<= %v", f.value)
	default:
		return fmt.Sprintf("???%v %v", f.cmp, f.value)
	}
}

func (f *cmpFloat) String() string {
	return fmt.Sprintf("CmpFloat %v %v", f.cmp, f.value)
}

func (f *exactFieldMatch) String() string {
	var fieldNames []string
	for fieldName := range f.fields {
		fieldNames = append(fieldNames, fieldName)
	}
	return fmt.Sprintf("ExactFieldMatch (%v)", strings.Join(fieldNames, ", "))
}

func (f *and) String() string {
	var filterstrings []string
	for _, filter := range f.filters {
		filterstrings = append(filterstrings, "("+filter.String()+")")
	}
	return strings.Join(filterstrings, " && ")
}

func (f *or) String() string {
	var filterstrings []string
	for _, filter := range f.filters {
		filterstrings = append(filterstrings, "("+filter.String()+")")
	}
	return strings.Join(filterstrings, " || ")
}

func (f *not) String() string {
	return fmt.Sprintf("NOT (%v)", f.filter.String())
}

func (f *exists) String() string {
	return "exists"
}

func (f *regex) String() string {
	return fmt.Sprintf("regex(%v)", f.re.String())
}

func (f *all) String() string {
	var valueStrings []string
	for _, value := range f.values {
		valueStrings = append(valueStrings, fmt.Sprintf("%v", value))
	}
	return fmt.Sprintf("all(%v)", strings.Join(valueStrings, ", "))
}

func (f *elemMatch) String() string {
	var queryStrings []string
	for _, query := range f.queries {
		queryStrings = append(queryStrings, fmt.Sprintf("%v", query))
	}
	return fmt.Sprintf("elemMatch(%v)", strings.Join(queryStrings, ", "))
}

func intValue(item any) (int64, bool) {
	switch v := item.(type) {
	case int:
		return int64(v), true
	case int8:
		return int64(v), true
	case int16:
		return int64(v), true
	case int32:
		return int64(v), true
	case int64:
		return v, true
	default:
		return 0, false
	}
}

func floatValue(item any) (float64, bool) {
	switch v := item.(type) {
	case int:
		return float64(v), true
	case int8:
		return float64(v), true
	case int16:
		return float64(v), true
	case int32:
		return float64(v), true
	case int64:
		return float64(v), true
	case float32:
		return float64(v), true
	case float64:
		return v, true
	default:
		return 0, false
	}
}
