// Package gogen provides implementations of the builder interfaces defined by the [golang] plugin
//
// The builders are intended for use by plugins that define golang namespace nodes.  For example, the
// [goproc] plugin defines a process node that collects together golang instance nodes.  The [goproc]
// plugin then uses the builders defined here to accumulate the code declarations of those golang instances,
// and to generate the main file for the process.
//
// [golang]: https://github.com/Blueprint-uServices/blueprint/tree/main/plugins/golang
// [goproc]: https://github.com/Blueprint-uServices/blueprint/tree/main/plugins/goproc
package gogen
