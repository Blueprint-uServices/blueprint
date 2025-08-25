package kubernetes

import (
	"testing"

	"github.com/blueprint-uservices/blueprint/blueprint/pkg/blueprint/logging"
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/ir"
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/wiring"
	"github.com/blueprint-uservices/blueprint/plugins/docker"
	"github.com/blueprint-uservices/blueprint/plugins/golang"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Helper function to create a new wiring spec for tests
func newTestWiringSpec(name string) wiring.WiringSpec {
	// Disable compiler logging for tests unless explicitly enabled
	logging.DisableCompilerLogging()
	spec := wiring.NewWiringSpec(name)
	return spec
}

// Helper to build IR and return result
func buildIR(t *testing.T, spec wiring.WiringSpec, toInstantiate ...string) (*ir.ApplicationNode, error) {
	return spec.BuildIR(toInstantiate...)
}

func TestNewDeployment(t *testing.T) {
	spec := newTestWiringSpec("TestNewDeployment")

	deploymentName := "test-app"
	deployment := NewDeployment(spec, deploymentName)

	// Build IR
	app, err := buildIR(t, spec, deploymentName)
	require.NoError(t, err)
	require.NotNil(t, app)

	// Find the deployment node
	deploymentNode := app.GetChildren()[deploymentName]
	require.NotNil(t, deploymentNode)

	k8sDeployment, ok := deploymentNode.(*KubernetesDeployment)
	require.True(t, ok)
	assert.Equal(t, deploymentName, k8sDeployment.DeploymentName)
	assert.Equal(t, deploymentName, deployment)
}

func TestAddContainerToDeployment(t *testing.T) {
	spec := newTestWiringSpec("TestAddContainerToDeployment")

	// Create a golang service
	service := golang.Service(spec, "myservice")

	// Create deployment and add service
	deployment := NewDeployment(spec, "test-deployment")
	AddContainerToDeployment(spec, deployment, service)

	// Build IR
	app, err := buildIR(t, spec, deployment)
	require.NoError(t, err)

	// Find deployment node
	deploymentNode := app.GetChildren()[deployment]
	require.NotNil(t, deploymentNode)

	k8sDeployment, ok := deploymentNode.(*KubernetesDeployment)
	require.True(t, ok)
	assert.Len(t, k8sDeployment.Containers, 1)
}

func TestAddMultipleContainersToDeployment(t *testing.T) {
	spec := newTestWiringSpec("TestAddMultipleContainers")

	// Create multiple services
	service1 := golang.Service(spec, "service1")
	service2 := golang.Service(spec, "service2")
	service3 := golang.Service(spec, "service3")

	// Create deployment and add all services
	deployment := NewDeployment(spec, "multi-container-app")
	AddContainerToDeployment(spec, deployment, service1)
	AddContainerToDeployment(spec, deployment, service2)
	AddContainerToDeployment(spec, deployment, service3)

	// Build IR
	app, err := buildIR(t, spec, deployment)
	require.NoError(t, err)

	// Find deployment node
	deploymentNode := app.GetChildren()[deployment]
	require.NotNil(t, deploymentNode)

	k8sDeployment, ok := deploymentNode.(*KubernetesDeployment)
	require.True(t, ok)
	assert.Len(t, k8sDeployment.Containers, 3)
}

func TestSetNamespace(t *testing.T) {
	spec := newTestWiringSpec("TestSetNamespace")

	deployment := NewDeployment(spec, "test-app")
	SetNamespace(spec, deployment, "production")

	// Build IR
	app, err := buildIR(t, spec, deployment)
	require.NoError(t, err)

	// Find deployment node
	deploymentNode := app.GetChildren()[deployment]
	require.NotNil(t, deploymentNode)

	k8sDeployment, ok := deploymentNode.(*KubernetesDeployment)
	require.True(t, ok)
	assert.Equal(t, "production", k8sDeployment.Namespace)
}

func TestSetReplicas(t *testing.T) {
	spec := newTestWiringSpec("TestSetReplicas")

	deployment := NewDeployment(spec, "test-app")
	SetReplicas(spec, deployment, 5)

	// Build IR
	app, err := buildIR(t, spec, deployment)
	require.NoError(t, err)

	// Find deployment node
	deploymentNode := app.GetChildren()[deployment]
	require.NotNil(t, deploymentNode)

	k8sDeployment, ok := deploymentNode.(*KubernetesDeployment)
	require.True(t, ok)
	assert.Equal(t, 5, k8sDeployment.Replicas)
}

func TestConfigureCluster(t *testing.T) {
	spec := newTestWiringSpec("TestConfigureCluster")

	deployment := NewDeployment(spec, "test-app")
	ConfigureCluster(spec, deployment, "https://k8s.example.com", "/path/to/kubeconfig", "auth-token-123")

	// Build IR
	app, err := buildIR(t, spec, deployment)
	require.NoError(t, err)

	// Find deployment node
	deploymentNode := app.GetChildren()[deployment]
	require.NotNil(t, deploymentNode)

	k8sDeployment, ok := deploymentNode.(*KubernetesDeployment)
	require.True(t, ok)
	assert.Equal(t, "https://k8s.example.com", k8sDeployment.ClusterConfig.Endpoint)
	assert.Equal(t, "/path/to/kubeconfig", k8sDeployment.ClusterConfig.Kubeconfig)
	assert.Equal(t, "auth-token-123", k8sDeployment.ClusterConfig.AuthToken)
}

func TestCompleteDeploymentConfiguration(t *testing.T) {
	spec := newTestWiringSpec("TestCompleteConfiguration")

	// Create services
	frontend := golang.Service(spec, "frontend")
	backend := golang.Service(spec, "backend")

	// Create and configure deployment
	deployment := NewDeployment(spec, "full-app")
	AddContainerToDeployment(spec, deployment, frontend)
	AddContainerToDeployment(spec, deployment, backend)
	SetNamespace(spec, deployment, "staging")
	SetReplicas(spec, deployment, 3)
	ConfigureCluster(spec, deployment, "https://staging.k8s.local", "", "staging-token")

	// Build IR
	app, err := buildIR(t, spec, deployment)
	require.NoError(t, err)

	// Verify complete configuration
	deploymentNode := app.GetChildren()[deployment]
	require.NotNil(t, deploymentNode)

	k8sDeployment, ok := deploymentNode.(*KubernetesDeployment)
	require.True(t, ok)
	assert.Equal(t, "full-app", k8sDeployment.DeploymentName)
	assert.Equal(t, "staging", k8sDeployment.Namespace)
	assert.Equal(t, 3, k8sDeployment.Replicas)
	assert.Len(t, k8sDeployment.Containers, 2)
	assert.Equal(t, "https://staging.k8s.local", k8sDeployment.ClusterConfig.Endpoint)
	assert.Equal(t, "staging-token", k8sDeployment.ClusterConfig.AuthToken)
}

func TestNamespaceHandler(t *testing.T) {
	spec := newTestWiringSpec("TestNamespaceHandler")

	// Create a Kubernetes namespace using Define
	k8sNs := spec.Define("kubernetes", "k8s-app", namespaceHandler)

	// Create services
	service1 := golang.Service(spec, "service1")
	service2 := golang.Service(spec, "service2")

	// Place services in the Kubernetes namespace
	k8sNs.Place(service1)
	k8sNs.Place(service2)

	// Configure the deployment
	deployment := k8sNs.Instantiate()
	SetNamespace(spec, deployment, "k8s-namespace")
	SetReplicas(spec, deployment, 2)

	// Build IR
	app, err := buildIR(t, spec, deployment)
	require.NoError(t, err)

	// Verify namespace created a deployment with the services
	deploymentNode := app.GetChildren()[deployment]
	require.NotNil(t, deploymentNode)

	k8sDeployment, ok := deploymentNode.(*KubernetesDeployment)
	require.True(t, ok)
	assert.Equal(t, "k8s-app", k8sDeployment.DeploymentName)
	assert.Equal(t, "k8s-namespace", k8sDeployment.Namespace)
	assert.Equal(t, 2, k8sDeployment.Replicas)
	assert.Len(t, k8sDeployment.Containers, 2)
}

func TestAddContainerWithDockerContainer(t *testing.T) {
	spec := newTestWiringSpec("TestDockerContainer")

	// Create a docker container
	container := docker.Container(spec, "redis")

	// Create deployment and add container
	deployment := NewDeployment(spec, "test-deployment")
	AddContainerToDeployment(spec, deployment, container)

	// Build IR
	app, err := buildIR(t, spec, deployment)
	require.NoError(t, err)

	// Verify container was added
	deploymentNode := app.GetChildren()[deployment]
	require.NotNil(t, deploymentNode)

	k8sDeployment, ok := deploymentNode.(*KubernetesDeployment)
	require.True(t, ok)
	assert.Len(t, k8sDeployment.Containers, 1)
}

func TestEmptyDeployment(t *testing.T) {
	spec := newTestWiringSpec("TestEmptyDeployment")

	// Create deployment without adding any containers
	deployment := NewDeployment(spec, "empty-deployment")

	// Build IR
	app, err := buildIR(t, spec, deployment)
	require.NoError(t, err)

	// Verify empty deployment
	deploymentNode := app.GetChildren()[deployment]
	require.NotNil(t, deploymentNode)

	k8sDeployment, ok := deploymentNode.(*KubernetesDeployment)
	require.True(t, ok)
	assert.Equal(t, "empty-deployment", k8sDeployment.DeploymentName)
	assert.Empty(t, k8sDeployment.Containers)
	assert.Equal(t, "", k8sDeployment.Namespace) // Default empty
	assert.Equal(t, 0, k8sDeployment.Replicas)   // Default 0
}

func TestDefaultValues(t *testing.T) {
	spec := newTestWiringSpec("TestDefaultValues")

	// Create deployment with minimal configuration
	deployment := NewDeployment(spec, "minimal-app")
	service := golang.Service(spec, "myservice")
	AddContainerToDeployment(spec, deployment, service)

	// Build IR
	app, err := buildIR(t, spec, deployment)
	require.NoError(t, err)

	// Verify defaults
	deploymentNode := app.GetChildren()[deployment]
	require.NotNil(t, deploymentNode)

	k8sDeployment, ok := deploymentNode.(*KubernetesDeployment)
	require.True(t, ok)
	assert.Equal(t, "", k8sDeployment.Namespace)                // Empty namespace (will default to "default" in manifest)
	assert.Equal(t, 0, k8sDeployment.Replicas)                  // 0 replicas (will default to 1 in manifest)
	assert.Equal(t, "", k8sDeployment.ClusterConfig.Endpoint)   // Empty endpoint
	assert.Equal(t, "", k8sDeployment.ClusterConfig.Kubeconfig) // Empty kubeconfig
	assert.Equal(t, "", k8sDeployment.ClusterConfig.AuthToken)  // Empty token
}

func TestMultipleDeployments(t *testing.T) {
	spec := newTestWiringSpec("TestMultipleDeployments")

	// Create multiple deployments
	deployment1 := NewDeployment(spec, "app1")
	deployment2 := NewDeployment(spec, "app2")

	service1 := golang.Service(spec, "service1")
	service2 := golang.Service(spec, "service2")

	AddContainerToDeployment(spec, deployment1, service1)
	AddContainerToDeployment(spec, deployment2, service2)

	SetNamespace(spec, deployment1, "namespace1")
	SetNamespace(spec, deployment2, "namespace2")

	// Build IR for both deployments
	app, err := buildIR(t, spec, deployment1, deployment2)
	require.NoError(t, err)

	// Verify both deployments exist
	deployment1Node := app.GetChildren()[deployment1]
	deployment2Node := app.GetChildren()[deployment2]
	require.NotNil(t, deployment1Node)
	require.NotNil(t, deployment2Node)

	k8sDeployment1, ok1 := deployment1Node.(*KubernetesDeployment)
	k8sDeployment2, ok2 := deployment2Node.(*KubernetesDeployment)
	require.True(t, ok1)
	require.True(t, ok2)

	assert.Equal(t, "app1", k8sDeployment1.DeploymentName)
	assert.Equal(t, "namespace1", k8sDeployment1.Namespace)
	assert.Len(t, k8sDeployment1.Containers, 1)

	assert.Equal(t, "app2", k8sDeployment2.DeploymentName)
	assert.Equal(t, "namespace2", k8sDeployment2.Namespace)
	assert.Len(t, k8sDeployment2.Containers, 1)
}
