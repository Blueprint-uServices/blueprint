package linuxcontainer

import (
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/blueprint/ioutil"
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/ir"
	"github.com/blueprint-uservices/blueprint/plugins/linux"
)

// RegisterAsDefaultBuilder should be invoked by a wiring spec if it wishes to use linuxcontainer as the default
// way of combining process instances.
//
// If you are using the [cmdbuilder], then linuxcontainer is automatically set as the default builder and you
// do not need to call this function.
//
// Default builders are responsible for building any process instances that exist in a wiring spec but aren't
// explicitly added to a container within that wiring spec.  The Blueprint compiler groups these
// "floating" process instances into a default linux container with the name "linux".
//
// [cmdbuilder]: https://github.com/Blueprint-uServices/blueprint/tree/main/plugins/cmdbuilder
func RegisterAsDefaultBuilder() {
	ir.RegisterDefaultNamespace[linux.Process]("linux", buildDefaultLinuxWorkspace)
}

func buildDefaultLinuxWorkspace(outputDir string, nodes []ir.IRNode) error {
	ctr := newLinuxContainerNode("linux")
	ctr.Nodes = nodes
	ctrDir, err := ioutil.CreateNodeDir(outputDir, "linux")
	if err != nil {
		return err
	}
	return ctr.GenerateArtifacts(ctrDir)
}
