# Plugins

This module contains a range of plugins that implement most of Blueprint's features and functionality.

A large number of plugins are consolidated here for convenience; however to develop a plugin it is perfectly valid for it to reside in a different repository or module.

### Related documentation
- [../docs/manual/wiring.md](docs/manual/wiring.md) describes creating a wiring spec for an application, which will make use of plugins
- [../docs/manual/plugins.md](docs/manual/plugins.md) gives an overview of the available plugins and their functionality
- [../docs/manual/plugin_development.md](docs/manual/plugin_development.md) describes how to implement your own Blueprint plugin

### Other modules in this repository
- [../blueprint](../blueprint) implements Blueprint's compiler as well as the [WiringSpec API](../blueprint/pkg/wiring) used by Blueprint applications.
- [../runtime](../runtime) contains runtime code that is used by Blueprint applications and by plugins
- [../examples](../examples) contains some example applications