package goproc

import (
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/blueprint/ioutil"
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/ir"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/golang"
)

func RegisterAsDefaultBuilder() {
	/* any unattached golang nodes will be instantiated in a "default" golang workspace */
	ir.RegisterDefaultNamespace[golang.Node]("goproc", buildDefaultGolangWorkspace)
	ir.RegisterDefaultBuilder[*Process]("goproc", buildDefaultGolangProcess)
}

/*
If the Blueprint application contains any floating golang nodes, they get
built by this function.
*/
func buildDefaultGolangWorkspace(outputDir string, nodes []ir.IRNode) error {
	proc := newGolangProcessNode("default")
	proc.ContainedNodes = nodes
	return proc.GenerateArtifacts(outputDir)
}

/*
If the Blueprint application contains any floating goproc.Process nodes, they
get built by this function.
*/
func buildDefaultGolangProcess(outputDir string, node ir.IRNode) error {
	if proc, isProc := node.(*Process); isProc {
		procDir, err := ioutil.CreateNodeDir(outputDir, node.Name())
		if err != nil {
			return err
		}
		if err := proc.GenerateArtifacts(procDir); err != nil {
			return err
		}
	}
	return nil
}
