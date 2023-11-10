package ir

import "gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/blueprint/stringutil"

func CleanName(name string) string {
	return stringutil.CleanName(name)
}

/*
A helper method to filter out nodes of a specific type from a slice of IRnodes
*/
func Filter[T any](nodes []IRNode) []T {
	var ts []T
	for _, node := range nodes {
		if t, isT := node.(T); isT {
			ts = append(ts, t)
		}
	}
	return ts
}

func FilterNodes[T any](nodes []IRNode) []IRNode {
	var ts []IRNode
	for _, node := range nodes {
		if _, isT := node.(T); isT {
			ts = append(ts, node)
		}
	}
	return ts
}

/*
Remove nodes of the given type
*/
func Remove[T any](nodes []IRNode) []IRNode {
	var remaining []IRNode
	for _, node := range nodes {
		if _, isT := node.(T); !isT {
			remaining = append(remaining, node)
		}
	}
	return remaining
}
