package grpc

import "github.com/blueprint-uservices/blueprint/plugins/golang"

// The generated gRPC code imports these modules, so the generated module needs to require them.
// Without these requires, the go mod tidy that runs after code generation picks the newest
// releases, and those can need a newer Go than the generated go.work and Dockerfile use
// (go 1.25, set in golang/gogen/workspacebuilder.go and goproc/linuxgen/dockerfile_buildcommands.go).
//
// These are minimum versions. grpc v1.68.1 is the newest release whose go.mod doesn't require
// otel, so generated apps keep otel at v1.26.0, the version the opentelemetry plugin requires.
// It is also new enough for code from protoc-gen-go-grpc v1.5 and later, which needs grpc v1.64
// or later. Tested with protoc-gen-go v1.34.0 and v1.36.12 and protoc-gen-go-grpc v1.3.0 and
// v1.6.2. Newer generators may need these raised.
var requiredModules = []struct{ path, version string }{
	{"google.golang.org/grpc", "v1.68.1"},
	{"google.golang.org/protobuf", "v1.36.12"},
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
