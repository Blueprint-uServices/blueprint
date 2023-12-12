package dockerdeployment

import (
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/coreplugins/pointer"
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/ir"
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/wiring"
)

var prop_CHILDREN = "Children"

// Adds a child node to an existing container deployment
func AddContainerToDeployment(spec wiring.WiringSpec, deploymentName, containerName string) {
	spec.AddProperty(deploymentName, prop_CHILDREN, containerName)
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
		deploymentNamespace := wiring.CreateNamespace(spec, namespace, deployment)
		_, err := pointer.InstantiateFromProperty(spec, deploymentNamespace, prop_CHILDREN)
		return deployment, err
	})

	return deploymentName
}
