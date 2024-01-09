// Package ioutil implements filesystem related utility methods primarily for use by
// plugins that produce artifacts onto the local filesystem.
package ioutil

import (
	"errors"
	"os"
	"path/filepath"

	"github.com/blueprint-uservices/blueprint/blueprint/pkg/blueprint"
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/blueprint/stringutil"
)

// Returns true if the specified path exists and is a directory; false otherwise
func IsDir(path string) bool {
	if info, err := os.Stat(path); err == nil && info.IsDir() {
		return true
	}
	return false
}

// Returns nil if the specified path exists and is a directory; if not returns an error.
// If the specified path does not exist, then createIfAbsent dictates whether the
// path is either created, or an error is returned.
// This method can also return an error if it was unable to create a directory at
// the given path.
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

// Creates a subdirectory in the provided workspaceDir.
// The provided name is first sanitized using [stringutil.CleanName]
func CreateNodeDir(workspaceDir string, name string) (string, error) {
	nodeDir := filepath.Join(workspaceDir, stringutil.CleanName(name))
	if err := CheckDir(nodeDir, true); err != nil {
		return "", blueprint.Errorf("unable to create output dir for %v at %v due to %v", name, nodeDir, err.Error())
	}
	return nodeDir, nil
}
