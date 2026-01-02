package counter

import (
	"log"

	"github.com/blueprint-uservices/blueprint/blueprint/pkg/coreplugins/analysis"
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/ir"
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/wiring"
)

type IRNodeCounterPass struct {
}

func NewIRNodeCounterPass() analysis.IRAnalysisPass {
	return &IRNodeCounterPass{}
}

// Implements analysis.IRAnalysisPass
func (p *IRNodeCounterPass) Analyze(spec wiring.WiringSpec, app *ir.ApplicationNode) (bool, error) {
	cntr := len(spec.Defs())
	log.Printf("[%v]: Wiring spec has %v nodes: %v\n", p.Name(), cntr, spec.Defs())
	return false, nil
}

// Implements analysis.IRAnalysisPass
func (p *IRNodeCounterPass) Name() string {
	return "IRNodeCounterPass"
}

// Implements analysis.IRAnalysisPass
func (p *IRNodeCounterPass) ImplementsAnalysisPass() {}
