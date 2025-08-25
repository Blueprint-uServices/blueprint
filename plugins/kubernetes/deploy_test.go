package kubernetes

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"testing"

	"github.com/blueprint-uservices/blueprint/blueprint/pkg/blueprint"
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/coreplugins/address"
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/coreplugins/pointer"
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/ir"
	"github.com/blueprint-uservices/blueprint/plugins/docker"
	"github.com/blueprint-uservices/blueprint/plugins/golang"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

// mockBuildContext implements ir.BuildContext for testing
type mockBuildContext struct {
	outputDir string
	files     map[string]string
}

func newMockBuildContext(outputDir string) *mockBuildContext {
	return &mockBuildContext{
		outputDir: outputDir,
		files:     make(map[string]string),
	}
}

func (m *mockBuildContext) OutputDir() string {
	return m.outputDir
}

func (m *mockBuildContext) Info(msg string, args ...any) {
	// No-op for testing
}

func (m *mockBuildContext) Warn(msg string, args ...any) {
	// No-op for testing
}

func (m *mockBuildContext) Error(msg string, args ...any) error {
	return fmt.Errorf(msg, args...)
}

func (m *mockBuildContext) VisitChildren(node ir.IRNode) error {
	// No-op for testing
	return nil
}

func (m *mockBuildContext) DependsOn(target ir.IRNode, dependencies ...ir.IRNode) {
	// No-op for testing
}

func (m *mockBuildContext) ImplementsGolangNode(node ir.IRNode) error {
	// No-op for testing
	return nil
}

func (m *mockBuildContext) ImplementsGolangService(node ir.IRNode) (*golang.Service, error) {
	if service, ok := node.(*golang.Service); ok {
		return service, nil
	}
	return nil, fmt.Errorf("node does not implement golang.Service")
}

func (m *mockBuildContext) ImplementsGRPCServer(node ir.IRNode) (ir.IRNode, error) {
	// No-op for testing
	return nil, fmt.Errorf("not implemented")
}

// WriteFile captures file content for testing
func (m *mockBuildContext) WriteFile(filename, content string) error {
	m.files[filename] = content
	return nil
}

// ReadFile simulates reading a file
func (m *mockBuildContext) ReadFile(filename string) ([]byte, error) {
	if content, ok := m.files[filename]; ok {
		return []byte(content), nil
	}
	return nil, fmt.Errorf("file not found: %s", filename)
}

// GetWrittenFile retrieves a file that was written during generation
func (m *mockBuildContext) GetWrittenFile(filename string) (string, bool) {
	content, ok := m.files[filename]
	return content, ok
}

// GetAllFiles returns all written files
func (m *mockBuildContext) GetAllFiles() map[string]string {
	return m.files
}

// Additional methods to satisfy ir.BuildContext interface
func (m *mockBuildContext) Visited(node ir.IRNode) bool {
	return false
}

func (m *mockBuildContext) HasNode(name string) bool {
	return false
}

func (m *mockBuildContext) GetNode(name string) (ir.IRNode, error) {
	return nil, fmt.Errorf("not implemented")
}

func (m *mockBuildContext) VisitNode(name string, propagate bool, fn ir.VisitFunc) error {
	return nil
}

// Helper function to create a test deployment with containers
func createTestDeployment() *KubernetesDeployment {
	// Create a test service
	service := &golang.Service{
		Node: blueprint.Node{
			NodeName: "test-service",
		},
	}

	// Create a test container
	container := &docker.Container{
		Node: blueprint.Node{
			NodeName: "test-container",
		},
		ImageName: "test-image:latest",
		Ports: map[string]*address.BindConfig{
			"http": {
				Port: 8080,
			},
		},
		EnvironmentVariables: map[string]string{
			"ENV_VAR_1": "value1",
			"ENV_VAR_2": "value2",
		},
	}

	deployment := &KubernetesDeployment{
		Node: blueprint.Node{
			NodeName: "test-deployment",
		},
		DeploymentName: "test-deployment",
		Namespace:      "test-namespace",
		Replicas:       3,
		Containers:     []any{container},
		ClusterConfig: ClusterConfiguration{
			Endpoint:   "https://k8s.example.com",
			Kubeconfig: "/path/to/kubeconfig",
		},
	}

	return deployment
}

