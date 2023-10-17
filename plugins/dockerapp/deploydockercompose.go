package dockerapp

import "gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/core"

/*
Docker compose is the default docker app deployer.  It simply
generates a docker-compose.yml file on the local filesystem.
*/

type DockerCompose interface {
	core.ArtifactGenerator
}

func (node *Deployment) GenerateArtifacts(dir string) error {
	// builder, err := dockergen.NewDockerComposeBuilder(dir)
	// if err != nil {
	// 	return err
	// }

	// for _, node := range node.ContainedNodes {
	// 	if n, valid := node.(docker.BuildsContainerImage); valid {
	// 		if err := n.PrepareDockerfile(builder); err != nil {
	// 			return err
	// 		}
	// 	}
	// }

	// for _, node := range node.ContainedNodes {
	// 	if n, valid := node.(docker.Container); valid {
	// 		if err := n.AddToDockerCompose(builder); err != nil {
	// 			return err
	// 		}
	// 	}
	// }
	return nil
}
