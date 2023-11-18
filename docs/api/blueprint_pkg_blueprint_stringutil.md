---
title: blueprint/pkg/blueprint/stringutil
---
# blueprint/pkg/blueprint/stringutil
```go
package stringutil // import "gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/blueprint/stringutil"
```
```go
Package stringutil implements utility methods for common string manipulations
used by Blueprint plugins.
```
## FUNCTIONS

## func Capitalize
```go
func Capitalize(s string) string
```
Returns s with the first letter converted to uppercase.

## func CleanName
```go
func CleanName(name string) string
```
Returns name with only alphanumeric characters and all other symbols
converted to underscores.

CleanName is primarily used by plugins to convert user-defined service
names into names that are valid as e.g. environment variables, command line
arguments, etc.

## func Indent
```go
func Indent(str string, amount int) string
```
Indents str by the specified amount by prepending whitespace (space
characters) at the beginning of every line. str can be a multi-line string;
whitespace will be prepended to every line.

## func Reindent
```go
func Reindent(str string, amount int) string
```
Indents str by the specified amount by adding or removing whitespace (space
characters) at the beginning of every line. If str is already indented by
some amount, then this method will add or remove whitespace so that the
result is indented by amount total whitespace.

## func ReplaceSuffix
```go
func ReplaceSuffix(s string, suffix string, replacement string) string
```
If s ends with the provided suffix, removes the suffix from s and replaces
it with replacement. If s does not end with the provided suffix, then simply
returns s . replacement


