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
	AllNodes  map[string]blueprint.IRNode // All of the nodes used as dependencies
	Args      map[string]blueprint.IRNode // Nodes that will be passed as arguments.
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
		RunFuncs: make(map[string]string),
		AllNodes: make(map[string]blueprint.IRNode),
		Args:     make(map[string]blueprint.IRNode),
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

var runCommandTemplate = `{{RunFuncName .Name}}() {
	cd $WORKSPACE_DIR
	
	{{ range $i, $dep := .Dependencies -}}
	{{Get $dep.Name}}

	{{end -}}

	run_{{RunFuncName .Name}}() {
		{{.RunFuncBody}}
	}

	if run_{{RunFuncName .Name}}; then
		if [ -z "${ {{- EnvVarName .Name}}+x}" ]; then
			echo "${WORKSPACE_NAME} error starting {{.Name}}: function {{RunFuncName .Name}} did not set {{EnvVarName .Name}}"
			return 1
		else
			echo "${WORKSPACE_NAME} started {{.Name}}"
			return 0
		fi
	else
		exitcode=$?
		echo "${WORKSPACE_NAME} aborting {{.Name}} due to exitcode ${exitcode} from {{RunFuncName .Name}}"
		return $exitcode
	fi
}`

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

	for _, dep := range deps {
		builder.AllNodes[dep.Name()] = dep
	}

	templateArgs := runCommandArgs{
		Name:         name,
		Dependencies: deps,
		RunFuncBody:  blueprint.Reindent(runfunc, 8),
	}

	actualRunFunc, err := ExecuteTemplate("runfunc", runCommandTemplate, templateArgs)

	builder.RunFuncs[name] = actualRunFunc
	return err
}

/*
Indicate that the provided node is an argument that will be passed in as an environment variable
from the calling environment
*/
func (builder *ProcGraphBuilderImpl) AddArg(node blueprint.IRNode) error {
	builder.Args[node.Name()] = node
	return nil
}

var runfileTemplate = `#!/bin/bash
cd "$(dirname "$0")"

WORKSPACE_NAME="{{.Info.Name}}"
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
	echo "Running {{.Info.Name}}"

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

func (builder *ProcGraphBuilderImpl) Build() error {
	// Figure out which nodes are used as dependencies but aren't declared locally; add them as args
	for name, node := range builder.AllNodes {
		if _, isLocal := builder.RunFuncs[name]; !isLocal {
			if err := builder.AddArg(node); err != nil {
				return err
			}
		}
	}

	filePath := filepath.Join(builder.workspace.Info().Path, builder.info.FilePath)
	return ExecuteTemplateToFile("runfile", runfileTemplate, builder, filePath)
}

func (builder *ProcGraphBuilderImpl) ImplementsBuildContext() {}
