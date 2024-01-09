package goproc

import (
	"github.com/Blueprint-uServices/blueprint/blueprint/pkg/blueprint/ioutil"
	"github.com/Blueprint-uServices/blueprint/blueprint/pkg/ir"
	"github.com/Blueprint-uServices/blueprint/plugins/golang"
)

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
