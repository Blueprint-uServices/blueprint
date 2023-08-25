package golang

import (
	"errors"
	"fmt"
	"os"
)

// Utility class to make sure we only build each artifact once
type VisitTracker struct {
	visited map[string]any
}

/*
Multiple instances of a node can exist across a Blueprint application that generates and uses the same code.
This method is used by nodes to determine whether code has already been generated in this workspace by a
different instance of the same node type.
The first call to this method for a given name will return false; subsequent calls will return true
*/
func (tracker *VisitTracker) Visited(name string) bool {
	_, has_visited := tracker.visited[name]
	if !has_visited {
		tracker.visited[name] = nil
	}
	return has_visited
}

// Returns true if the specified path exists and is a directory; false otherwise
func isDir(path string) bool {
	if info, err := os.Stat(path); err == nil && info.IsDir() {
		return true
	}
	return false
}

/*
Checks if the specified path exists and is a directory.
If `createIfAbsent` is true, then this will attempt to create the directory
*/
func checkDir(path string, createIfAbsent bool) error {
	if info, err := os.Stat(path); err == nil {
		if info.IsDir() {
			return nil
		} else {
			return fmt.Errorf("expected %s to be a directory but it is not", path)
		}
	} else if errors.Is(err, os.ErrNotExist) {
		if !createIfAbsent {
			return fmt.Errorf("expected directory %s but it does not exist", path)
		}
		err = os.MkdirAll(path, 0755)
		if err != nil {
			return fmt.Errorf("unable to create directory %s due to %s", path, err.Error())
		}
		return nil
	} else {
		return fmt.Errorf("unexpected error for directory %s due to %s", path, err.Error())
	}
}
