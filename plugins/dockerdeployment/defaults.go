package dockerdeployment

import (
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/ir"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/docker"
)

func init() {
	RegisterBuilders()
}

// to trigger module initialization and register builders
func RegisterBuilders() {
	ir.RegisterDefaultNamespace[docker.Container]("containerdeployment", buildDefaultContainerWorkspace)
}

func buildDefaultContainerWorkspace(outputDir string, nodes []ir.IRNode) error {
	ctr := newContainerDeployment("default")
	ctr.ContainedNodes = nodes
	return ctr.GenerateArtifacts(outputDir)
}
