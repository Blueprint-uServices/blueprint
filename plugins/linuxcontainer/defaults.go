package linuxcontainer

import (
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/blueprint/ioutil"
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/ir"
	"github.com/blueprint-uservices/blueprint/plugins/linux"
)

/*
Registers a linux process workspace as the default way of combining and building processes
*/

// to trigger module initialization and register builders
func RegisterAsDefaultBuilder() {
	ir.RegisterDefaultNamespace[linux.Process]("linuxcontainer", buildDefaultLinuxWorkspace)
}

func buildDefaultLinuxWorkspace(outputDir string, nodes []ir.IRNode) error {
	ctr := newLinuxContainerNode("linuxprocesses")
	ctr.Nodes = nodes
	ctrDir, err := ioutil.CreateNodeDir(outputDir, "linuxprocesses")
	if err != nil {
		return err
	}
	return ctr.GenerateArtifacts(ctrDir)
}
