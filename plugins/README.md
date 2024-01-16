# Plugins

Much of Blueprint's features and functionality are implemented by plugins contained in this module.

Each plugin resides in its own subdirectory and provides its own documentation.

The Blueprint [User Manual](../docs/manual/plugins.md) provides a high-level overview of some of the more prominent plugins.

## WiringSpec

Most plugins make use of a blueprint WiringSpec from the [blueprint/pkg/wiring](../blueprint/pkg/wiring) package.  See the documentation for that package or from the Blueprint [User Manual](../docs/manual/) for details on initializing a wiring spec.

## Developing Plugins

If you want to develop your own plugin, it is not mandatory for the plugin to live in this directory.  It can reside in a different repository and module.