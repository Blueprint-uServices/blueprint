package linuxgen

import (
	"strings"

	"github.com/blueprint-uservices/blueprint/blueprint/pkg/blueprint"
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/blueprint/stringutil"
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/ir"
)

/*
Used to generate bash run funcs that get gathered together into a single run.sh

Process nodes provide the bash command needed to execute the process; but that's
not enough.  We need to make sure dependencies are correctly instantiated when needed.

This code provides a wrapper function implementation around the commands provided
by the process nodes
*/

func GenerateRunFunc(name string, runfunc string, deps ...ir.IRNode) (string, error) {
	runfunc = getFuncBody(runfunc)
	if runfunc == "" {
		return "", blueprint.Errorf("invalid runfunc for process %v %v", name, runfunc)
	}

	templateArgs := runFuncTemplateArgs{
		Name:         name,
		Dependencies: deps,
		RunFuncBody:  stringutil.Reindent(runfunc, 8),
	}

	return ExecuteTemplate("runfunc", runFuncTemplate, templateArgs)
}

func getFuncBody(runcmd string) string {
	from := strings.Index(runcmd, "{") + 1
	to := strings.LastIndex(runcmd, "}")
	if from < to {
		return runcmd[from:to]
	} else {
		return ""
	}
}

type runFuncTemplateArgs struct {
	Name         string
	Dependencies []ir.IRNode
	RunFuncBody  string
}

var runFuncTemplate = `{{RunFuncName .Name}}() {
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
