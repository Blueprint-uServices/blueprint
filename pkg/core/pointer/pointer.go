package pointer

import (
	"fmt"
	"strings"

	"gitlab.mpi-sws.org/cld/blueprint/pkg/blueprint"
)

type PointerDef struct {
	name         string
	srcHead      string
	srcModifiers []string
	srcTail      string
	dstHead      string
	dstModifiers []string
	dst          string
}

func (ptr PointerDef) String() string {
	b := strings.Builder{}
	b.WriteString("[")
	b.WriteString(strings.Join(ptr.srcModifiers, " -> "))
	b.WriteString("] -> [")
	b.WriteString(strings.Join(ptr.dstModifiers, " -> "))
	b.WriteString("]")
	return b.String()
}

func CreatePointer(wiring blueprint.WiringSpec, name string, ptrType any, dst string) (*PointerDef, error) {
	prop := wiring.GetProperty(name, "ptr")
	if prop != nil {
		_, is_ptr := prop.(*PointerDef)
		if !is_ptr {
			return nil, fmt.Errorf("%s already exists and has existing metadata for the \"ptr\" property but it is not valid pointer metadata", name)
		} else {
			return nil, fmt.Errorf("pointer %s already exists", name)
		}
	}

	ptr := &PointerDef{}
	ptr.name = name
	ptr.srcModifiers = nil
	ptr.srcHead = name + ".src"
	ptr.srcTail = ptr.srcHead
	ptr.dstHead = dst
	ptr.dstModifiers = nil
	ptr.dst = dst

	wiring.Alias(ptr.srcTail, ptr.dstHead)

	wiring.Define(name, ptrType, func(scope blueprint.Scope) (blueprint.IRNode, error) {
		node, err := scope.Get(ptr.srcHead)
		if err != nil {
			return nil, err
		}

		scope.Defer(func() error {
			_, err := ptr.instantiateDst(scope)
			return err
		})

		return node, nil
	})

	wiring.SetProperty(name, "ptr", ptr)

	return ptr, nil
}

func IsPointer(wiring blueprint.WiringSpec, name string) bool {
	prop := wiring.GetProperty(name, "ptr")
	if prop == nil {
		return false
	}
	_, is_ptr := prop.(*PointerDef)
	return is_ptr
}

func GetPointer(wiring blueprint.WiringSpec, name string) (*PointerDef, error) {
	prop := wiring.GetProperty(name, "ptr")
	if prop != nil {
		ptr, is_ptr := prop.(*PointerDef)
		if !is_ptr {
			return nil, fmt.Errorf("%s already exists and has existing metadata for the \"ptr\" property but it is not valid pointer metadata", name)
		}
		return ptr, nil
	}
	return nil, fmt.Errorf("cannot get pointer definition %s as it is has not been defined", name)
}

func (ptr *PointerDef) AddSrcModifier(wiring blueprint.WiringSpec, modifierName string) string {
	wiring.Alias(ptr.srcTail, modifierName)
	ptr.srcTail = modifierName + ".ptr.src.next"
	wiring.Alias(ptr.srcTail, ptr.dstHead)
	ptr.srcModifiers = append(ptr.srcModifiers, modifierName)

	return ptr.srcTail
}

func (ptr *PointerDef) AddDstModifier(wiring blueprint.WiringSpec, modifierName string) string {
	nextDst := ptr.dstHead
	ptr.dstHead = modifierName
	wiring.Alias(ptr.srcTail, ptr.dstHead)
	ptr.dstModifiers = append(ptr.dstModifiers, modifierName)
	return nextDst
}

func (ptr *PointerDef) instantiateDst(scope blueprint.Scope) (blueprint.IRNode, error) {
	_, err := scope.Get(ptr.dstHead)
	if err != nil {
		return nil, err
	}

	for _, modifier := range ptr.dstModifiers {
		_, err := scope.Get(modifier)
		if err != nil {
			return nil, err
		}
	}

	return scope.Get(ptr.dst)
}
