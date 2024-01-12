package dockercompose

import (
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/blueprint/ioutil"
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/ir"
	"github.com/blueprint-uservices/blueprint/plugins/docker"
)

// RegisterAsDefaultBuilder should be invoked by a wiring spec if it wishes to use docker-compose as the default
// way of combining container instances.
//
// If you are using the [cmdbuilder], then docker-compose is automatically set as the default builder and you
// do not need to call this function again.
//
// Default builders are responsible for building any container instances that exist in a wiring spec but aren't
// explicitly added to a container deployment within that wiring spec.  The Blueprint compiler groups these
// "floating" container instances into a default dockercompose deployment with the name "docker".
//
// [cmdbuilder]: https://github.com/Blueprint-uServices/blueprint/tree/main/plugins/cmdbuilder
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
