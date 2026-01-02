package workflow

import (
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/ir"
)

func FilterWorkflowNodes(nodes []ir.IRNode) []*workflowHandler {
	return ir.Filter[*workflowHandler](nodes)
}
