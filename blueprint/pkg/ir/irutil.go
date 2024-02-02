package ir

import (
	"sort"
	"strings"

	"github.com/blueprint-uservices/blueprint/blueprint/pkg/blueprint/stringutil"
)

// Returns name with only alphanumeric characters and all other
// symbols converted to underscores.
//
// CleanName is primarily used by plugins to convert user-defined
// service names into names that are valid as e.g. environment variables,
// command line arguments, etc.
func CleanName(name string) string {
	return stringutil.CleanName(name)
}

// Reports whether nodeType is an instance of type T
func Is[T any](nodeType any) bool {
	_, isT := nodeType.(T)
	return isT
}

// Returns a slice containing only nodes of type T
func Filter[T any](nodes []IRNode) []T {
	var ts []T
	for _, node := range nodes {
		if t, isT := node.(T); isT {
			ts = append(ts, t)
		}
	}
	return ts
}

// Returns a slice containing only nodes of type T,
// and a slice containing all other nodes
func Split[T any](nodes []IRNode) (remaining []IRNode, matches []T) {
	for _, node := range nodes {
		if t, isT := node.(T); isT {
			matches = append(matches, t)
		} else {
			remaining = append(remaining, node)
		}
	}
	return
}

// Returns a slice containing only nodes of type T
func FilterNodes[T any](nodes []IRNode) []IRNode {
	var ts []IRNode
	for _, node := range nodes {
		if _, isT := node.(T); isT {
			ts = append(ts, node)
		}
	}
	return ts
}

// Returns a slice containing all nodes except those of type T
func Remove[T any](nodes []IRNode) []IRNode {
	var remaining []IRNode
	for _, node := range nodes {
		if _, isT := node.(T); !isT {
			remaining = append(remaining, node)
		}
	}
	return remaining
}

func PrettyPrintNamespace(instanceName string, namespaceType string, argNodes []IRNode, childNodes []IRNode) string {
	var b strings.Builder
	b.WriteString(instanceName)
	b.WriteString(" = ")
	b.WriteString(namespaceType)
	b.WriteString("(")
	var args sort.StringSlice
	for _, arg := range argNodes {
		args = append(args, arg.Name())
	}
	args.Sort()
	b.WriteString(strings.Join(args, ", "))
	b.WriteString(") {\n")
	var children sort.StringSlice
	for _, child := range childNodes {
		children = append(children, child.String())
	}
	children.Sort()
	b.WriteString(stringutil.Indent(strings.Join(children, "\n"), 2))
	b.WriteString("\n}")
	return b.String()
}
