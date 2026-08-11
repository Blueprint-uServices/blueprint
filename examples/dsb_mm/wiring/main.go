package main

import (
	"github.com/Blueprint-uservices/blueprint/examples/dsb_mm/wiring/specs"
	"github.com/blueprint-uservices/blueprint/plugins/cmdbuilder"
)

func main() {
	name := "Media"
	cmdbuilder.MakeAndExecute(
		name,
		specs.Default,
	)
}
