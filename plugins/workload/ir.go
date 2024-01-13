package workload

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/blueprint-uservices/blueprint/blueprint/pkg/blueprint/ioutil"
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/ir"
	"github.com/blueprint-uservices/blueprint/plugins/goproc"
)

// Wraps a goproc.Process in order to control its artifact generation
type workloadGenerator struct {
	ir.ArtifactGenerator

	WorkloadName string // The name of the service impl for the workload generator
	ProcName     string // The name of the proc of the workload generator
	ProcNode     *goproc.Process
}

func newWorkloadGenerator(name string, procName string) *workloadGenerator {
	wlgen := &workloadGenerator{
		WorkloadName: name,
		ProcName:     procName,
	}
	return wlgen
}

// Implements [ir.IRNode]
func (w *workloadGenerator) Name() string {
	return w.WorkloadName
}

// Implements [ir.IRNode]
func (w *workloadGenerator) String() string {
	return ir.PrettyPrintNamespace(w.WorkloadName, "WorkloadGenerator", w.ProcNode.Edges, w.ProcNode.Nodes)
}

// Implements [ir.ArtifactGenerator]
func (w *workloadGenerator) GenerateArtifacts(workspaceDir string) error {
	// Create a subdir for the actual process artifacts
	procDir, err := ioutil.CreateNodeDir(workspaceDir, w.ProcName)
	if err != nil {
		return err
	}

	// Generate process artifacts into the subdir
	err = w.ProcNode.GenerateArtifacts(procDir)
	if err != nil {
		return err
	}

	// Build the process
	mainPath := filepath.Join(procDir, w.ProcNode.ProcName)
	cmd := exec.Command("go", "build", "-o", "../../..", "-C", mainPath)
	var out strings.Builder
	cmd.Stdout = &out
	cmd.Stderr = &out
	fmt.Sprintf("go build -o ../../.. -C %v\n", mainPath)
	if err := cmd.Run(); err != nil {
		return err
	}

	// Now that the executable is built, remove the proc source dir
	return os.RemoveAll(workspaceDir)
}
