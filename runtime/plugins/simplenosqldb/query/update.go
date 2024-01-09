package query

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/blueprint-uservices/blueprint/runtime/core/backend"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/bsontype"
)

type Update interface {
	Apply(itemRef any) error
	String() string
}

type (
	set struct {
		t     bsontype.Type
		value []byte
	}

	push struct {
		t     bsontype.Type
		value []byte
	}

	addtoset struct {
		t     bsontype.Type
		value []byte
	}

	pull struct {
		filter Filter
	}

	unsetfield struct {
		fieldName string
	}

	unsetelement struct {
		index int
	}

	incint struct {
		amount int64
	}

	updatefield struct {
		fieldName      string
		createIfAbsent bool
		update         Update
	}

	updateindex struct {
		index          int
		createIfAbsent bool
		update         Update
	}

	updateall struct {
		updates []Update
	}

	broadcastupdate struct {
		update Update
	}
)

func SetValue(value any) (Update, error) {
	t, v, err := bson.MarshalValue(value)
	return &set{t: t, value: v}, err
}

func PushValue(value any) (Update, error) {
	t, v, err := bson.MarshalValue(value)
	return &push{t: t, value: v}, err
}

func AddToSet(value any) (Update, error) {
	t, v, err := bson.MarshalValue(value)
	return &addtoset{t: t, value: v}, err
}

func PullMatches(filter Filter) (Update, error) {
	return &pull{filter: filter}, nil
}

func UnsetField(fieldName string) Update {
	return &unsetfield{fieldName: fieldName}
}

func UnsetElement(index int) Update {
	return &unsetelement{index: index}
}

func IncInt(amount int64) Update {
	return &incint{amount: amount}
}

func UpdateField(fieldName string, update Update, createIfAbsent bool) Update {
	return &updatefield{fieldName: fieldName, update: update, createIfAbsent: createIfAbsent}
}

func UpdateIndex(index int, update Update, createIfAbsent bool) Update {
	return &updateindex{index: index, update: update, createIfAbsent: createIfAbsent}
}

func UpdatePath(selector string, update Update) Update {
	splits := strings.Split(selector, ".")
	for i := len(splits) - 1; i >= 0; i-- {
		if j, err := strconv.Atoi(splits[i]); err == nil {
			update = UpdateIndex(j, update, true)
		} else {
			update = UpdateField(splits[i], update, true)
			if i != 0 {
				update = BroadcastUpdate(update) // broadcast in case the field is an array type
			}
		}
	}
	return update
}

func UnsetPath(selector string) Update {
	splits := strings.Split(selector, ".")

	var update Update
	if i, err := strconv.Atoi(splits[len(splits)-1]); err == nil {
		update = UnsetElement(i)
	} else {
		update = UnsetField(splits[len(splits)-1])
	}

	for i := len(splits) - 2; i >= 0; i-- {
		if j, err := strconv.Atoi(splits[i]); err == nil {
			update = UpdateIndex(j, update, false)
		} else {
			update = UpdateField(splits[i], update, false)
			if i != 0 {
				update = BroadcastUpdate(update) // broadcast in case the field is an array type
			}
		}
	}
	return update
}

func UpdateAll(updates []Update) Update {
	all := &updateall{}
	for _, update := range updates {
		switch u := update.(type) {
		case *updateall:
			all.updates = append(all.updates, u.updates...)
		default:
			all.updates = append(all.updates, update)
		}
	}
	return all
}

func BroadcastUpdate(update Update) Update {
	return &broadcastupdate{update: update}
}

func (s *set) Apply(itemRef any) error {
	var v any
	err := bson.UnmarshalValue(s.t, s.value, &v)
	if err != nil {
		return err
	}

	dst_ptr := reflect.ValueOf(itemRef)
	if dst_ptr.Kind() != reflect.Pointer || dst_ptr.IsNil() {
		return fmt.Errorf("set unable to apply update to non-pointer type %v", reflect.TypeOf(itemRef))
	}
	dst_val := reflect.Indirect(dst_ptr)

	if dst_val.Kind() != reflect.Interface {
		return fmt.Errorf("set unable to apply update to non-interface type %v", reflect.TypeOf(dst_val))
	}

	dst_val.Set(reflect.ValueOf(v))

	return nil
}

