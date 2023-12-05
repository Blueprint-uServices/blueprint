package query

import (
	"fmt"
	"reflect"
	"strings"

	"go.mongodb.org/mongo-driver/bson"
)

/*
Simple BSON query parser
*/

func ParseFilter(filter bson.D) (Filter, error) {
	return parseQuery(filter)
}

var arrayOperators = map[string]struct{}{}

func init() {
	for _, op := range []string{"$all", "$elemMatch", "$size"} {
		arrayOperators[op] = struct{}{}
	}
}

func parseQueries(a bson.A) ([]Filter, error) {
	var filters []Filter
	for _, v := range a {
		d, isD := v.(bson.D)
		if !isD {
			return nil, fmt.Errorf("invalid query %v; expected a bson.D", v)
		}
		filter, err := parseQuery(d)
		if err != nil {
			return nil, err
		}
		filters = append(filters, filter)
	}
	return filters, nil
}

func parseQuery(d bson.D) (Filter, error) {
	var filters []Filter
	for _, e := range d {
		filter, err := parseFilterCondition(e)
		if err != nil {
			return nil, err
		}
		filters = append(filters, filter)
	}
	return And(filters...), nil
}

/*
Parses a value, e.g. the value part of

{ <field>: <value> }
or
{ <field1>: { <operator1>: <value1> }
*/
func parseValues(d bson.D) (Filter, error) {
	var filters []Filter
	var fieldsSeen []string
	for _, e := range d {
		if strings.HasPrefix(e.Key, "$") {
			return nil, fmt.Errorf("values cannot contain operators; found %v in %v", e.Key, d)
		}

		if strings.Contains(e.Key, ".") {
			return nil, fmt.Errorf("cannot use selector expression %v within value condition %v", e.Key, d)
		}

		filter, err := parseValue(e.Value)
		if err != nil {
			return nil, err
		}
		filters = append(filters, Select(e.Key, filter))
		fieldsSeen = append(fieldsSeen, e.Key)
	}
	// Struct matches must be exact matches
	filters = append(filters, ExactFieldMatch(fieldsSeen...))
	return And(filters...), nil
}

func parseValue(value any) (Filter, error) {
	switch v := value.(type) {
	case bson.D:
		return parseValues(v)
	case bson.A:
		return parseArrayValues(v)
	case bson.E:
		return nil, fmt.Errorf("values must be composed of bson.D, bson.A, or value literals; not bson.E found in %v", v)
	case bson.M:
		return nil, fmt.Errorf("values must be composed of bson.D, bson.A, or value literals; not bson.M found in %v", v)
	default:
		return Equals(v), nil
	}
}

func parseArrayValues(a bson.A) (Filter, error) {
	/* All elements of the array must match exactly */
	var arrayFilters []Filter
	for i := range a {
		switch v := a[i].(type) {
		case bson.D:
			{
				filter, err := parseValues(v)
				if err != nil {
					return nil, err
				}
				arrayFilters = append(arrayFilters, Index(i, filter))
			}
		case bson.A:
			{
				filter, err := parseArrayValues(v)
				if err != nil {
					return nil, err
				}
				arrayFilters = append(arrayFilters, Index(i, filter))
			}
			return nil, fmt.Errorf("array values must be composed of bson.D or value literals; not bson.A found in %v", v)
		case bson.E:
			return nil, fmt.Errorf("array values must be composed of bson.D or value literals; not bson.E found in %v", v)
		case bson.M:
			return nil, fmt.Errorf("array values must be composed of bson.D or value literals; not bson.M found in %v", v)
		default:
			arrayFilters = append(arrayFilters, Index(i, Equals(v)))
		}
	}
	return And(arrayFilters...), nil
}

