package goproc

import (
	"os"
	"path/filepath"

	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/blueprint"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/linux"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/linuxcontainer/linuxgen"
	"golang.org/x/mod/modfile"
)

/*
Goprocs can be deployed to linux, which simply follows the same process as the
BasicGoProc deployer, but also adds a run.sh script that pulls process arguments
from the local environment.

The LinuxGoProc deployer doesn't set up the linux environment with necessary
dependencies (e.g. installing Go); it is expected that the user will do this.
*/

type LinuxGoProc interface {
	linux.Process
	linux.ProvidesProcessArtifacts
	linux.InstantiableProcess
}

type runFuncTemplateArgs struct {
	Name           string
	GoWorkspaceDir string
	GoMainFile     string
	Args           []blueprint.IRNode
}

var runFuncTemplate = `
run_{{RunFuncName .Name}} {
	export CGO_ENABLED=1
	cd {{.GoWorkspaceDir}}
	go run {{.GoMainFile}}
	{{- range $i, $arg := .Args}} --{{$arg.Name}}=${{EnvVarName $arg.Name}}{{end}} &
	{{EnvVarName .Name}}=$!
	return $?
}`

/*
From process.ProvidesProcessArtifacts
*/
func (node *Process) AddProcessArtifacts(builder linux.ProcessWorkspace) error {
	if builder.Visited(node.Name() + ".artifacts") {
		return nil
	}

	// Create the workspace dir
	outputDir, err := builder.CreateProcessDir(node.ProcName)
	if err != nil {
		return err
	}

	// switch builder.(type) {
	// case *linuxgen.ProcessWorkspaceImpl: p
	// }
	return node.GenerateArtifacts(outputDir)
}

/*
From process.InstantiableProcess
*/
func (node *Process) AddProcessInstance(builder linux.ProcessWorkspace) error {
	if builder.Visited(node.InstanceName + ".instance") {
		return nil
	}

	mainFile, err := node.findMainFile(builder)
	if err != nil {
		return err
	}

	workspacePath := builder.Info().Path
	procDir := filepath.Join(workspacePath, node.ProcName)
	mainFilePath, err := filepath.Rel(procDir, mainFile)
	if err != nil {
		return err
	}

	templateArgs := runFuncTemplateArgs{
		Name:           node.InstanceName,
		GoWorkspaceDir: filepath.ToSlash(node.ProcName),
		GoMainFile:     filepath.ToSlash(mainFilePath),
		Args:           node.ArgNodes,
	}

	runfunc, err := linuxgen.ExecuteTemplate("rungoproc", runFuncTemplate, templateArgs)
	if err != nil {
		return err
	}

	return builder.DeclareRunCommand(node.InstanceName, runfunc, node.ArgNodes...)
}

func (node *Process) findMainFile(builder linux.ProcessWorkspace) (string, error) {
	goWorkspaceDir := filepath.Join(builder.Info().Path, node.ProcName)
	entries, err := os.ReadDir(goWorkspaceDir)
	if err != nil {
		return "", err
	}
	for _, e := range entries {
		if e.IsDir() {
			modDir := filepath.Join(goWorkspaceDir, e.Name())
			modFileName := filepath.Join(modDir, "go.mod")
			modFileData, err := os.ReadFile(modFileName)
			if err != nil {
				continue
			}
			f, err := modfile.Parse(modFileName, modFileData, nil)
			if err != nil {
				continue
			}
			if f.Module.Mod.Path == node.ModuleName {
				return filepath.Join(modDir, "main.go"), nil
			}
		}
	}
	return "", blueprint.Errorf("unable to find main.go file for golang process %v in %v", node.InstanceName, goWorkspaceDir)
}

func (node *Process) ImplementsLinuxProcess() {}
