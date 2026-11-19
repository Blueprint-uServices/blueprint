package http

import "github.com/blueprint-uservices/blueprint/plugins/golang"

// The generated HTTP server imports these modules, so the generated module needs to require them.
// Without these requires, the go mod tidy that runs after code generation picks the newest
// releases, which can need a newer Go than the generated go.work and Dockerfile use (go 1.25).
// See requiredModules in the grpc plugin.
var requiredModules = []struct{ path, version string }{
	{"github.com/gorilla/mux", "v1.8.1"},
}

// Adds a go.mod require for each of requiredModules, once per generated module.
func requireModules(builder golang.ModuleBuilder) error {
	for _, mod := range requiredModules {
		if !builder.Visited(mod.path) {
			if err := builder.Require(mod.path, mod.version); err != nil {
				return err
			}
		}
	}
	return nil
}
