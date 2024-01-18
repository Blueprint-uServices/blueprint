// Package environment provides a plugin for generating a .env file in the root Blueprint output directory that automatically sets
// address configuration variables (hostnames and ports for dial and bind addresses).
//
// The plugin is intended for convenience so that Blueprint users do not have to manually allocate ports and pass them as
// environment variables.  However, in more complex deployment, Blueprint users may wish to disable this plugin to afford themselves
// more control.
//
// # Wiring Spec Usage
//
// The environment plugin is automatically enabled if you are using the [cmdbuilder].  See the cmdbuilder documentation for
// command-line flags you can pass to control environment generation.  Otherwise you can manually use the environment plugin
// from your wiring spec as follows
//
//	environment.AssignPorts(spec, 12345)
//
// Ports will be automatically assigned to services starting from 12345 and incrementing.
//
// # Generated Artifacts
//
// The plugin will generate several .env files to the root output directory.  The .env files take the form:
//
//	USER_SERVICE_GRPC_DIAL_ADDR=user_service:12345
//	USER_SERVICE_GRPC_BIND_ADDR=0.0.0.0:12345
//
// The plugin generates two different env files:
//   - .local.env assumes all services will be deployed on a single machine; it uses localhost for dial hostnames and 0.0.0.0 for
//     bind hostnames, e.g. localhost:12345 and 0.0.0.0:12345
//   - .distributed.env uses the service name as dial hostname and 0.0.0.0 for bind hostname, e.g. user_service:12345 and 0.0.0.0:12345
//
// # Running Artifacts
//
// Before running the application or a client, you can source one of the .env files to avoid having to manually set
// environment variables or command line arguments.
//
// For example, if you are running a docker-compose deployment, you can run:
//
//	cd build
//	source .local.env
//	cd docker
//	docker-compose up
//
// Similarly, workload generator clients and tests will check environment variables for default values.
//
// If you are using .distributed.env then the hostnames for services will need to be mapped in your /etc/hosts file
//
// The plugin does not guarantee that the ports (e.g. 12345) are actually available for use on any machine.  This is up to the user.
package environment
