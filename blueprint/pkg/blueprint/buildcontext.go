package blueprint

/*
Base interface used during build time; plugins that generate artifacts use this
*/
type (
	VisitTracker interface {
		Visited(name string) bool
	}

	BuildContext interface {
		VisitTracker
		ImplementsBuildContext()
	}
)

/*
In Blueprint, it is possible for there to be multiple different IRNode instances, across that application,
that generate and use the same code.  It is possible that the corresponding plugin does not want to generate
that code multiple times.  The VisitTracker provides a simple way for plugins to prevent generating output
code multiple times.

In methods where code gets generated (e.g. in golang Instantiable), before generating any code, plugins
invoke `VisitTracker.Visited` with a unique identifier (e.g. representing the node, instance, or plugin).
The first invocation for the identifier returns true; subsequent invocations return false.
*/
type VisitTrackerImpl struct {
	visited map[string]any
}

/*
Multiple instances of a node can exist across a Blueprint application that generates and uses the same code.
This method is used by nodes to determine whether code has already been generated in this workspace by a
different instance of the same node type.
The first call to this method for a given name will return false; subsequent calls will return true
*/
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
