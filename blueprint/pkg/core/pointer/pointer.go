package pointer

import (
	"strings"

	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/blueprint"
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/core/address"
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

func CreatePointer(wiring blueprint.WiringSpec, name string, ptrType any, dst string) *PointerDef {
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
			// TODO: this only needs to happen once
			_, err := ptr.InstantiateDst(scope)
			return err
		})

		return node, nil
	})

	wiring.SetProperty(name, "ptr", ptr)

	return ptr
}

func IsPointer(wiring blueprint.WiringSpec, name string) bool {
	prop := wiring.GetProperty(name, "ptr")
	if prop == nil {
		return false
	}
	_, is_ptr := prop.(*PointerDef)
	return is_ptr
}

func GetPointer(wiring blueprint.WiringSpec, name string) *PointerDef {
	prop := wiring.GetProperty(name, "ptr")
	if prop != nil {
		if ptr, is_ptr := prop.(*PointerDef); is_ptr {
			return ptr
		}
	}
	return nil
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
	ptr.dstModifiers = append([]string{ptr.dstHead}, ptr.dstModifiers...)
	return nextDst
}

func (ptr *PointerDef) InstantiateDst(scope blueprint.Scope) (blueprint.IRNode, error) {
	scope.Info("Instantiating pointer %s.dst from scope %s", ptr.name, scope.Name())
	for _, modifier := range ptr.dstModifiers {
		node, err := scope.Get(modifier)
		if err != nil {
			return nil, err
		}

		addr, is_addr := node.(address.Address)
		if is_addr {
			dstName, err := address.DestinationOf(scope, modifier)
			if err != nil {
				return nil, err
			}
			if addr.GetDestination() != nil {
				// Destination has already been instantiated, stop instantiating now
				scope.Info("Destination %s of %s has already been instantiated", dstName, addr.Name())
				return nil, nil
			} else {
				dst, err := scope.Instantiate(dstName)
				if err != nil {
					return nil, err
				}
				err = addr.SetDestination(dst)
				if err != nil {
					return nil, err
				}
			}
		}
	}

	return scope.Get(ptr.dst)
}
