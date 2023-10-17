package goproc

/*
Goprocs can be deployed to Docker containers, in which case the deployment
process is the same as for Linux, but we can also provide the appropriate
environment for building and running the Goproc with Docker-specific
build commands
*/

type DockerGoProc interface {
	LinuxGoProc
	// docker.ProvidesBuildCommands
}
