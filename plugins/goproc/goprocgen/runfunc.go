package goprocgen

import (
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/blueprint"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/linuxcontainer/linuxgen"
)

/*
Generates command-line function to run a goproc
*/
func GenerateRunFunc(procName string, args ...blueprint.IRNode) (string, error) {
	templateArgs := runFuncTemplateArgs{
		Name: procName,
		Args: args,
	}
	return linuxgen.ExecuteTemplate("goproc_runfunc", runFuncTemplate, templateArgs)
}

/*
Generates command-line function to run a goproc in a Docker container

When a goproc is in a Docker container, it is executed differently
from when it's not.
*/
func GenerateDockerRunFunc(procName string, args ...blueprint.IRNode) (string, error) {
	templateArgs := runFuncTemplateArgs{
		Name: procName,
		Args: args,
	}
	return linuxgen.ExecuteTemplate("goproc_dockerrunfunc", dockerRunFuncTemplate, templateArgs)
}

type runFuncTemplateArgs struct {
	Name string
	Args []blueprint.IRNode
}

var dockerRunFuncTemplate = `
run_{{RunFuncName .Name}} {
	cd {{.Name}}
    ./{{.Name}}
	{{- range $i, $arg := .Args}} --{{$arg.Name}}=${{EnvVarName $arg.Name}}{{end}} &
	{{EnvVarName .Name}}=$!
	return $?
}`

var runFuncTemplate = `
run_{{RunFuncName .Name}} {
	export CGO_ENABLED=1
	cd {{.Name}}
	go run {{.Name}}/main.go
	{{- range $i, $arg := .Args}} --{{$arg.Name}}=${{EnvVarName $arg.Name}}{{end}} &
	{{EnvVarName .Name}}=$!
	return $?
}`
