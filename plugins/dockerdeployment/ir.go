package dockerdeployment

import (
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/ir"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/docker"
)

/* A deployment is a collection of containers */
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

// Implements SimpleNamespaceHandler
func (deployment *Deployment) Accepts(nodeType any) bool {
	_, isDockerContainerNode := nodeType.(docker.Container)
	return isDockerContainerNode
}

// Implements SimpleNamespaceHandler
func (deployment *Deployment) AddEdge(name string, edge ir.IRNode) error {
	deployment.Edges = append(deployment.Edges, edge)
	return nil
}

// Implements SimpleNamespaceHandler
func (deployment *Deployment) AddNode(name string, node ir.IRNode) error {
	deployment.Nodes = append(deployment.Nodes, node)
	return nil
}
