package kubepod

import (
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/ir"
	"github.com/blueprint-uservices/blueprint/plugins/docker"
)

// A Kubernetes pod deployer. It generates the pod config files on the local filesystem.
type kubePod interface {
	ir.ArtifactGenerator
}

// A workspace used when deploying a set of containers as a Kubernetes Pod
//
// Implements docker.ContainerWorkspace defined in docker/ir.go
//
// This workspace generates Pod files at the root of the output directory.
type kubePodWorkspace struct {
	ir.VisitTrackerImpl

	info docker.ContainerWorkspaceInfo

	ImageDirs map[string]string
}
