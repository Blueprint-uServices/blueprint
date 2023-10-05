package irutil

import "gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/blueprint"

type BuildContext interface {
	VisitTracker
	Visit(nodes []blueprint.IRNode) error
}

type NullBuildContext struct{}

func (v *NullBuildContext) Visit(nodes []blueprint.IRNode) error {
	return nil
}

func (v *NullBuildContext) Visited(name string) bool {
	return false
}
