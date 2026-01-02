package property

import (
	"log"

	"github.com/blueprint-uservices/blueprint/blueprint/pkg/coreplugins/analysis"
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/ir"
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/wiring"
	"github.com/blueprint-uservices/blueprint/plugins/workflow"
)

type PropertyPrintPass struct {
}

func NewPropertyPrintPass() analysis.IRAnalysisPass {
	return &PropertyPrintPass{}
}

// Implements analysis.IRAnalysisPass
func (p *PropertyPrintPass) Analyze(spec wiring.WiringSpec, app *ir.ApplicationNode) (bool, error) {
	all_defs := spec.Defs()
	for _, d := range all_defs {
		log.Printf("[%v] Properties for node %v: ", p.Name(), d)
		def := spec.GetDef(d)
		for k, v := range def.Properties {
			log.Printf("\t%v : %v\n", k, v)
		}
	}
	all_nodes := app.GetAllIRNodes()
	handlers := workflow.FilterWorkflowNodes(all_nodes)
	log.Println("Application has ", len(all_nodes), " nodes")
	log.Println("Application has ", len(handlers), " handlers")
	for _, handler := range handlers {
		log.Printf("Handler %v has %d methods\n", handler.Name(), len(handler.ServiceInfo.Struct.Methods))
	}
	return false, nil
}

// Implements analysis.IRAnalysisPass
func (p *PropertyPrintPass) Name() string {
	return "PropertyPrintPass"
}

// Implements analysis.IRAnalysisPass
func (p *PropertyPrintPass) ImplementsAnalysisPass() {}
