package ioutil

import (
	"errors"
	"os"
	"path/filepath"

	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/blueprint"
)

// Returns true if the specified path exists and is a directory; false otherwise
func IsDir(path string) bool {
	if info, err := os.Stat(path); err == nil && info.IsDir() {
		return true
	}
	return false
}

/*
Checks if the specified path exists and is a directory.
If `createIfAbsent` is true, then this will attempt to create the directory
*/
func CheckDir(path string, createIfAbsent bool) error {
	if info, err := os.Stat(path); err == nil {
		if info.IsDir() {
			return nil
		} else {
			return blueprint.Errorf("expected %s to be a directory but it is not", path)
		}
	} else if errors.Is(err, os.ErrNotExist) {
		if !createIfAbsent {
			return blueprint.Errorf("expected directory %s but it does not exist", path)
		}
		err = os.MkdirAll(path, 0755)
		if err != nil {
			return blueprint.Errorf("unable to create directory %s due to %s", path, err.Error())
		}
		return nil
	} else {
		return blueprint.Errorf("unexpected error for directory %s due to %s", path, err.Error())
	}
}

/*
Creates a subdirectory under the provided workspaceDir for the provided node.

The node's name is used to name the subdirectory (the node name is first cleaned).

# Returns the path to the subdirectory

Will return an error if the subdirectory already exists
*/
func CreateNodeDir(workspaceDir string, name string) (string, error) {
	nodeDir := filepath.Join(workspaceDir, blueprint.CleanName(name))
	if err := CheckDir(nodeDir, true); err != nil {
		return "", blueprint.Errorf("unable to create output dir for %v at %v due to %v", name, nodeDir, err.Error())
	}
	return nodeDir, nil
}
