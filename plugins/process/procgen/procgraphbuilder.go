package procgen

import (
	"path/filepath"
	"strings"

	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/blueprint"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/process"
)

type ProcGraphBuilderImpl struct {
	blueprint.VisitTrackerImpl
	workspace process.ProcWorkspaceBuilder
	info      process.ProcGraphInfo
	RunFuncs  map[string]string
}

func NewProcGraphBuilderImpl(workspace process.ProcWorkspaceBuilder, name string, fileName string) (*ProcGraphBuilderImpl, error) {
	dir, file := filepath.Split(filepath.Clean(fileName))

	builder := &ProcGraphBuilderImpl{
		workspace: workspace,
		info: process.ProcGraphInfo{
			Workspace: workspace.Info(),
			Name:      name,
			FileName:  file,
			FileDir:   filepath.ToSlash(filepath.Clean(dir)),
			FilePath:  filepath.ToSlash(filepath.Clean(fileName)),
		},
	}

	return builder, nil
}

func (builder *ProcGraphBuilderImpl) Info() process.ProcGraphInfo {
	return builder.info
}

type runCommandArgs struct {
	Name         string
	Dependencies []blueprint.IRNode
	RunFuncBody  string
}

var runCommandTemplate = `
{{RunFuncName .Name}}() {
	cd $WORKSPACE_DIR
	
	{{ range $i, $dep := .Dependencies}}
	{{Get $dep.Name}}
	{{end}}

	function run_{{RunFuncName .Name}}() {
		{{.RunFuncBody}}
	}

	if run_{{RunFuncName .Name}}; then
		if [ -z "${ {{- EnvVarName .Name}}+x}" ]; then
			echo "${PROCNAME} error starting {{.Name}}: function {{RunFuncName .Name}} did not set {{EnvVarName .Name}}"
			return 1
		else
			echo "${PROCNAME} started {{.Name}}"
			return 0
		fi
	else
		exitcode=$?
		echo "${PROCNAME} aborting {{.Name}} due to exitcode ${exitcode} from {{RunFuncName .Name}}"
		return $exitcode
	fi
}
`

func getFuncBody(runcmd string) string {
	from := strings.Index(runcmd, "{") + 1
	to := strings.LastIndex(runcmd, "}")
	if from < to {
		return runcmd[from:to]
	} else {
		return ""
	}
}

func (builder *ProcGraphBuilderImpl) DeclareRunCommand(name string, runfunc string, deps ...blueprint.IRNode) error {
	runfunc = getFuncBody(runfunc)
	if runfunc == "" {
		return blueprint.Errorf("invalid runfunc for process %v %v", name, runfunc)
	}

	templateArgs := runCommandArgs{
		Name:         name,
		Dependencies: deps,
		RunFuncBody:  runfunc,
	}

	actualRunFunc, err := ExecuteTemplate("runfunc", runCommandTemplate, templateArgs)
	builder.RunFuncs[name] = actualRunFunc
	return err
}

var runfileTemplate = `#!/bin/bash

PROCNAME="{{.Info.Name}}"
WORKSPACE_DIR=$(pwd)

{{range $i, $f := .RunFuncs}}
{{$f}}
{{end}}

`

func (builder *ProcGraphBuilderImpl) Build() error {
	return ExecuteTemplateToFile("runfile", runfileTemplate, builder, builder.info.FilePath)
}

func (builder *ProcGraphBuilderImpl) ImplementsBuildContext() {}
