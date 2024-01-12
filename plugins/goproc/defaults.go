package goproc

import (
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/blueprint/ioutil"
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/ir"
	"github.com/blueprint-uservices/blueprint/plugins/golang"
)

// RegisterAsDefaultBuilder should be invoked by a wiring spec if it wishes to use goproc as the default
// way of combining golang instances.
//
// If you are using the [cmdbuilder], then goproc is automatically set as the default builder and you
// do not need to call this function.
//
// Default builders are responsible for building any golang instances that exist in a wiring spec but aren't
// explicitly added to a goproc within that wiring spec.  The Blueprint compiler groups these
// "floating" golang instances into a default golang process with the name "goproc".
//
// [cmdbuilder]: https://github.com/Blueprint-uServices/blueprint/tree/main/plugins/cmdbuilder
func RegisterAsDefaultBuilder() {
	/* any unattached golang nodes will be instantiated in a "default" golang workspace */
	ir.RegisterDefaultNamespace[golang.Node]("goproc", buildDefaultGolangWorkspace)
}

/*
If the Blueprint application contains any floating golang nodes, they get
built by this function.
*/
func buildDefaultGolangWorkspace(outputDir string, nodes []ir.IRNode) error {
	proc := newGolangProcessNode("golang")
	proc.Nodes = nodes
	procDir, err := ioutil.CreateNodeDir(outputDir, "golang")
	if err != nil {
		return err
	}
	return proc.GenerateArtifacts(procDir)
}