/*
A root condition of a filter document
*/
func parseFilterCondition(e bson.E) (Filter, error) {
	/*
		A root condition of a filter document can contain operators, but only logical operators
	*/
	if f, err := parseLogicalOperator(e); f != nil || err != nil {
		return f, err
	}
	if strings.HasPrefix(e.Key, "$") {
		return nil, fmt.Errorf("encountered unexpected condition key %v in %v", e.Key, e)
	}

	// Now we expect a field selector and a condition or a (possibly nested) value
	switch v := e.Value.(type) {
	case bson.D:
		{
			hasOperators := false
			for _, e := range v {
				if strings.HasPrefix(e.Key, "$") {
					hasOperators = true
					break
				}
			}

			if !hasOperators {
				// If no operators, then it's just a direct value match or an
				// array OR value match
				filter, err := parseValues(v)
				return Lookup(e.Key, Broadcast(filter)), err
			}

			// Array operators can't mix with regular value operators
			if filter, err := parseArrayOperators(v); err != nil || filter != nil {
				return Lookup(e.Key, filter), err
			}

			// $exists: false is a special case
			if filter, err := parseExistsOperators(e.Key, v); err != nil || filter != nil {
				return filter, err
			}

			// We have regular value operators
			filter, err := parseValueOperators(v)
			return Lookup(e.Key, Broadcast(filter)), err
		}
	case bson.A:
		{
			// Exact match an array value
			filter, err := parseArrayValues(v)
			return Lookup(e.Key, filter), err
		}
	case bson.E:
		return nil, fmt.Errorf("condition values should be specified using bson.D not bson.E")
	case bson.M:
		return nil, fmt.Errorf("condition values should be specified using bson.D not bson.M")
	default:
		return Lookup(e.Key, Broadcast(Equals(v))), nil
	}
}

/*
If the element has a logical operator as key, parses it, or returns nil.
*/
func parseLogicalOperator(e bson.E) (Filter, error) {
	switch e.Key {
	case "$and":
		{
			a, isA := e.Value.(bson.A)
			if !isA {
				return nil, fmt.Errorf(`invalid query: "$and" key must correspond to a bson.A value: %v`, e)
			}
			filters, err := parseQueries(a)
			return And(filters...), err
		}
	case "$not":
		{
			d, isD := e.Value.(bson.D)
			if !isD {
				return nil, fmt.Errorf(`invalid query: "$not" key must correspond to a bson.D value: %v`, e)
			}
			filter, err := parseQuery(d)
			return Not(filter), err
		}
	case "$or":
		{
			a, isA := e.Value.(bson.A)
			if !isA {
				return nil, fmt.Errorf(`invalid query: "$or" key must correspond to a bson.A value: %v`, e)
			}
			filters, err := parseQueries(a)
			return Or(filters...), err
		}
	case "$nor":
		{
			a, isA := e.Value.(bson.A)
			if !isA {
				return nil, fmt.Errorf(`invalid query: "$nor" key must correspond to a bson.A value: %v`, e)
			}
			filters, err := parseQueries(a)
			return Not(Or(filters...)), err
		}
	default:
		// Not a logical operator
		return nil, nil
	}
}

/*
Parses the provided document for the following array operators:
$all, $elemMatch, $size
If anything is found in the document that does not match these,
this will return nil.
*/
func parseArrayOperators(d bson.D) (Filter, error) {
	filters := []Filter{}
	for _, e := range d {
		filter, err := parseArrayOperator(e)
		if err != nil {
			return nil, err
		}
		if filter != nil {
			filters = append(filters, filter)
		}
	}
	if len(filters) == 0 {
		return nil, nil
	}
	if len(filters) != len(d) {
		return nil, fmt.Errorf("invalid mix of array and non-array filter operators %v", d)
	}

	return And(filters...), nil
}

