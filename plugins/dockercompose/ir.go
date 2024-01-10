package dockercompose

import (
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/ir"
)

// An IRNode representing a docker-compose deployment, which is simply a collection of
// container instances.
type Deployment struct {
	/* The implemented build targets for dockercompose.DockerCompose nodes */
	dockerComposeDeployer /* Can be deployed as a docker-compose file; implemented in deploydockercompose.go */

	DeploymentName string
	Nodes          []ir.IRNode
	Edges          []ir.IRNode
}

// Implements IRNode
func (node *Deployment) Name() string {
	return node.DeploymentName
}

// Implements IRNode
func (node *Deployment) String() string {
	return ir.PrettyPrintNamespace(node.DeploymentName, "DockerApp", node.Edges, node.Nodes)
}
