package linuxgen

import (
	"path/filepath"

	"github.com/blueprint-uservices/blueprint/blueprint/pkg/ir"
)

/*
Within a process workspace, a single run.sh file will be generated
at the root of the workspace.  This run.sh file will check that
all necessary environment variables have been set, and will then
instantiate all of the processes in the workspace.
*/

type RunScript struct {
	WorkspaceName string
	WorkspaceDir  string
	FileName      string
	FilePath      string
	RunFuncs      map[string]string    // Function bodies provided by processes
	AllNodes      map[string]ir.IRNode // All nodes seen by this run script
	Args          map[string]ir.IRNode // Arguments that will be set in calling the environment
}

/*
Creates a new run.sh that will check environment variables are set
and invokes the run scripts of the processes within the workspace
*/
func NewRunScript(workspaceName, workspaceDir, fileName string) *RunScript {
	return &RunScript{
		WorkspaceName: workspaceName,
		WorkspaceDir:  workspaceDir,
		FileName:      fileName,
		FilePath:      filepath.Join(workspaceDir, fileName),
		RunFuncs:      make(map[string]string),
		AllNodes:      make(map[string]ir.IRNode),
		Args:          make(map[string]ir.IRNode),
	}
}

/*
Indicate that the specified node is required within this namespace;
either it is built by its own runfunc, or it must be provided as
argument.

We use this so that the generated run.sh knows which environment variables
will be needed or used by the processes it runs.
*/
func (run *RunScript) Require(node ir.IRNode) {
	run.AllNodes[node.Name()] = node
}

func (run *RunScript) Add(procName, runfunc string, deps ...ir.IRNode) {
	// Save the runfunc
	run.RunFuncs[procName] = runfunc

	// Note down all nodes that must be instantiated or have env var set
	for _, node := range deps {
		run.AllNodes[node.Name()] = node
	}
}

func (run *RunScript) GenerateRunScript() error {
	/*
	 Before generating the run.sh we must figure out which arguments are required
	 by the processes and therefore must be set by the calling environment for
	 run.sh to succeed
	*/
	for name, node := range run.AllNodes {
		if _, hasRunFunc := run.RunFuncs[name]; !hasRunFunc {
			// This node doesn't have a run func, so it must be an arg
			run.Args[name] = node
		}
	}

	// Now we can generate the run.sh
	return ExecuteTemplateToFile("run.sh", runfileTemplate, run, run.FilePath)
}

var runfileTemplate = `#!/bin/bash

WORKSPACE_NAME="{{.WorkspaceName}}"
WORKSPACE_DIR=$(pwd)

usage() { 
	echo "Usage: $0 [-h]" 1>&2
	echo "  Required environment variables:"
	
	{{range $name, $arg := .Args -}}
	if [ -z "${ {{- EnvVarName .Name}}+x}" ]; then
		echo "    {{EnvVarName .Name}} (missing)"
	else
		echo "    {{EnvVarName .Name}}=${{EnvVarName .Name}}"
	fi
	{{end}}	
	exit 1; 
}

while getopts "h" flag; do
	case $flag in
		*)
		usage
		;;
	esac
done

{{range $name, $f := .RunFuncs}}
{{$f}}
{{end}}

run_all() {
	echo "Running {{.WorkspaceName}}"

	# Check that all necessary environment variables are set
	echo "Required environment variables:"
	missing_vars=0
	{{- range $name, $arg := .Args}}
	if [ -z "${ {{- EnvVarName .Name}}+x}" ]; then
		echo "  {{EnvVarName .Name}} (missing)"
		missing_vars=$((missing_vars+1))
	else
		echo "  {{EnvVarName .Name}}=${{EnvVarName .Name}}"
	fi
	{{end}}	

	if [ "$missing_vars" -gt 0 ]; then
		echo "Aborting due to missing environment variables"
		return 1
	fi

	{{range $name, $f := .RunFuncs -}}
	{{RunFuncName $name}}
	{{end}}
	wait
}

run_all
`