func parseArrayOperator(e bson.E) (Filter, error) {
	switch e.Key {
	case "$all":
		{
			values, isA := e.Value.(bson.A)
			if !isA {
				return nil, fmt.Errorf("$all must be an array of values but got %v", e)
			}
			var valuefilters []Filter
			for _, value := range values {
				valuefilter, err := parseValue(value)
				if err != nil {
					return nil, err
				}
				valuefilters = append(valuefilters, valuefilter)
			}
			return All(valuefilters...), nil
		}
	case "$elemMatch":
		{
			d, isD := e.Value.(bson.D)
			if !isD {
				return nil, fmt.Errorf("$elemMatch must be a bson.D but got %v", e)
			}
			filter, err := parseQuery(d)
			return Broadcast(filter), err
		}
	case "$size":
	}
	return nil, nil
}

// $exists: false is a special case that can also ignore all other operators
// This only returns a filter for $exists: false; not $exists: true
func parseExistsOperators(key string, d bson.D) (Filter, error) {
	for _, e := range d {
		if e.Key == "$exists" {
			if v, isBool := e.Value.(bool); isBool {
				if !v {
					return Not(Lookup(e.Key, Exists())), nil
				} else {
					return nil, nil
				}
			} else {
				return nil, fmt.Errorf("$exists requires a bool value but got %v", e)
			}
		}
	}
	return nil, nil
}

/*
Parses a value that has some operators in it
*/
func parseValueOperators(d bson.D) (Filter, error) {
	filters := []Filter{}
	for _, e := range d {

		filter, err := parseValueOperator(e)
		if err != nil {
			return nil, err
		}
		filters = append(filters, filter)
	}
	return And(filters...), nil
}

func parseValueOperator(e bson.E) (Filter, error) {

	if filter, err := parseNumericOperator(e); err != nil || filter != nil {
		return filter, err
	}

	switch e.Key {
	case "$eq":
		return parseValue(e.Value)
	case "$ne":
		{
			f, err := parseValue(e.Value)
			return Not(f), err
		}
	case "$in":
		{
			v := reflect.ValueOf(e.Value)
			if v.Kind() != reflect.Slice {
				return nil, fmt.Errorf("$in requires a bson.A value or a slice, but got %v", e)
			}
			var filters []Filter
			for i := 0; i < v.Len(); i++ {
				filter, err := parseValue(v.Index(i).Interface())
				if err != nil {
					return nil, err
				}
				filters = append(filters, filter)
			}
			if len(filters) == 0 {
				return Not(Exists()), nil
			}
			return Or(filters...), nil
		}
	case "$nin":
		{
			return nil, fmt.Errorf("$nin not implemented; use $nor and $in instead")
			// filter, err := parseValueOperator(bson.E{"$in", e.Value})
			// return Not(filter), err
		}
	case "$regex":
		{
			if strval, isStr := e.Value.(string); isStr {
				return Regex(strval)
			}
			return nil, fmt.Errorf("$regex value must be string but got %v", e.Value)
		}
	case "$text":
		fallthrough
	case "$where": // not supported
		fallthrough
	case "$type": // not supported
		fallthrough
	case "$expr": // not supported
		fallthrough
	default:
		return nil, fmt.Errorf("unsupported operator %v", e)
	}
}

func parseNumericOperator(e bson.E) (Filter, error) {
	if v, isInt := intValue(e.Value); isInt {
		switch e.Key {
		case "$eq":
			return CmpInt(v, Eq), nil
		case "$gt":
			return CmpInt(v, Gt), nil
		case "$gte":
			return CmpInt(v, Gte), nil
		case "$lt":
			return CmpInt(v, Lt), nil
		case "$lte":
			return CmpInt(v, Lte), nil
		}
	} else if v, isFloat := floatValue(e.Value); isFloat {
		switch e.Key {
		case "$eq":
			return CmpFloat(v, Eq), nil
		case "$gt":
			return CmpFloat(v, Gt), nil
		case "$gte":
			return CmpFloat(v, Gte), nil
		case "$lt":
			return CmpFloat(v, Lt), nil
		case "$lte":
			return CmpFloat(v, Lte), nil
		}
	}
	return nil, nil
}
