package pointer

import (
	"strings"

	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/core/address"
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/ir"
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/wiring"
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

func CreatePointer(spec wiring.WiringSpec, name string, ptrType any, dst string) *PointerDef {
	ptr := &PointerDef{}
	ptr.name = name
	ptr.srcModifiers = nil
	ptr.srcHead = name + ".src"
	ptr.srcTail = ptr.srcHead
	ptr.dstHead = dst
	ptr.dstModifiers = nil
	ptr.dst = dst

	spec.Alias(ptr.srcTail, ptr.dstHead)

	spec.Define(name, ptrType, func(namespace wiring.Namespace) (ir.IRNode, error) {
		var node ir.IRNode
		if err := namespace.Get(ptr.srcHead, &node); err != nil {
			return nil, err
		}

		namespace.Defer(func() error {
			_, err := ptr.InstantiateDst(namespace)
			return err
		})

		return node, nil
	})

	spec.SetProperty(name, "ptr", ptr)

	return ptr
}

func IsPointer(spec wiring.WiringSpec, name string) bool {
	var ptr *PointerDef
	return spec.GetProperty(name, "ptr", &ptr) == nil
}

func GetPointer(spec wiring.WiringSpec, name string) *PointerDef {
	var ptr *PointerDef
	spec.GetProperty(name, "ptr", &ptr)
	return ptr
}

func (ptr *PointerDef) AddSrcModifier(spec wiring.WiringSpec, modifierName string) string {
	spec.Alias(ptr.srcTail, modifierName)
	ptr.srcTail = modifierName + ".ptr.src.next"
	spec.Alias(ptr.srcTail, ptr.dstHead)
	ptr.srcModifiers = append(ptr.srcModifiers, modifierName)

	return ptr.srcTail
}

func (ptr *PointerDef) AddDstModifier(spec wiring.WiringSpec, modifierName string) string {
	nextDst := ptr.dstHead
	ptr.dstHead = modifierName
	spec.Alias(ptr.srcTail, ptr.dstHead)
	ptr.dstModifiers = append([]string{ptr.dstHead}, ptr.dstModifiers...)
	return nextDst
}

func (ptr *PointerDef) InstantiateDst(namespace wiring.Namespace) (ir.IRNode, error) {
	namespace.Info("Instantiating pointer %s.dst from namespace %s", ptr.name, namespace.Name())
	for _, modifier := range ptr.dstModifiers {
		var addr address.Node
		err := namespace.Get(modifier, &addr)

		// Want to find the final dstModifier that points to an address, then instantiate the address
		if err == nil {
			dstName, err := address.DestinationOf(namespace, modifier)
			if err != nil {
				return nil, err
			}
			if addr.GetDestination() != nil {
				// Destination has already been instantiated, stop instantiating now
				namespace.Info("Destination %s of %s has already been instantiated", dstName, addr.Name())
				return nil, nil
			} else {
				namespace.Info("Instantiating %s of %s", dstName, addr.Name())
				var dst ir.IRNode
				if err := namespace.Instantiate(dstName, &dst); err != nil {
					return nil, err
				}
				err = addr.SetDestination(dst)
				if err != nil {
					return nil, err
				}
			}
		} else {
			namespace.Info("Skipping %v, not an address", modifier)
		}
	}

	var node ir.IRNode
	err := namespace.Get(ptr.dst, &node)
	return node, err
}
