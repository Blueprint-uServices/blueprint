---
title: blueprint/pkg/blueprint
---
# blueprint/pkg/blueprint
```go
package blueprint // import "gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/blueprint"
```
```go
Package blueprint provides some common utility methods for logging, io, errors
and string manipulation.
```
## CONSTANTS

```go
const (
	MAX_ERR_SIZE = 2048
)
```
## FUNCTIONS

## func Errorf
```go
func Errorf(format string, a ...any) error
```
Generates an error in the same way as fmt.Errorf but also includes the call
stack.

Plugins should generally use this method, because it enables us to more
easily tie errors back to the plugins and wiring specs that caused the
error.


