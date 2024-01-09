package dockerdeployment

import (
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/blueprint/ioutil"
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/ir"
	"github.com/blueprint-uservices/blueprint/plugins/docker"
)

// to trigger module initialization and register builders
func RegisterAsDefaultBuilder() {
	ir.RegisterDefaultNamespace[docker.Container]("containerdeployment", buildDefaultContainerWorkspace)
}

func buildDefaultContainerWorkspace(outputDir string, nodes []ir.IRNode) error {
	ctr := &Deployment{DeploymentName: "docker", Nodes: nodes}
	subdir, err := ioutil.CreateNodeDir(outputDir, "docker")
	if err != nil {
		return err
	}
	return ctr.GenerateArtifacts(subdir)
}
