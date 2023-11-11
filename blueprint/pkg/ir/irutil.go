package ir

import "gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/blueprint/stringutil"

// Returns name with only alphanumeric characters and all other
// symbols converted to underscores.
//
// CleanName is primarily used by plugins to convert user-defined
// service names into names that are valid as e.g. environment variables,
// command line arguments, etc.
func CleanName(name string) string {
	return stringutil.CleanName(name)
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
