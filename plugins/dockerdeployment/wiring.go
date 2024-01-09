package dockerdeployment

import (
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/coreplugins/namespaceutil"
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/ir"
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/wiring"
	"github.com/blueprint-uservices/blueprint/plugins/docker"
)

// Adds a child node to an existing container deployment
func AddContainerToDeployment(spec wiring.WiringSpec, deploymentName, containerName string) {
	namespaceutil.AddNodeTo[Deployment](spec, deploymentName, containerName)
}

// Adds a deployment that explicitly instantiates all of the containers provided.
// The deployment will also implicitly instantiate any of the dependencies of the containers
func NewDeployment(spec wiring.WiringSpec, deploymentName string, containers ...string) string {
	// If any children were provided in this call, add them to the process via a property
	for _, containerName := range containers {
		AddContainerToDeployment(spec, deploymentName, containerName)
	}

	spec.Define(deploymentName, &Deployment{}, func(namespace wiring.Namespace) (ir.IRNode, error) {
		deployment := &Deployment{DeploymentName: deploymentName}
		_, err := namespaceutil.InstantiateNamespace(namespace, &DeploymentNamespace{deployment})
		return deployment, err
	})

	return deploymentName
}

// A [wiring.NamespaceHandler] used to build container deployments
type DeploymentNamespace struct {
	*Deployment
}

// Implements [wiring.NamespaceHandler]
func (deployment *Deployment) Accepts(nodeType any) bool {
	_, isDockerContainerNode := nodeType.(docker.Container)
	return isDockerContainerNode
}

// Implements [wiring.NamespaceHandler]
func (deployment *Deployment) AddEdge(name string, edge ir.IRNode) error {
	deployment.Edges = append(deployment.Edges, edge)
	return nil
}

// Implements [wiring.NamespaceHandler]
func (deployment *Deployment) AddNode(name string, node ir.IRNode) error {
	deployment.Nodes = append(deployment.Nodes, node)
	return nil
}
