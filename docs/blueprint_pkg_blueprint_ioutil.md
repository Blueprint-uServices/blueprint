---
title: blueprint/pkg/blueprint/ioutil
---
# blueprint/pkg/blueprint/ioutil
```go
package ioutil // import "gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/blueprint/ioutil"
```
```go
Package ioutil implements filesystem related utility methods primarily for use
by plugins that produce artifacts onto the local filesystem.
```
## FUNCTIONS

## func CheckDir
```go
func CheckDir(path string, createIfAbsent bool) error
```
Returns nil if the specified path exists and is a directory; if not returns
an error. If the specified path does not exist, then createIfAbsent dictates
whether the path is either created, or an error is returned. This method
can also return an error if it was unable to create a directory at the given
path.

## func CreateNodeDir
```go
func CreateNodeDir(workspaceDir string, name string) (string, error)
```
Creates a subdirectory in the provided workspaceDir. The provided name is
first sanitized using stringutil.CleanName

## func IsDir
```go
func IsDir(path string) bool
```
Returns true if the specified path exists and is a directory; false
otherwise