func TestGenerateArtifacts(t *testing.T) {
	ctx := context.Background()
	deployment := createTestDeployment()
	mockCtx := newMockBuildContext("/tmp/test")

	err := deployment.GenerateArtifacts(ctx, mockCtx)
	require.NoError(t, err)

	// Check that deployment manifest was generated
	deploymentYAML, ok := mockCtx.GetWrittenFile(filepath.Join("/tmp/test", "kubernetes", "manifests", "test-deployment-deployment.yaml"))
	require.True(t, ok, "Deployment manifest should be generated")
	assert.Contains(t, deploymentYAML, "kind: Deployment")
	assert.Contains(t, deploymentYAML, "name: test-deployment")
	assert.Contains(t, deploymentYAML, "namespace: test-namespace")
	assert.Contains(t, deploymentYAML, "replicas: 3")

	// Check that service manifest was generated
	serviceYAML, ok := mockCtx.GetWrittenFile(filepath.Join("/tmp/test", "kubernetes", "manifests", "test-container-service.yaml"))
	require.True(t, ok, "Service manifest should be generated")
	assert.Contains(t, serviceYAML, "kind: Service")
	assert.Contains(t, serviceYAML, "name: test-container")
	assert.Contains(t, serviceYAML, "port: 8080")

	// Check that ConfigMap was generated
	configMapYAML, ok := mockCtx.GetWrittenFile(filepath.Join("/tmp/test", "kubernetes", "manifests", "test-deployment-configmap.yaml"))
	require.True(t, ok, "ConfigMap should be generated")
	assert.Contains(t, configMapYAML, "kind: ConfigMap")
	assert.Contains(t, configMapYAML, "ENV_VAR_1: value1")
	assert.Contains(t, configMapYAML, "ENV_VAR_2: value2")

	// Check that deployment scripts were generated
	_, ok = mockCtx.GetWrittenFile(filepath.Join("/tmp/test", "kubernetes", "deploy.sh"))
	require.True(t, ok, "Linux deployment script should be generated")

	_, ok = mockCtx.GetWrittenFile(filepath.Join("/tmp/test", "kubernetes", "deploy.bat"))
	require.True(t, ok, "Windows deployment script should be generated")

	// Check that README was generated
	readme, ok := mockCtx.GetWrittenFile(filepath.Join("/tmp/test", "kubernetes", "README.md"))
	require.True(t, ok, "README should be generated")
	assert.Contains(t, readme, "Kubernetes Deployment")
	assert.Contains(t, readme, "test-deployment")
}

func TestKubernetesWorkspace(t *testing.T) {
	deployment := createTestDeployment()
	workspace := &kubernetesWorkspace{
		deployment: deployment,
	}

	// Test AddEnvironmentVariable
	err := workspace.AddEnvironmentVariable("TEST_VAR", "test_value")
	require.NoError(t, err)

	container := deployment.Containers[0].(*docker.Container)
	assert.Equal(t, "test_value", container.EnvironmentVariables["TEST_VAR"])

	// Test AddBindAddr
	bindCfg := &address.BindConfig{
		Port: 9090,
	}
	err = workspace.AddBindAddr("grpc", bindCfg)
	require.NoError(t, err)
	assert.Equal(t, bindCfg, container.Ports["grpc"])

	// Test AddContainerInstance
	newContainer := &docker.Container{
		Node: blueprint.Node{
			NodeName: "new-container",
		},
		ImageName: "new-image:latest",
	}
	err = workspace.AddContainerInstance(newContainer)
	require.NoError(t, err)
	assert.Len(t, deployment.Containers, 2)
	assert.Equal(t, newContainer, deployment.Containers[1])
}

func TestProcessArgNodes(t *testing.T) {
	// Create a deployment with multiple containers
	container1 := &docker.Container{
		Node: blueprint.Node{
			NodeName: "service1",
		},
		ImageName: "image1:latest",
		Ports: map[string]*address.BindConfig{
			"http": {Port: 8080},
		},
		EnvironmentVariables: make(map[string]string),
	}

	container2 := &docker.Container{
		Node: blueprint.Node{
			NodeName: "service2",
		},
		ImageName: "image2:latest",
		Ports: map[string]*address.BindConfig{
			"grpc": {Port: 9090},
		},
		EnvironmentVariables: make(map[string]string),
	}

	deployment := &KubernetesDeployment{
		Node: blueprint.Node{
			NodeName: "test-deployment",
		},
		DeploymentName: "test-deployment",
		Namespace:      "default",
		Containers:     []any{container1, container2},
	}

	workspace := &kubernetesWorkspace{
		deployment: deployment,
	}

	// Create pointers to services
	ptr1 := &pointer.Pointer{
		Node: blueprint.Node{
			NodeName: "ptr_service1",
		},
		Wrapped: container1,
	}

	ptr2 := &pointer.Pointer{
		Node: blueprint.Node{
			NodeName: "ptr_service2",
		},
		Wrapped: container2,
	}

	// Process arg nodes with service references
	args := []ir.IRNode{ptr1, ptr2}
	processArgNodes(workspace, args)

	// Check that environment variables were added
	assert.Equal(t, "service1:8080", container1.EnvironmentVariables["SERVICE1_ADDR"])
	assert.Equal(t, "service2:9090", container1.EnvironmentVariables["SERVICE2_ADDR"])
}

