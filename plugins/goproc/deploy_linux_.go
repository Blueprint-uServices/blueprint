package goproc

import (
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/ir"
	"github.com/blueprint-uservices/blueprint/plugins/docker"
	"github.com/blueprint-uservices/blueprint/plugins/goproc/linuxgen"
	"github.com/blueprint-uservices/blueprint/plugins/linux"
)

// This file name ends with an underscore because Go has magic filenames that won't compile

/*
Goprocs can be deployed to linux, which simply follows the same process as the
BasicGoProc deployer, but also adds a run.sh script that pulls process arguments
from the local environment.

The LinuxGoProc deployer doesn't set up the linux environment with necessary
dependencies (e.g. installing Go); it is expected that the user will do this.
*/

type linuxDeployer interface {
	linux.Process
	linux.ProvidesProcessArtifacts
	linux.InstantiableProcess
}

// Implements linux.ProvidesProcessArtifacts
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
		procName := ir.CleanName(node.Name())
		buildCmds, err := linuxgen.GenerateDockerfileBuildCommands(procName)
		dockerWorkspace.AddDockerfileCommands(procName, buildCmds)
		return err
	}
	return nil
}

// Implements linux.InstantiableProcess
func (node *Process) AddProcessInstance(builder linux.ProcessWorkspace) error {
	if builder.Visited(node.InstanceName + ".instance") {
		return nil
	}

	procName := ir.CleanName(node.Name())

	var runfunc string
	var err error
	switch builder.(type) {
	case docker.ProcessWorkspace:
		runfunc, err = linuxgen.GenerateBinaryRunFunc(procName, node.Edges...)
	default:
		runfunc, err = linuxgen.GenerateRunFunc(procName, node.Edges...)
	}
	if err != nil {
		return err
	}

	return builder.DeclareRunCommand(node.InstanceName, runfunc, node.Edges...)
}

func (node *Process) ImplementsLinuxProcess() {}
