package process

import (
	"strings"

	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/blueprint"
)

func EnvVar(name string) string {
	return strings.ToUpper(blueprint.CleanName(name))
}
