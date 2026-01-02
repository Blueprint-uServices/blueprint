package analysis

import (
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/ir"
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/wiring"
)

// IRAnalysisPass is a Blueprint compiler pass that operates on an ir.ApplicationNode.
// An analysis pass may optionally modify the ApplicationNode to add/remove/modify any of the nodes in the ApplicationNode
type IRAnalysisPass interface {
	Analyze(spec wiring.WiringSpec, app *ir.ApplicationNode) (bool, error)
	Name() string
	ImplementsAnalysisPass()
}
