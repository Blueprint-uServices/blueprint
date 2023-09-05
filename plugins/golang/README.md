# plugins/golang

This contains some of the core golang-related IR interfaces and code parsing logic.

This package doesn't define any concrete IR node implementations; only interfaces representing golang nodes and golang code-generation nodes.

The [plugins/goproc](../goproc) plugin has implementations of some of these things