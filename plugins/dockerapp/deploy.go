package dockerapp

import (
	"fmt"

	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/blueprint"
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/core"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/docker"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/dockerapp/workspace"
	"golang.org/x/exp/slog"
)

func init() {
	RegisterBuilders()
}

// to trigger module initialization and register builders
func RegisterBuilders() {
	blueprint.RegisterDefaultNamespace[docker.Container]("containerdeployment", buildDefaultContainerWorkspace)
}

/*
Docker compose is the default docker app deployer.  It simply
generates a docker-compose.yml file on the local filesystem.
*/

type DockerCompose interface {
	core.ArtifactGenerator
}

func (node *Deployment) GenerateArtifacts(dir string) error {
	slog.Info(fmt.Sprintf("Collecting container instances for %s in %s", node.Name(), dir))
	workspace := workspace.NewDockerComposeWorkspace(node.Name(), dir)
	return node.generateArtifacts(workspace)
}

/*
The basic build process of a docker-compose deployment
*/
func (node *Deployment) generateArtifacts(workspace docker.ContainerWorkspace) error {
	// Add any locally-built container images
	for _, child := range node.ContainedNodes {
		if n, valid := child.(docker.ProvidesContainerImage); valid {
			if err := n.AddContainerArtifacts(workspace); err != nil {
				return err
			}
		}
	}

	// Collect all container instances
	for _, child := range node.ContainedNodes {
		if n, valid := child.(docker.ProvidesContainerInstance); valid {
			if err := n.AddContainerInstance(workspace); err != nil {
				return err
			}
		}
	}

	return workspace.Finish()
}

func buildDefaultContainerWorkspace(outputDir string, nodes []blueprint.IRNode) error {
	ctr := newContainerDeployment("default")
	ctr.ContainedNodes = nodes
	return ctr.GenerateArtifacts(outputDir)
}
