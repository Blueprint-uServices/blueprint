// Package rabbitmq provides a plugin to generate and include a rabbitmq instance in a Blueprint application.
//
// The package provides a built-in rabbitmq container that provides the server-side implementation
// and a go-client for connecting to the client.
//
// The applications must use a backend.Queue (runtime/core/backend) as the interface in the workflow.
package rabbitmq

import "gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/wiring"

// PrebuiltContainer generate the IRNodes for a mysql server docker container that uses the latest mysql/mysql image
// and the clients needed by the generated application to communicate with the server.
func PrebuiltContainer(spec wiring.WiringSpec, name string) string {
	// TODO: Implement
	return name
}
