package linux

import (
	"strings"

	"github.com/blueprint-uservices/blueprint/blueprint/pkg/ir"
)

func EnvVar(name string) string {
	return strings.ToUpper(ir.CleanName(name))
}

func FuncName(name string) string {
	return strings.ToLower(ir.CleanName(name))
}
