---
title: blueprint/pkg/blueprint/logging
---
# blueprint/pkg/blueprint/logging
```go
package logging // import "gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/blueprint/logging"
```
```go
Package logging implements a custom slog logger for Blueprint.
```
```go
The custom logger adds additional callsite information to logging statements, to
provide more information during the compilation process about which plugins are
producing logs or errors, and to tie that information back to the corresponding
wiring line.
```
## FUNCTIONS

## func DisableCompilerLogging
```go
func DisableCompilerLogging()
```
Disables logging by the compiler; useful when running tests to suppress
verbose output.

## func EnableCompilerLogging
```go
func EnableCompilerLogging()
```
Compiler logging is enabled by default; this method is useful for tests to
disable and enable logging in order to suppress output.


## TYPES

Used to tie logging statements and errors back to the wiring file line that
caused the error
```go
type Callsite struct {
	Source     *sourceFileInfo
	LineNumber int
	Func       string
	FuncName   string
}
```
## func 
```go
func (cs Callsite) String() string
```

Used to tie logging statements and errors back to the wiring file line that
caused the error
```go
type Callstack struct {
	Stack []Callsite
}
```
## func GetCallstack
```go
func GetCallstack() *Callstack
```
Gets the current callstack including file information. Blueprint's wiring
spec uses this so that logging statements and error messages can be
attributed back to the appropriate wiring spec line.

## func 
```go
func (stack *Callstack) String() string
```


