package query

import (
	"fmt"
	"reflect"
	"strings"

	"go.mongodb.org/mongo-driver/bson"
)

func ParseUpdate(update bson.D) (Update, error) {
	var updates []Update
	for _, op := range update {
		switch op.Key {
		case "$set":
			{
				sets, err := parseSet(op.Value)
				if err != nil {
					return nil, err
				}
				updates = append(updates, sets...)
			}
		case "$unset":
			{
				unsets, err := parseUnset(op.Value)
				if err != nil {
					return nil, err
				}
				updates = append(updates, unsets...)
			}
		case "$inc":
			{
				incs, err := parseInc(op.Value)
				if err != nil {
					return nil, err
				}
				updates = append(updates, incs...)
			}
		case "$push":
			{
				pushes, err := parsePush(op.Value)
				if err != nil {
					return nil, err
				}
				updates = append(updates, pushes...)
			}
		case "$pull":
			{
				pulls, err := parsePull(op.Value)
				if err != nil {
					return nil, err
				}
				updates = append(updates, pulls...)
			}
		case "$addToSet":
			{
				adds, err := parseAddToSet(op.Value)
				if err != nil {
					return nil, err
				}
				updates = append(updates, adds...)
			}
		default:
			return nil, fmt.Errorf("unsupported update op %v", op.Key)
		}
	}
	return UpdateAll(updates), nil
}

func parseSet(args any) ([]Update, error) {
	d, isD := args.(bson.D)
	if !isD {
		return nil, fmt.Errorf("invalid $set operator; expected a bson.D, got %v", args)
	}
	var updates []Update
	for _, e := range d {
		update, err := SetValue(e.Value)
		if err != nil {
			return nil, err
		}
		updates = append(updates, UpdatePath(e.Key, update))
	}
	return updates, nil
}

func parseUnset(args any) ([]Update, error) {
	d, isD := args.(bson.D)
	if !isD {
		return nil, fmt.Errorf("invalid $unset operator; expected a bson.D, got %v", args)
	}
	var updates []Update
	for _, e := range d {
		updates = append(updates, UnsetPath(e.Key))
	}
	return updates, nil
}

func parseInc(args any) ([]Update, error) {
	d, isD := args.(bson.D)
	if !isD {
		return nil, fmt.Errorf("invalid $inc operator; expected a bson.D, got %v", args)
	}
	var updates []Update
	for _, e := range d {
		switch v := e.Value.(type) {
		case int:
			updates = append(updates, UpdatePath(e.Key, IncInt(int64(v))))
		case int64:
			updates = append(updates, UpdatePath(e.Key, IncInt(v)))
		case int32:
			updates = append(updates, UpdatePath(e.Key, IncInt(int64(v))))
		case int16:
			updates = append(updates, UpdatePath(e.Key, IncInt(int64(v))))
		case int8:
			updates = append(updates, UpdatePath(e.Key, IncInt(int64(v))))
		// case float64:
		// 	updates = append(updates, IncFloat(v))
		// case float32:
		// 	updates = append(updates, IncFloat(float64(v)))
		default:
			return nil, fmt.Errorf("invalid $inc argument; expect an int, got %v %v", reflect.TypeOf(e.Value), e.Value)
		}
	}
	return updates, nil
}

func parsePush(args any) ([]Update, error) {
	d, isD := args.(bson.D)
	if !isD {
		return nil, fmt.Errorf("invalid $push operator; expected a bson.D, got %v", args)
	}
	var updates []Update
	for _, e := range d {
		// Check that the value isn't a bunch of modifiers; not currently implemented
		if ed, eIsD := e.Value.(bson.D); eIsD {
			for _, e2 := range ed {
				if strings.HasPrefix(e2.Key, "$") {
					return nil, fmt.Errorf("modifiers not currently supported for $push operation; got %v", e2)
				}
			}
		}

		update, err := PushValue(e.Value)
		if err != nil {
			return nil, err
		}
		updates = append(updates, UpdatePath(e.Key, update))
	}
	return updates, nil
}

func parseAddToSet(args any) ([]Update, error) {
	d, isD := args.(bson.D)
	if !isD {
		return nil, fmt.Errorf("invalid $addToSet operator; expected a bson.D, got %v", args)
	}
	var updates []Update
	for _, e := range d {
		// Check that the value isn't a bunch of modifiers; not currently implemented
		if ed, eIsD := e.Value.(bson.D); eIsD {
			for _, e2 := range ed {
				if strings.HasPrefix(e2.Key, "$") {
					return nil, fmt.Errorf("modifiers not currently supported for $push operation; got %v", e2)
				}
			}
		}

		update, err := AddToSet(e.Value)
		if err != nil {
			return nil, err
		}
		updates = append(updates, UpdatePath(e.Key, update))
	}
	return updates, nil
}

func parsePull(args any) ([]Update, error) {
	d, isD := args.(bson.D)
	if !isD {
		return nil, fmt.Errorf("invalid $pull operator; expected a bson.D, got %v", args)
	}
	var updates []Update
	for _, e := range d {
		// Check that the value isn't a bunch of modifiers; not currently implemented
		if ed, eIsD := e.Value.(bson.D); eIsD {
			for _, e2 := range ed {
				if strings.HasPrefix(e2.Key, "$") {
					return nil, fmt.Errorf("modifiers not currently supported for $pull operation; got %v", e2)
				}
			}
		}

		var filter Filter
		var err error
		switch v := e.Value.(type) {
		case bson.D:
			filter, err = parseQuery(v)
		default:
			filter = Equals(v)
		}
		if err != nil {
			return nil, err
		}

		update, err := PullMatches(filter)
		if err != nil {
			return nil, err
		}
		updates = append(updates, UpdatePath(e.Key, update))
	}
	return updates, nil
}
