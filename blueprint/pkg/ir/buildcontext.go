package ir

type (
	// A Blueprint application can potentially have multiple IR node instances spread across the application
	// that generate the same code.
	//
	// Visit tracker is a utility method used during artifact generation to prevent nodes from unnecessarily
	// generating the same artifact repeatedly, when once will suffice.
	VisitTracker interface {
		// Returns false on the first invocation of name; true on subsequent invocations
		Visited(name string) bool
	}

	// All artifact generation occurs in the context of some BuildContext.
	//
	// Plugins that control the artifact generation process should implement this interface.
	BuildContext interface {
		VisitTracker
		ImplementsBuildContext()
	}
)

// Basic implementation of the [VisitTracker] interface
type VisitTrackerImpl struct {
	visited map[string]any
}

func (tracker *VisitTrackerImpl) Visited(name string) bool {
	if tracker.visited == nil {
		tracker.visited = make(map[string]any)
	}
	_, has_visited := tracker.visited[name]
	if !has_visited {
		tracker.visited[name] = nil
	}
	return has_visited
}