func (p *push) Apply(itemRef any) error {
	var v any
	err := bson.UnmarshalValue(p.t, p.value, &v)
	if err != nil {
		return err
	}

	itemVal, err := backend.GetPointerValue(itemRef)
	if err != nil {
		return err
	}

	if itemVal == nil {
		itemVal = bson.A{}
	}

	a, isA := itemVal.(bson.A)
	if !isA {
		return fmt.Errorf("push expected a bson.A but instead found a %v %v", reflect.TypeOf(itemVal), itemVal)
	}

	a = append(a, v)
	return backend.CopyResult(a, itemRef)
}

func (p *addtoset) Apply(itemRef any) error {
	var v any
	err := bson.UnmarshalValue(p.t, p.value, &v)
	if err != nil {
		return err
	}

	itemVal, err := backend.GetPointerValue(itemRef)
	if err != nil {
		return err
	}

	if itemVal == nil {
		itemVal = bson.A{}
	}

	a, isA := itemVal.(bson.A)
	if !isA {
		return fmt.Errorf("expected a bson.A but instead found a %v %v", reflect.TypeOf(itemVal), itemVal)
	}

	// Check if it exists
	for i := range a {
		if reflect.DeepEqual(v, a[i]) {
			return nil
		}
	}
	a = append(a, v)

	return backend.CopyResult(a, itemRef)
}

func (p *pull) Apply(itemRef any) error {
	itemVal, err := backend.GetPointerValue(itemRef)
	if err != nil {
		return err
	}

	if itemVal == nil {
		itemVal = bson.A{}
	}

	a, isA := itemVal.(bson.A)
	if !isA {
		return fmt.Errorf("pull expected a bson.A but instead found a %v %v", reflect.TypeOf(itemVal), itemVal)
	}

	var updated bson.A
	updated = nil
	for i := range a {
		if p.filter.Apply(a[i]) {
			if updated == nil {
				updated = append(bson.A{}, a[:i]...)
			}
		} else if updated != nil {
			updated = append(updated, a[i])
		}
	}

	if updated != nil {
		return backend.CopyResult(updated, itemRef)
	}
	return nil
}

func (u *unsetfield) Apply(itemRef any) error {
	itemVal, err := backend.GetPointerValue(itemRef)
	if err != nil || itemVal == nil {
		return err
	}

	v, isD := itemVal.(bson.D)
	if !isD {
		return fmt.Errorf("unsetfield expected a bson.D but instead found a %v %v", reflect.TypeOf(itemVal), itemVal)
	}

	// See if the key exists within the document; if so remove it
	for i := range v {
		if v[i].Key == u.fieldName {
			v = append(v[:i], v[i+1:]...)
			return backend.CopyResult(v, itemRef)
		}
	}
	return nil
}

func (u *unsetelement) Apply(itemRef any) error {
	itemVal, err := backend.GetPointerValue(itemRef)
	if err != nil || itemVal == nil {
		return err
	}

	v, isA := itemVal.(bson.A)
	if !isA {
		return fmt.Errorf("unsetelement expected a bson.A but instead found a %v %v", reflect.TypeOf(itemVal), itemVal)
	}

	// Ensure the slice length
	if u.index < len(v) {
		v[u.index] = nil
	}
	return nil
}

func (i *incint) Apply(itemRef any) error {
	itemVal, err := backend.GetPointerValue(itemRef)
	if err != nil {
		return err
	}

	if itemVal == nil {
		itemVal = int64(0)
	}

	switch v := itemVal.(type) {
	case float64:
		return backend.CopyResult(v+float64(i.amount), itemRef)
	case float32:
		return backend.CopyResult(v+float32(i.amount), itemRef)
	case int64:
		return backend.CopyResult(v+i.amount, itemRef)
	case int32:
		return backend.CopyResult(v+int32(i.amount), itemRef)
	case int16:
		return backend.CopyResult(v+int16(i.amount), itemRef)
	case int8:
		return backend.CopyResult(v+int8(i.amount), itemRef)
	case int:
		return backend.CopyResult(v+int(i.amount), itemRef)
	default:
		return fmt.Errorf("incint unable to increment non-integer %v", itemVal)
	}
}

