package dockerdeployment

import (
	"github.com/Blueprint-uServices/blueprint/blueprint/pkg/blueprint/ioutil"
	"github.com/Blueprint-uServices/blueprint/blueprint/pkg/ir"
	"github.com/Blueprint-uServices/blueprint/plugins/docker"
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
