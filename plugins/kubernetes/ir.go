package kubernetes

import (
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/ir"
)

// KubernetesDeployment is an IRNode representing a Kubernetes deployment,
// which is a collection of container instances that will be deployed to a
// Kubernetes cluster.
type KubernetesDeployment struct {
	/* The implemented build targets for kubernetes.KubernetesDeployment nodes */
	kubernetesDeployer /* Can be deployed as Kubernetes manifests; implemented in deploy.go */

	DeploymentName string
	Namespace      string
	Replicas       int32
	ClusterConfig  *ClusterConfiguration
	Nodes          []ir.IRNode
	Edges          []ir.IRNode
}

// ClusterConfiguration holds runtime configuration for connecting to a Kubernetes cluster
type ClusterConfiguration struct {
	// Path to kubeconfig file (optional - can be provided at runtime)
	KubeconfigPath string
	// Kubernetes API server endpoint (optional - can be provided at runtime)
	APIServer string
	// Authentication token (optional - can be provided at runtime)
	Token string
	// Namespace to deploy to (can be overridden at runtime)
	Namespace string
}

// Implements IRNode
func (node *KubernetesDeployment) Name() string {
	return node.DeploymentName
}

// Implements IRNode
func (node *KubernetesDeployment) String() string {
	return ir.PrettyPrintNamespace(node.DeploymentName, "KubernetesDeployment", node.Edges, node.Nodes)
}

// SetNamespace sets the Kubernetes namespace for this deployment
func (node *KubernetesDeployment) SetNamespace(namespace string) {
	node.Namespace = namespace
}

// SetReplicas sets the number of replicas for this deployment
func (node *KubernetesDeployment) SetReplicas(replicas int32) {
	node.Replicas = replicas
}

// SetClusterConfig sets the cluster configuration for this deployment
func (node *KubernetesDeployment) SetClusterConfig(config *ClusterConfiguration) {
	node.ClusterConfig = config
}