func (s *updatefield) Apply(itemRef any) error {
	itemVal, err := backend.GetPointerValue(itemRef)
	if err != nil {
		return err
	}

	if itemVal == nil {
		itemVal = bson.D{}
	}

	v, isD := itemVal.(bson.D)
	if !isD {
		return fmt.Errorf("updatefield expected a bson.D but instead found a %v %v", reflect.TypeOf(itemVal), itemVal)
	}

	// See if the key exists within the document; if so update in place
	for i := range v {
		if v[i].Key == s.fieldName {
			return s.update.Apply(&v[i].Value)
		}
	}

	// Key does not exist within the document; add new element and copy back to ptr
	if !s.createIfAbsent {
		return nil
	}
	v = append(v, bson.E{Key: s.fieldName})
	err = s.update.Apply(&v[len(v)-1].Value)
	if err != nil {
		return err
	}
	return backend.CopyResult(v, itemRef)
}

func (s *updateindex) Apply(itemRef any) error {
	itemVal, err := backend.GetPointerValue(itemRef)
	if err != nil {
		return err
	}

	if itemVal == nil {
		if !s.createIfAbsent {
			return nil
		}
		itemVal = bson.A{}
	}

	v, isA := itemVal.(bson.A)
	if !isA {
		return fmt.Errorf("updateindex expected a bson.A but instead found a %v %v", reflect.TypeOf(itemVal), itemVal)
	}

	// Ensure the slice length
	if !s.createIfAbsent && len(v) <= s.index {
		return nil
	}
	if cap(v) <= s.index {
		newv := make(bson.A, 0, s.index+1)
		newv = append(newv, v...)
		v = newv
	}
	v = v[:s.index+1]

	// Update at the specified index and copy back to ptr
	err = s.update.Apply(&v[s.index])
	if err != nil {
		return err
	}
	return backend.CopyResult(v, itemRef)

}

func (a *updateall) Apply(itemRef any) error {
	for _, update := range a.updates {
		if err := update.Apply(itemRef); err != nil {
			return err
		}
	}
	return nil
}

func (b *broadcastupdate) Apply(itemRef any) error {
	item_ptr := reflect.ValueOf(itemRef)
	if item_ptr.Kind() != reflect.Pointer {
		return fmt.Errorf("broadcastupdate expect ptr type but got %v", reflect.TypeOf(itemRef))
	}
	item_val := reflect.Indirect(item_ptr)

	if item_val.Kind() != reflect.Slice {
		return b.update.Apply(itemRef)
	}

	for i := 0; i < item_val.Len(); i++ {
		err := b.update.Apply(item_val.Index(i).Addr().Interface())
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *set) String() string {
	var v interface{}
	bson.UnmarshalValue(s.t, s.value, &v)
	return fmt.Sprintf("set %v", v)
}

func (p *push) String() string {
	var v interface{}
	bson.UnmarshalValue(p.t, p.value, &v)
	return fmt.Sprintf("push %v", v)
}

func (p *addtoset) String() string {
	var v interface{}
	bson.UnmarshalValue(p.t, p.value, &v)
	return fmt.Sprintf("addtoset %v", v)
}

func (p *pull) String() string {
	return fmt.Sprintf("pull: %v", p.filter)
}

func (s *unsetfield) String() string {
	return fmt.Sprintf("unset %v", s.fieldName)
}

func (s *unsetelement) String() string {
	return fmt.Sprintf("unset %v", s.index)
}

func (i *incint) String() string {
	return fmt.Sprintf(" += %v", i.amount)
}

func (u *updatefield) String() string {
	return fmt.Sprintf(".%v %v", u.fieldName, u.update)
}

func (u *updateindex) String() string {
	return fmt.Sprintf("[%v] %v", u.index, u.update)
}

func (a *updateall) String() string {
	var strs []string
	for _, update := range a.updates {
		strs = append(strs, fmt.Sprintf("%v", update))
	}
	return strings.Join(strs, "; ")
}

func (b *broadcastupdate) String() string {
	return fmt.Sprintf("broadcast %v ", b.update)
}
