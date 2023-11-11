package linux

import (
	"strings"

	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/ir"
)

func EnvVar(name string) string {
	return strings.ToUpper(ir.CleanName(name))
}

func FuncName(name string) string {
	return strings.ToLower(ir.CleanName(name))
}
