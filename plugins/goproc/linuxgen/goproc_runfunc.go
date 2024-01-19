// Package linuxgen implements code generation for the goproc plugin.
package linuxgen

import (
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/ir"
	"github.com/blueprint-uservices/blueprint/plugins/linuxcontainer/linuxgen"
)

/*
Generates command-line function to run a goproc
*/
func GenerateRunFunc(procName string, args ...ir.IRNode) (string, error) {
	templateArgs := runFuncTemplateArgs{
		Name: procName,
		Args: args,
	}
	return linuxgen.ExecuteTemplate("goproc_runfunc", runFuncTemplate, templateArgs)
}

/*
Generates command-line function to run a goproc that has been built to a binary
using `go build`
*/
func GenerateBinaryRunFunc(procName string, args ...ir.IRNode) (string, error) {
	templateArgs := runFuncTemplateArgs{
		Name: procName,
		Args: args,
	}
	return linuxgen.ExecuteTemplate("goproc_binaryrunfunc", binaryRunFuncTemplate, templateArgs)
}

type runFuncTemplateArgs struct {
	Name string
	Args []ir.IRNode
}

var binaryRunFuncTemplate = `
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
	cd {{.Name}}/{{.Name}}
	go run .
	{{- range $i, $arg := .Args}} --{{$arg.Name}}=${{EnvVarName $arg.Name}}{{end}} &
	{{EnvVarName .Name}}=$!
	return $?
}`
