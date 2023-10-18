package goproc

import (
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/blueprint"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/docker"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/goproc/goprocgen"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/linux"
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

	// Generate the regular artifacts for the process
	if err := node.GenerateArtifacts(outputDir); err != nil {
		return err
	}

	// If it's a docker container, we can also add Dockerfile build commands
	if dockerWorkspace, isDocker := builder.(docker.ProcessWorkspace); isDocker {
		procName := blueprint.CleanName(node.Name())
		buildCmds, err := goprocgen.GenerateDockerfileBuildCommands(procName)
		dockerWorkspace.AddDockerfileCommands(procName, buildCmds)
		return err
	}
	return nil
}

/*
From process.InstantiableProcess
*/
func (node *Process) AddProcessInstance(builder linux.ProcessWorkspace) error {
	if builder.Visited(node.InstanceName + ".instance") {
		return nil
	}

	procName := blueprint.CleanName(node.Name())

	var runfunc string
	var err error
	switch builder.(type) {
	case docker.ProcessWorkspace:
		runfunc, err = goprocgen.GenerateDockerRunFunc(procName, node.ArgNodes...)
	default:
		runfunc, err = goprocgen.GenerateRunFunc(procName, node.ArgNodes...)
	}
	if err != nil {
		return err
	}

	return builder.DeclareRunCommand(node.InstanceName, runfunc, node.ArgNodes...)
}

func (node *Process) ImplementsLinuxProcess() {}
