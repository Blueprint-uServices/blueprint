# Developing a Blueprint Plugin

TODO

This doc describes how to develop a blueprint plugin.  Many concepts will need to be covered, probably split into multiple files:

 * Breakdown of the core components of a plugin:
   * Adding wiring spec functions in `wiring.go`
   * Implementing IR node(s) in `ir.go`
   * Implementing runtime components in `runtime`
   * Implementing code gen in `codegen` subdir
   * Adding the plugin to `plugins`
   * Package structure conventions
   * Emphasis on godoc documentation
   * Extending documentation listing the available plugins and their purposes
 * Explanation of what a wiring spec does
 * Explanation of what happens when building a wiring spec => IR
 * Explanation of code generation steps
 * Namespaces
 * Addresses
 * Config
 * Pointers
 