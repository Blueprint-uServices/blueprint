package dockergen

import (
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/blueprint"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/docker"
)

/*
Implements the ContainerWorkspace defined in docker/ir.go

A basic implementation that:
 (a) gathers any locally-defined container artifacts
 (b) gathers container instance declarations
(c) generates a docker-compose file
*/

type DockerComposeWorkspace struct {
	blueprint.VisitTrackerImpl

	info docker.ContainerWorkspaceInfo

	ImageDirs map[string]string // map from image name to directory

}

type DockerComposeContainerInstance struct {
	Hostname string
}
