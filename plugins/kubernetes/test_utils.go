package kubernetes

import (
	"fmt"
	"path/filepath"
	"strings"
	"testing"

	"github.com/blueprint-uservices/blueprint/blueprint/pkg/blueprint"
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/coreplugins/address"
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/wiring"
	"github.com/blueprint-uservices/blueprint/plugins/docker"
	"github.com/blueprint-uservices/blueprint/plugins/golang"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

// TestHelpers provides utility functions for testing the Kubernetes plugin
type TestHelpers struct {
	t *testing.T
}

// NewTestHelpers creates a new instance of TestHelpers
func NewTestHelpers(t *testing.T) *TestHelpers {
	return &TestHelpers{t: t}
}

// CreateTestService creates a golang.Service for testing
func (h *TestHelpers) CreateTestService(name string, options ...ServiceOption) *golang.Service {
	service := &golang.Service{
		Node: blueprint.Node{
			NodeName: name,
		},
	}

	// Apply options
	for _, opt := range options {
		opt(service)
	}

	return service
}

// ServiceOption is a function that configures a service
type ServiceOption func(*golang.Service)

// CreateTestContainer creates a docker.Container for testing
func (h *TestHelpers) CreateTestContainer(name, image string, options ...ContainerOption) *docker.Container {
	container := &docker.Container{
		Node: blueprint.Node{
			NodeName: name,
		},
		ImageName:            image,
		Ports:                make(map[string]*address.BindConfig),
		EnvironmentVariables: make(map[string]string),
	}

	// Apply options
	for _, opt := range options {
		opt(container)
	}

	return container
}

// ContainerOption is a function that configures a container
type ContainerOption func(*docker.Container)

// WithPort adds a port to a container
func WithPort(name string, port int) ContainerOption {
	return func(c *docker.Container) {
		c.Ports[name] = &address.BindConfig{
			Port: port,
		}
	}
}

// WithEnvVar adds an environment variable to a container
func WithEnvVar(key, value string) ContainerOption {
	return func(c *docker.Container) {
		c.EnvironmentVariables[key] = value
	}
}

// WithCommand sets the command for a container
func WithCommand(command []string) ContainerOption {
	return func(c *docker.Container) {
		c.Command = command
	}
}

// CreateTestDeployment creates a KubernetesDeployment for testing
func (h *TestHelpers) CreateTestDeployment(name string, options ...DeploymentOption) *KubernetesDeployment {
	deployment := &KubernetesDeployment{
		Node: blueprint.Node{
			NodeName: name,
		},
		DeploymentName: name,
		Namespace:      "default",
		Replicas:       1,
		Containers:     []any{},
		ClusterConfig:  ClusterConfiguration{},
	}

	// Apply options
	for _, opt := range options {
		opt(deployment)
	}

	return deployment
}

// DeploymentOption is a function that configures a deployment
type DeploymentOption func(*KubernetesDeployment)

// WithNamespace sets the namespace for a deployment
func WithNamespace(namespace string) DeploymentOption {
	return func(d *KubernetesDeployment) {
		d.Namespace = namespace
	}
}

// WithReplicas sets the number of replicas for a deployment
func WithReplicas(replicas int) DeploymentOption {
	return func(d *KubernetesDeployment) {
		d.Replicas = replicas
	}
}

// WithClusterConfig sets the cluster configuration for a deployment
func WithClusterConfig(endpoint, kubeconfig, token string) DeploymentOption {
	return func(d *KubernetesDeployment) {
		d.ClusterConfig = ClusterConfiguration{
			Endpoint:   endpoint,
			Kubeconfig: kubeconfig,
			AuthToken:  token,
		}
	}
}

// WithContainer adds a container to a deployment
func WithContainer(container any) DeploymentOption {
	return func(d *KubernetesDeployment) {
		d.Containers = append(d.Containers, container)
	}
}

// CreateTestWiringSpec creates a wiring spec for testing
func (h *TestHelpers) CreateTestWiringSpec(name string) wiring.WiringSpec {
	return wiring.NewWiringSpec(name)
}

// AssertYAMLValid validates that a string contains valid YAML
func (h *TestHelpers) AssertYAMLValid(yamlContent string) {
	var data interface{}
	err := yaml.Unmarshal([]byte(yamlContent), &data)
	require.NoError(h.t, err, "YAML content should be valid")
}

// AssertYAMLContains checks if YAML content contains expected fields
func (h *TestHelpers) AssertYAMLContains(yamlContent string, expectedFields map[string]interface{}) {
	var data map[string]interface{}
	err := yaml.Unmarshal([]byte(yamlContent), &data)
	require.NoError(h.t, err, "YAML content should be valid")

	for key, expectedValue := range expectedFields {
		actualValue := h.getNestedValue(data, key)
		require.NotNil(h.t, actualValue, "Field %s should exist in YAML", key)
		require.Equal(h.t, expectedValue, actualValue, "Field %s should have expected value", key)
	}
}

// getNestedValue retrieves a nested value from a map using dot notation
func (h *TestHelpers) getNestedValue(data map[string]interface{}, path string) interface{} {
	parts := strings.Split(path, ".")
	current := interface{}(data)

	for _, part := range parts {
		switch v := current.(type) {
		case map[string]interface{}:
			current = v[part]
		case map[interface{}]interface{}:
			current = v[part]
		default:
			return nil
		}
	}

	return current
}

// AssertFileGenerated checks if a file was generated in the mock context
func (h *TestHelpers) AssertFileGenerated(ctx *mockBuildContext, expectedPath string) string {
	content, ok := ctx.GetWrittenFile(expectedPath)
	require.True(h.t, ok, "File %s should be generated", expectedPath)
	return content
}

// AssertFileNotGenerated checks if a file was not generated
func (h *TestHelpers) AssertFileNotGenerated(ctx *mockBuildContext, unexpectedPath string) {
	_, ok := ctx.GetWrittenFile(unexpectedPath)
	require.False(h.t, ok, "File %s should not be generated", unexpectedPath)
}

// AssertManifestCount checks the number of manifest files generated
func (h *TestHelpers) AssertManifestCount(ctx *mockBuildContext, expectedCount int) {
	manifestCount := 0
	for path := range ctx.GetAllFiles() {
		if strings.Contains(path, "manifests/") && strings.HasSuffix(path, ".yaml") {
			manifestCount++
		}
	}
	require.Equal(h.t, expectedCount, manifestCount, "Should generate expected number of manifest files")
}

// CreateMicroserviceSetup creates a typical microservice setup for testing
func (h *TestHelpers) CreateMicroserviceSetup() (*KubernetesDeployment, map[string]*docker.Container) {
	containers := make(map[string]*docker.Container)

	// Frontend service
	containers["frontend"] = h.CreateTestContainer("frontend", "frontend:v1.0",
		WithPort("http", 3000),
		WithEnvVar("API_URL", "http://api:8080"),
		WithEnvVar("NODE_ENV", "production"),
	)

	// API service
	containers["api"] = h.CreateTestContainer("api", "api:v1.0",
		WithPort("http", 8080),
		WithEnvVar("DB_HOST", "database"),
		WithEnvVar("DB_PORT", "5432"),
		WithEnvVar("CACHE_HOST", "redis"),
	)

	// Database service
	containers["database"] = h.CreateTestContainer("database", "postgres:13",
		WithPort("postgresql", 5432),
		WithEnvVar("POSTGRES_DB", "myapp"),
		WithEnvVar("POSTGRES_USER", "user"),
		WithEnvVar("POSTGRES_PASSWORD", "password"),
	)

	// Cache service
	containers["redis"] = h.CreateTestContainer("redis", "redis:6-alpine",
		WithPort("redis", 6379),
	)

	// Create deployment with all containers
	deployment := h.CreateTestDeployment("microservice-app",
		WithNamespace("production"),
		WithReplicas(3),
		WithClusterConfig("https://k8s.prod.example.com", "/etc/kubeconfig", "prod-token"),
	)

	for _, container := range containers {
		deployment.Containers = append(deployment.Containers, container)
	}

	return deployment, containers
}

// ValidateServiceManifest validates a Kubernetes Service manifest
func (h *TestHelpers) ValidateServiceManifest(yamlContent string, expectedName string, expectedPort int) {
	h.AssertYAMLValid(yamlContent)
	h.AssertYAMLContains(yamlContent, map[string]interface{}{
		"apiVersion":        "v1",
		"kind":              "Service",
		"metadata.name":     expectedName,
		"spec.ports.0.port": expectedPort,
	})
}

// ValidateDeploymentManifest validates a Kubernetes Deployment manifest
func (h *TestHelpers) ValidateDeploymentManifest(yamlContent string, expectedName string, expectedReplicas int) {
	h.AssertYAMLValid(yamlContent)
	h.AssertYAMLContains(yamlContent, map[string]interface{}{
		"apiVersion":    "apps/v1",
		"kind":          "Deployment",
		"metadata.name": expectedName,
		"spec.replicas": expectedReplicas,
	})
}

// ValidateConfigMapManifest validates a Kubernetes ConfigMap manifest
func (h *TestHelpers) ValidateConfigMapManifest(yamlContent string, expectedName string) {
	h.AssertYAMLValid(yamlContent)
	h.AssertYAMLContains(yamlContent, map[string]interface{}{
		"apiVersion":    "v1",
		"kind":          "ConfigMap",
		"metadata.name": expectedName,
	})
}

// GenerateTestArtifacts generates artifacts and returns the mock context
func (h *TestHelpers) GenerateTestArtifacts(deployment *KubernetesDeployment, outputDir string) *mockBuildContext {
	ctx := newMockBuildContext(outputDir)
	err := deployment.GenerateArtifacts(nil, ctx)
	require.NoError(h.t, err, "Artifact generation should succeed")
	return ctx
}

// AssertScriptContains checks if a script file contains expected content
func (h *TestHelpers) AssertScriptContains(ctx *mockBuildContext, scriptPath string, expectedContent []string) {
	script := h.AssertFileGenerated(ctx, scriptPath)
	for _, content := range expectedContent {
		require.Contains(h.t, script, content, "Script should contain: %s", content)
	}
}

// CreateComplexDeploymentScenario creates a complex deployment scenario for integration testing
func (h *TestHelpers) CreateComplexDeploymentScenario() []*KubernetesDeployment {
	deployments := []*KubernetesDeployment{}

	// User-facing services deployment
	userServices := h.CreateTestDeployment("user-services",
		WithNamespace("frontend"),
		WithReplicas(5),
		WithContainer(h.CreateTestContainer("web", "nginx:latest",
			WithPort("http", 80),
			WithPort("https", 443),
		)),
		WithContainer(h.CreateTestContainer("app", "app:v2.0",
			WithPort("http", 8080),
			WithEnvVar("CONFIG_SERVER", "http://config:8888"),
		)),
	)
	deployments = append(deployments, userServices)

	// Backend services deployment
	backendServices := h.CreateTestDeployment("backend-services",
		WithNamespace("backend"),
		WithReplicas(3),
		WithContainer(h.CreateTestContainer("api", "api:v2.0",
			WithPort("http", 8080),
			WithPort("grpc", 9090),
		)),
		WithContainer(h.CreateTestContainer("worker", "worker:v2.0",
			WithEnvVar("QUEUE_URL", "amqp://rabbitmq:5672"),
		)),
	)
	deployments = append(deployments, backendServices)

	// Data layer deployment
	dataLayer := h.CreateTestDeployment("data-layer",
		WithNamespace("data"),
		WithReplicas(1),
		WithContainer(h.CreateTestContainer("postgres", "postgres:13",
			WithPort("postgresql", 5432),
		)),
		WithContainer(h.CreateTestContainer("redis", "redis:6",
			WithPort("redis", 6379),
		)),
		WithContainer(h.CreateTestContainer("elasticsearch", "elasticsearch:7.10",
			WithPort("http", 9200),
			WithPort("transport", 9300),
		)),
	)
	deployments = append(deployments, dataLayer)

	return deployments
}

// CompareYAMLStructure compares two YAML structures for equality
func (h *TestHelpers) CompareYAMLStructure(yaml1, yaml2 string) bool {
	var data1, data2 interface{}
	err1 := yaml.Unmarshal([]byte(yaml1), &data1)
	err2 := yaml.Unmarshal([]byte(yaml2), &data2)

	if err1 != nil || err2 != nil {
		return false
	}

	return fmt.Sprintf("%v", data1) == fmt.Sprintf("%v", data2)
}

// GetManifestPath returns the expected path for a manifest file
func (h *TestHelpers) GetManifestPath(outputDir, filename string) string {
	return filepath.Join(outputDir, "kubernetes", "manifests", filename)
}

// GetScriptPath returns the expected path for a script file
func (h *TestHelpers) GetScriptPath(outputDir, filename string) string {
	return filepath.Join(outputDir, "kubernetes", filename)
}