func TestManifestYAMLStructure(t *testing.T) {
	ctx := context.Background()
	deployment := createTestDeployment()
	mockCtx := newMockBuildContext("/tmp/test")

	err := deployment.GenerateArtifacts(ctx, mockCtx)
	require.NoError(t, err)

	// Parse and validate deployment YAML structure
	deploymentYAML, ok := mockCtx.GetWrittenFile(filepath.Join("/tmp/test", "kubernetes", "manifests", "test-deployment-deployment.yaml"))
	require.True(t, ok)

	var deploymentObj map[string]interface{}
	err = yaml.Unmarshal([]byte(deploymentYAML), &deploymentObj)
	require.NoError(t, err, "Deployment YAML should be valid")

	// Validate top-level fields
	assert.Equal(t, "apps/v1", deploymentObj["apiVersion"])
	assert.Equal(t, "Deployment", deploymentObj["kind"])

	// Validate metadata
	metadata := deploymentObj["metadata"].(map[string]interface{})
	assert.Equal(t, "test-deployment", metadata["name"])
	assert.Equal(t, "test-namespace", metadata["namespace"])

	// Validate spec
	spec := deploymentObj["spec"].(map[string]interface{})
	assert.Equal(t, 3, spec["replicas"])
}

func TestMultipleContainersDeployment(t *testing.T) {
	ctx := context.Background()

	// Create deployment with multiple containers
	container1 := &docker.Container{
		Node:      blueprint.Node{NodeName: "frontend"},
		ImageName: "frontend:latest",
		Ports: map[string]*address.BindConfig{
			"http": {Port: 3000},
		},
		EnvironmentVariables: map[string]string{
			"API_URL": "http://backend:8080",
		},
	}

	container2 := &docker.Container{
		Node:      blueprint.Node{NodeName: "backend"},
		ImageName: "backend:latest",
		Ports: map[string]*address.BindConfig{
			"http": {Port: 8080},
		},
		EnvironmentVariables: map[string]string{
			"DB_HOST": "database",
		},
	}

	deployment := &KubernetesDeployment{
		Node:           blueprint.Node{NodeName: "multi-container-deployment"},
		DeploymentName: "multi-container",
		Namespace:      "production",
		Replicas:       2,
		Containers:     []any{container1, container2},
		ClusterConfig: ClusterConfiguration{
			Endpoint: "https://prod.k8s.example.com",
		},
	}

	mockCtx := newMockBuildContext("/tmp/test")
	err := deployment.GenerateArtifacts(ctx, mockCtx)
	require.NoError(t, err)

	// Check deployment has both containers
	deploymentYAML, ok := mockCtx.GetWrittenFile(filepath.Join("/tmp/test", "kubernetes", "manifests", "multi-container-deployment.yaml"))
	require.True(t, ok)
	assert.Contains(t, deploymentYAML, "name: frontend")
	assert.Contains(t, deploymentYAML, "name: backend")
	assert.Contains(t, deploymentYAML, "image: frontend:latest")
	assert.Contains(t, deploymentYAML, "image: backend:latest")

	// Check services for both containers
	_, ok = mockCtx.GetWrittenFile(filepath.Join("/tmp/test", "kubernetes", "manifests", "frontend-service.yaml"))
	require.True(t, ok, "Frontend service should be generated")

	_, ok = mockCtx.GetWrittenFile(filepath.Join("/tmp/test", "kubernetes", "manifests", "backend-service.yaml"))
	require.True(t, ok, "Backend service should be generated")

	// Check ConfigMap has environment variables from both containers
	configMapYAML, ok := mockCtx.GetWrittenFile(filepath.Join("/tmp/test", "kubernetes", "manifests", "multi-container-configmap.yaml"))
	require.True(t, ok)
	assert.Contains(t, configMapYAML, "API_URL")
	assert.Contains(t, configMapYAML, "DB_HOST")
}

