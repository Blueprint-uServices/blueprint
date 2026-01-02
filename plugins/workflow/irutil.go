package workflow

import (
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/ir"
)

// FilterWorkflowNodes filters a provided list of nodes to return only workflow handlers
func FilterWorkflowNodes(nodes []ir.IRNode) []*workflowHandler {
	return ir.Filter[*workflowHandler](nodes)
}
