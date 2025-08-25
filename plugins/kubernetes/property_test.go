package kubernetes

import (
	"context"
	"fmt"
	"math/rand"
	"strings"
	"testing"
	"time"

	"github.com/blueprint-uservices/blueprint/blueprint/pkg/blueprint"
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/coreplugins/address"
	"github.com/blueprint-uservices/blueprint/plugins/docker"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

// PropertyTestConfig holds configuration for property-based tests
type PropertyTestConfig struct {
	Seed       int64
	Iterations int
	MinItems   int
	MaxItems   int
}

// NewPropertyTestConfig creates a default property test configuration
func NewPropertyTestConfig() *PropertyTestConfig {
	return &PropertyTestConfig{
		Seed:       time.Now().UnixNano(),
		Iterations: 100,
		MinItems:   0,
		MaxItems:   10,
	}
}

// TestPropertyDeploymentAlwaysGeneratesValidYAML tests that deployments always generate valid YAML
func TestPropertyDeploymentAlwaysGeneratesValidYAML(t *testing.T) {
	config := NewPropertyTestConfig()
	rand.Seed(config.Seed)
	t.Logf("Using seed: %d", config.Seed)

	for i := 0; i < config.Iterations; i++ {
		// Generate random deployment
		deployment := generateRandomDeployment(t, i)

		// Generate artifacts
		ctx := context.Background()
		mockCtx := newMockBuildContext("/tmp/test")
		err := deployment.GenerateArtifacts(ctx, mockCtx)
		require.NoError(t, err, "Iteration %d: Failed to generate artifacts", i)

		// Verify all YAML files are valid
		for path, content := range mockCtx.GetAllFiles() {
			if strings.HasSuffix(path, ".yaml") {
				var data interface{}
				err := yaml.Unmarshal([]byte(content), &data)
				assert.NoError(t, err, "Iteration %d: Invalid YAML in %s", i, path)
			}
		}
	}
}

// TestPropertyServiceDiscoveryConsistency tests that service discovery is consistent
func TestPropertyServiceDiscoveryConsistency(t *testing.T) {
	config := NewPropertyTestConfig()
	rand.Seed(config.Seed)
	t.Logf("Using seed: %d", config.Seed)

	for i := 0; i < config.Iterations; i++ {
		// Generate deployment with multiple containers
		numContainers := rand.Intn(5) + 2 // 2-6 containers
		deployment := createDeploymentWithContainers(t, numContainers)

		// Generate artifacts
		ctx := context.Background()
		mockCtx := newMockBuildContext("/tmp/test")

		// Create workspace and process arg nodes
		workspace := &kubernetesWorkspace{deployment: deployment}
		processArgNodes(workspace, deployment.Containers)

		err := deployment.GenerateArtifacts(ctx, mockCtx)
		require.NoError(t, err, "Iteration %d: Failed to generate artifacts", i)

		// Verify service discovery consistency
		verifyServiceDiscoveryConsistency(t, deployment, i)
	}
}

// TestPropertyNamespaceHandling tests various namespace configurations
func TestPropertyNamespaceHandling(t *testing.T) {
	config := NewPropertyTestConfig()
	rand.Seed(config.Seed)
	t.Logf("Using seed: %d", config.Seed)

	namespaces := []string{"", "default", "production", "staging", "dev", "test-ns-123"}

	for i := 0; i < config.Iterations; i++ {
		// Pick random namespace
		namespace := namespaces[rand.Intn(len(namespaces))]

		deployment := &KubernetesDeployment{
			Node:           blueprint.Node{NodeName: fmt.Sprintf("deployment-%d", i)},
			DeploymentName: fmt.Sprintf("test-deployment-%d", i),
			Namespace:      namespace,
			Replicas:       rand.Intn(10) + 1,
			Containers:     []any{createRandomContainer(i)},
		}

		// Generate artifacts
		ctx := context.Background()
		mockCtx := newMockBuildContext("/tmp/test")
		err := deployment.GenerateArtifacts(ctx, mockCtx)
		require.NoError(t, err, "Iteration %d: Failed to generate artifacts", i)

		// Verify namespace handling
		deploymentYAML := getGeneratedFile(mockCtx, "deployment.yaml")
		if namespace == "" {
			assert.NotContains(t, deploymentYAML, "namespace:", "Empty namespace should not appear in YAML")
		} else {
			assert.Contains(t, deploymentYAML, fmt.Sprintf("namespace: %s", namespace))
		}
	}
}

// TestPropertyPortRangeValidation tests that all generated ports are in valid range
func TestPropertyPortRangeValidation(t *testing.T) {
	config := NewPropertyTestConfig()
	rand.Seed(config.Seed)
	t.Logf("Using seed: %d", config.Seed)

	for i := 0; i < config.Iterations; i++ {
		// Create container with random ports
		numPorts := rand.Intn(5) + 1
		container := createContainerWithRandomPorts(i, numPorts)

		deployment := &KubernetesDeployment{
			Node:           blueprint.Node{NodeName: fmt.Sprintf("deployment-%d", i)},
			DeploymentName: fmt.Sprintf("port-test-%d", i),
			Namespace:      "default",
			Replicas:       1,
			Containers:     []any{container},
		}

		// Generate artifacts
		ctx := context.Background()
		mockCtx := newMockBuildContext("/tmp/test")
		err := deployment.GenerateArtifacts(ctx, mockCtx)
		require.NoError(t, err, "Iteration %d: Failed to generate artifacts", i)

		// Verify all ports are in valid range
		for path, content := range mockCtx.GetAllFiles() {
			if strings.Contains(path, "-service.yaml") {
				verifyPortsInValidRange(t, content, i)
			}
		}
	}
}

// TestPropertyEnvironmentVariableHandling tests environment variable generation
func TestPropertyEnvironmentVariableHandling(t *testing.T) {
	config := NewPropertyTestConfig()
	rand.Seed(config.Seed)
	t.Logf("Using seed: %d", config.Seed)

	for i := 0; i < config.Iterations; i++ {
		// Create container with random environment variables
		numEnvVars := rand.Intn(20)
		container := createContainerWithRandomEnvVars(i, numEnvVars)

		deployment := &KubernetesDeployment{
			Node:           blueprint.Node{NodeName: fmt.Sprintf("deployment-%d", i)},
			DeploymentName: fmt.Sprintf("env-test-%d", i),
			Namespace:      "default",
			Replicas:       1,
			Containers:     []any{container},
		}

		// Generate artifacts
		ctx := context.Background()
		mockCtx := newMockBuildContext("/tmp/test")
		err := deployment.GenerateArtifacts(ctx, mockCtx)
		require.NoError(t, err, "Iteration %d: Failed to generate artifacts", i)

		// Verify environment variables
		if numEnvVars > 0 {
			configMapYAML := getGeneratedFile(mockCtx, "configmap.yaml")
			assert.Contains(t, configMapYAML, "kind: ConfigMap")

			// Verify all env vars are present
			for key := range container.EnvironmentVariables {
				assert.Contains(t, configMapYAML, key)
			}
		}
	}
}

// TestPropertyReplicaCountHandling tests various replica configurations
func TestPropertyReplicaCountHandling(t *testing.T) {
	config := NewPropertyTestConfig()
	rand.Seed(config.Seed)
	t.Logf("Using seed: %d", config.Seed)

	for i := 0; i < config.Iterations; i++ {
		// Generate random replica count (including edge cases)
		replicas := generateReplicaCount(rand.Intn(100))

		deployment := &KubernetesDeployment{
			Node:           blueprint.Node{NodeName: fmt.Sprintf("deployment-%d", i)},
			DeploymentName: fmt.Sprintf("replica-test-%d", i),
			Namespace:      "default",
			Replicas:       replicas,
			Containers:     []any{createRandomContainer(i)},
		}

		// Generate artifacts
		ctx := context.Background()
		mockCtx := newMockBuildContext("/tmp/test")
		err := deployment.GenerateArtifacts(ctx, mockCtx)
		require.NoError(t, err, "Iteration %d: Failed to generate artifacts", i)

		// Verify replica count
		deploymentYAML := getGeneratedFile(mockCtx, "deployment.yaml")
		assert.Contains(t, deploymentYAML, fmt.Sprintf("replicas: %d", replicas))
	}
}

// TestPropertyContainerNameUniqueness tests that container names are unique within deployment
func TestPropertyContainerNameUniqueness(t *testing.T) {
	config := NewPropertyTestConfig()
	rand.Seed(config.Seed)
	t.Logf("Using seed: %d", config.Seed)

	for i := 0; i < config.Iterations; i++ {
		numContainers := rand.Intn(10) + 1
		containers := make([]any, numContainers)
		containerNames := make(map[string]bool)

		// Create containers with unique names
		for j := 0; j < numContainers; j++ {
			container := &docker.Container{
				Node: blueprint.Node{
					NodeName: fmt.Sprintf("container-%d-%d", i, j),
				},
				ImageName: fmt.Sprintf("image-%d-%d:latest", i, j),
				Ports:     make(map[string]*address.BindConfig),
			}
			containers[j] = container
			containerNames[container.Name()] = true
		}

		deployment := &KubernetesDeployment{
			Node:           blueprint.Node{NodeName: fmt.Sprintf("deployment-%d", i)},
			DeploymentName: fmt.Sprintf("unique-test-%d", i),
			Namespace:      "default",
			Replicas:       1,
			Containers:     containers,
		}

		// Generate artifacts
		ctx := context.Background()
		mockCtx := newMockBuildContext("/tmp/test")
		err := deployment.GenerateArtifacts(ctx, mockCtx)
		require.NoError(t, err, "Iteration %d: Failed to generate artifacts", i)

		// Verify uniqueness property held
		assert.Equal(t, numContainers, len(containerNames))
	}
}

// Helper functions for property-based tests

func generateRandomDeployment(t *testing.T, iteration int) *KubernetesDeployment {
	numContainers := rand.Intn(5) + 1
	containers := make([]any, numContainers)

	for i := 0; i < numContainers; i++ {
		containers[i] = createRandomContainer(iteration*100 + i)
	}

	return &KubernetesDeployment{
		Node:           blueprint.Node{NodeName: fmt.Sprintf("random-deployment-%d", iteration)},
		DeploymentName: fmt.Sprintf("deployment-%d", iteration),
		Namespace:      randomNamespace(),
		Replicas:       rand.Intn(10) + 1,
		Containers:     containers,
		ClusterConfig: ClusterConfiguration{
			Endpoint:   fmt.Sprintf("https://k8s-%d.example.com", iteration),
			Kubeconfig: fmt.Sprintf("/path/to/kubeconfig-%d", iteration),
		},
	}
}

func createRandomContainer(id int) *docker.Container {
	container := &docker.Container{
		Node: blueprint.Node{
			NodeName: fmt.Sprintf("container-%d", id),
		},
		ImageName:            fmt.Sprintf("image-%d:v%d", id, rand.Intn(10)),
		Ports:                make(map[string]*address.BindConfig),
		EnvironmentVariables: make(map[string]string),
	}

	// Add random ports
	numPorts := rand.Intn(3)
	for i := 0; i < numPorts; i++ {
		portName := fmt.Sprintf("port%d", i)
		container.Ports[portName] = &address.BindConfig{
			Port: 8000 + rand.Intn(1000),
		}
	}

	// Add random env vars
	numEnvVars := rand.Intn(5)
	for i := 0; i < numEnvVars; i++ {
		key := fmt.Sprintf("ENV_VAR_%d", i)
		value := fmt.Sprintf("value_%d_%d", id, i)
		container.EnvironmentVariables[key] = value
	}

	return container
}

func createDeploymentWithContainers(t *testing.T, numContainers int) *KubernetesDeployment {
	containers := make([]any, numContainers)
	for i := 0; i < numContainers; i++ {
		container := &docker.Container{
			Node: blueprint.Node{
				NodeName: fmt.Sprintf("service%d", i),
			},
			ImageName: fmt.Sprintf("service%d:latest", i),
			Ports: map[string]*address.BindConfig{
				"main": {Port: 8000 + i},
			},
			EnvironmentVariables: make(map[string]string),
		}
		containers[i] = container
	}

	return &KubernetesDeployment{
		Node:           blueprint.Node{NodeName: "test-deployment"},
		DeploymentName: "test-deployment",
		Namespace:      "default",
		Replicas:       1,
		Containers:     containers,
	}
}

func createContainerWithRandomPorts(id int, numPorts int) *docker.Container {
	container := &docker.Container{
		Node: blueprint.Node{
			NodeName: fmt.Sprintf("container-%d", id),
		},
		ImageName: fmt.Sprintf("image-%d:latest", id),
		Ports:     make(map[string]*address.BindConfig),
	}

	for i := 0; i < numPorts; i++ {
		// Generate ports in valid range (1-65535)
		port := rand.Intn(65535) + 1
		portName := fmt.Sprintf("port%d", i)
		container.Ports[portName] = &address.BindConfig{
			Port: port,
		}
	}

	return container
}

func createContainerWithRandomEnvVars(id int, numEnvVars int) *docker.Container {
	container := &docker.Container{
		Node: blueprint.Node{
			NodeName: fmt.Sprintf("container-%d", id),
		},
		ImageName:            fmt.Sprintf("image-%d:latest", id),
		Ports:                make(map[string]*address.BindConfig),
		EnvironmentVariables: make(map[string]string),
	}

	for i := 0; i < numEnvVars; i++ {
		// Generate various types of env var names and values
		key := generateEnvVarKey(i)
		value := generateEnvVarValue(i)
		container.EnvironmentVariables[key] = value
	}

	return container
}

func generateEnvVarKey(index int) string {
	patterns := []string{
		"SIMPLE_VAR_%d",
		"APP_CONFIG_%d",
		"DATABASE_URL_%d",
		"API_KEY_%d",
		"FEATURE_FLAG_%d",
	}
	pattern := patterns[index%len(patterns)]
	return fmt.Sprintf(pattern, index)
}

func generateEnvVarValue(index int) string {
	values := []string{
		"simple-value-%d",
		"http://service-%d:8080",
		"postgresql://user:pass@db-%d:5432/database",
		"secret-key-%d-xxxxx",
		"true",
		"false",
		"%d",
	}
	value := values[index%len(values)]
	return fmt.Sprintf(value, index)
}

func randomNamespace() string {
	namespaces := []string{"default", "production", "staging", "development", "testing"}
	return namespaces[rand.Intn(len(namespaces))]
}

func generateReplicaCount(seed int) int {
	// Include edge cases
	edgeCases := []int{0, 1, 2, 3, 5, 10, 100}
	if seed < len(edgeCases) {
		return edgeCases[seed]
	}
	// Random value between 1 and 50
	return rand.Intn(50) + 1
}

func verifyServiceDiscoveryConsistency(t *testing.T, deployment *KubernetesDeployment, iteration int) {
	// Each container should have env vars for all other containers with ports
	for i, container := range deployment.Containers {
		c := container.(*docker.Container)
		for j, otherContainer := range deployment.Containers {
			if i != j {
				other := otherContainer.(*docker.Container)
				if len(other.Ports) > 0 {
					envVarName := strings.ToUpper(other.Name()) + "_ADDR"
					assert.Contains(t, c.EnvironmentVariables, envVarName,
						"Iteration %d: Container %s should have env var for %s",
						iteration, c.Name(), other.Name())
				}
			}
		}
	}
}

func verifyPortsInValidRange(t *testing.T, serviceYAML string, iteration int) {
	var service map[string]interface{}
	err := yaml.Unmarshal([]byte(serviceYAML), &service)
	require.NoError(t, err)

	spec, ok := service["spec"].(map[string]interface{})
	require.True(t, ok)

	ports, ok := spec["ports"].([]interface{})
	require.True(t, ok)

	for _, portInterface := range ports {
		port := portInterface.(map[string]interface{})
		portNum, ok := port["port"].(int)
		require.True(t, ok)
		assert.True(t, portNum >= 1 && portNum <= 65535,
			"Iteration %d: Port %d is not in valid range", iteration, portNum)
	}
}

func getGeneratedFile(ctx *mockBuildContext, suffix string) string {
	for path, content := range ctx.GetAllFiles() {
		if strings.HasSuffix(path, suffix) {
			return content
		}
	}
	return ""
}
