package backend

import (
	"fmt"
	"reflect"
)

func GetPointerValue(val any) (any, error) {
	val_ptr := reflect.ValueOf(val)
	if val_ptr.Kind() != reflect.Pointer {
		return nil, fmt.Errorf("cannot indirect non-pointer type %v", val)
	}
	return reflect.Indirect(val_ptr).Interface(), nil
}

/*
Lots of APIs want to copy results into interfaces.  This is a helper method to do so.

src can be anything; dst must be a pointer to the same type as src
*/
func CopyResult(src any, dst any) error {
	dst_ptr := reflect.ValueOf(dst)
	if dst_ptr.Kind() != reflect.Pointer || dst_ptr.IsNil() {
		return fmt.Errorf("unable to copy result to type %v", reflect.TypeOf(dst))
	}
	dst_val := reflect.Indirect(dst_ptr)
	src_val := reflect.ValueOf(src)

	if dst_val.Kind() == reflect.Slice && src_val.Kind() == reflect.Slice {
		// Special handling for slices: we want to support copying []any to []mytype
		new_dst := reflect.MakeSlice(dst_val.Type(), src_val.Len(), src_val.Len())
		for i := 0; i < src_val.Len(); i++ {
			src_elem := src_val.Index(i).Interface()
			dst_elem := new_dst.Index(i).Addr().Interface()
			err := CopyResult(src_elem, dst_elem)
			if err != nil {
				return err
			}
		}
		dst_val.Set(new_dst)
		return nil
	} else {
		if !src_val.Type().AssignableTo(dst_val.Type()) {
			return fmt.Errorf("unable to copy incompatible types %v and %v", src_val.Type(), dst_val.Type())
		}
		dst_val.Set(src_val)
		return nil
	}
}

/*
Sets the zero value of a pointer
*/
func SetZero(dst any) error {
	receiver_ptr := reflect.ValueOf(dst)
	if receiver_ptr.Kind() != reflect.Pointer || receiver_ptr.IsNil() {
		return fmt.Errorf("unable to copy result to type %v", reflect.TypeOf(dst))
	}
	reflect.Indirect(receiver_ptr).SetZero()
	return nil
}
