// Command petshop compiles the PetShop Blueprint application.
package main

import (
	_ "github.com/blueprint-uservices/blueprint/examples/petshop/tests"
	"github.com/blueprint-uservices/blueprint/examples/petshop/wiring/specs"
	"github.com/blueprint-uservices/blueprint/plugins/cmdbuilder"
	"github.com/blueprint-uservices/blueprint/plugins/workflow/workflowspec"
)

func main() {
	workflowspec.AddModule("github.com/blueprint-uservices/blueprint/examples/petshop/tests")
	cmdbuilder.MakeAndExecute("PetShop", specs.Basic)
}