func TestEmptyDeployment(t *testing.T) {
	ctx := context.Background()

	// Create deployment with no containers
	deployment := &KubernetesDeployment{
		Node:           blueprint.Node{NodeName: "empty-deployment"},
		DeploymentName: "empty",
		Namespace:      "default",
		Replicas:       1,
		Containers:     []any{},
	}

	mockCtx := newMockBuildContext("/tmp/test")
	err := deployment.GenerateArtifacts(ctx, mockCtx)
	require.NoError(t, err)

	// Check that deployment manifest was still generated
	deploymentYAML, ok := mockCtx.GetWrittenFile(filepath.Join("/tmp/test", "kubernetes", "manifests", "empty-deployment.yaml"))
	require.True(t, ok)
	assert.Contains(t, deploymentYAML, "kind: Deployment")

	// Check that no services were generated
	files := mockCtx.GetAllFiles()
	for filename := range files {
		assert.False(t, strings.Contains(filename, "-service.yaml"), "No service files should be generated for empty deployment")
	}
}

func TestDeploymentScripts(t *testing.T) {
	ctx := context.Background()
	deployment := createTestDeployment()
	mockCtx := newMockBuildContext("/tmp/test")

	err := deployment.GenerateArtifacts(ctx, mockCtx)
	require.NoError(t, err)

	// Test Linux deployment script
	linuxScript, ok := mockCtx.GetWrittenFile(filepath.Join("/tmp/test", "kubernetes", "deploy.sh"))
	require.True(t, ok)
	assert.Contains(t, linuxScript, "#!/bin/bash")
	assert.Contains(t, linuxScript, "kubectl apply")
	assert.Contains(t, linuxScript, "KUBECONFIG=/path/to/kubeconfig")
	assert.Contains(t, linuxScript, "CLUSTER_ENDPOINT=https://k8s.example.com")

	// Test Windows deployment script
	windowsScript, ok := mockCtx.GetWrittenFile(filepath.Join("/tmp/test", "kubernetes", "deploy.bat"))
	require.True(t, ok)
	assert.Contains(t, windowsScript, "@echo off")
	assert.Contains(t, windowsScript, "kubectl apply")
	assert.Contains(t, windowsScript, "set KUBECONFIG=/path/to/kubeconfig")
}

func TestContainerWithoutPorts(t *testing.T) {
	ctx := context.Background()

	// Create container without exposed ports
	container := &docker.Container{
		Node:                 blueprint.Node{NodeName: "worker"},
		ImageName:            "worker:latest",
		Ports:                map[string]*address.BindConfig{}, // No ports
		EnvironmentVariables: map[string]string{"WORKER_ID": "123"},
	}

	deployment := &KubernetesDeployment{
		Node:           blueprint.Node{NodeName: "worker-deployment"},
		DeploymentName: "worker",
		Namespace:      "default",
		Replicas:       5,
		Containers:     []any{container},
	}

	mockCtx := newMockBuildContext("/tmp/test")
	err := deployment.GenerateArtifacts(ctx, mockCtx)
	require.NoError(t, err)

	// Check that no service was generated
	files := mockCtx.GetAllFiles()
	for filename := range files {
		assert.False(t, strings.Contains(filename, "worker-service.yaml"), "No service should be generated for container without ports")
	}

	// Check deployment was still generated
	_, ok := mockCtx.GetWrittenFile(filepath.Join("/tmp/test", "kubernetes", "manifests", "worker-deployment.yaml"))
	require.True(t, ok, "Deployment should be generated even without ports")
}

func TestGenerateArtifactsWithInvalidContainer(t *testing.T) {
	ctx := context.Background()

	// Create deployment with non-container node
	deployment := &KubernetesDeployment{
		Node:           blueprint.Node{NodeName: "invalid-deployment"},
		DeploymentName: "invalid",
		Namespace:      "default",
		Containers:     []any{"not-a-container"}, // Invalid container
	}

	mockCtx := newMockBuildContext("/tmp/test")
	err := deployment.GenerateArtifacts(ctx, mockCtx)
	require.Error(t, err, "Should error with invalid container type")
	assert.Contains(t, err.Error(), "expected docker.Container")
}
