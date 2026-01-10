// Package kubernetes is a plugin for deploying container instances to a Kubernetes cluster.
//
// # Wiring Spec Usage
//
// To use the kubernetes plugin in your wiring spec, you can declare a deployment, giving it a name and
// specifying which container instances to include:
//
//	kubernetes.NewDeployment(spec, "my_deployment", "container1", "container2")
//
// You can add containers to existing deployments:
//
//	kubernetes.AddContainerToDeployment(spec, "my_deployment", "container3")
//
// You can configure the Kubernetes namespace:
//
//	kubernetes.SetNamespace(spec, "my_deployment", "production")
//
// You can set the number of replicas:
//
//	kubernetes.SetReplicas(spec, "my_deployment", 3)
//
// You can provide cluster configuration:
//
//	config := &kubernetes.ClusterConfiguration{
//		KubeconfigPath: "/path/to/kubeconfig",
//		Namespace: "my-namespace",
//	}
//	kubernetes.ConfigureCluster(spec, "my_deployment", config)
//
// # Artifacts Generated
//
// During compilation, the plugin generates Kubernetes YAML manifests including:
//   - Deployment resources for container instances
//   - Service resources for inter-container networking
//   - ConfigMaps for environment variables
//   - An apply.sh script to deploy all resources
//
// # Running Artifacts
//
// The generated artifacts can be deployed using kubectl:
//
//	kubectl apply -f deployment.yaml
//	kubectl apply -f services.yaml
//	kubectl apply -f configmap.yaml
//
// Or use the generated script:
//
//	./apply.sh
//
// You will need to provide cluster credentials either via:
//   - KUBECONFIG environment variable
//   - ~/.kube/config file
//   - Explicit configuration in the wiring spec
package kubernetes

import (
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/coreplugins/namespaceutil"
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/ir"
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/wiring"
	"github.com/blueprint-uservices/blueprint/plugins/docker"
)

// AddContainerToDeployment can be used by wiring specs to add a container instance to an existing
// Kubernetes deployment.
func AddContainerToDeployment(spec wiring.WiringSpec, deploymentName, containerName string) {
	namespaceutil.AddNodeTo[KubernetesDeployment](spec, deploymentName, containerName)
}

// NewDeployment can be used by wiring specs to create a Kubernetes deployment that instantiates
// a number of containers.
//
// Further container instances can be added to the deployment by calling [AddContainerToDeployment].
//
// During compilation, generates Kubernetes YAML manifests that deploy the containers.
//
// Returns deploymentName.
func NewDeployment(spec wiring.WiringSpec, deploymentName string, containers ...string) string {
	// If any containers were provided in this call, add them to the deployment
	for _, containerName := range containers {
		AddContainerToDeployment(spec, deploymentName, containerName)
	}

	spec.Define(deploymentName, &KubernetesDeployment{}, func(namespace wiring.Namespace) (ir.IRNode, error) {
		deployment := &KubernetesDeployment{
			DeploymentName: deploymentName,
			Namespace:      "default",
			Replicas:       1,
		}
		_, err := namespaceutil.InstantiateNamespace(namespace, &kubernetesNamespace{deployment})
		return deployment, err
	})

	return deploymentName
}

// SetNamespace configures the Kubernetes namespace for a deployment.
// If not set, defaults to "default".
func SetNamespace(spec wiring.WiringSpec, deploymentName string, namespace string) {
	spec.Alter(deploymentName, func(ir ir.IRNode) error {
		if deployment, ok := ir.(*KubernetesDeployment); ok {
			deployment.SetNamespace(namespace)
		}
		return nil
	})
}

// SetReplicas configures the number of replicas for a deployment.
// If not set, defaults to 1.
func SetReplicas(spec wiring.WiringSpec, deploymentName string, replicas int32) {
	spec.Alter(deploymentName, func(ir ir.IRNode) error {
		if deployment, ok := ir.(*KubernetesDeployment); ok {
			deployment.SetReplicas(replicas)
		}
		return nil
	})
}

// ConfigureCluster provides cluster configuration for a deployment.
// This configuration can include kubeconfig path, API server endpoint, authentication token, etc.
// These values can also be provided at runtime through environment variables or command-line flags.
func ConfigureCluster(spec wiring.WiringSpec, deploymentName string, config *ClusterConfiguration) {
	spec.Alter(deploymentName, func(ir ir.IRNode) error {
		if deployment, ok := ir.(*KubernetesDeployment); ok {
			deployment.SetClusterConfig(config)
		}
		return nil
	})
}

// A [wiring.NamespaceHandler] used to build Kubernetes deployments
type kubernetesNamespace struct {
	*KubernetesDeployment
}

// Implements [wiring.NamespaceHandler]
func (deployment *KubernetesDeployment) Accepts(nodeType any) bool {
	_, isDockerContainerNode := nodeType.(docker.Container)
	return isDockerContainerNode
}

// Implements [wiring.NamespaceHandler]
func (deployment *KubernetesDeployment) AddEdge(name string, edge ir.IRNode) error {
	deployment.Edges = append(deployment.Edges, edge)
	return nil
}

// Implements [wiring.NamespaceHandler]
func (deployment *KubernetesDeployment) AddNode(name string, node ir.IRNode) error {
	deployment.Nodes = append(deployment.Nodes, node)
	return nil
}
