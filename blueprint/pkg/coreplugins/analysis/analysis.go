package analysis

import "github.com/blueprint-uservices/blueprint/blueprint/pkg/ir"

// IRAnalysisPass is a Blueprint compiler pass that operates on an ir.ApplicationNode.
// An analysis pass may optionally modify the ApplicationNode to add/remove/modify any of the nodes in the ApplicationNode
type IRAnalysisPass interface {
	Analyze(app *ir.ApplicationNode) (bool, error)
	Name() string
}
