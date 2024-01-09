package linux

import (
	"strings"

	"github.com/Blueprint-uServices/blueprint/blueprint/pkg/ir"
)

func EnvVar(name string) string {
	return strings.ToUpper(ir.CleanName(name))
}

func FuncName(name string) string {
	return strings.ToLower(ir.CleanName(name))
}
