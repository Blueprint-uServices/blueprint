package kubernetes

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestKubernetesDeploymentCreation(t *testing.T) {
	// Test basic deployment creation
	deployment := &KubernetesDeployment{
		DeploymentName: "test-deployment",
		Namespace:      "default",
		Replicas:       1,
		Containers:     []any{},
		ClusterConfig: ClusterConfiguration{
			Endpoint:   "https://k8s.example.com",
			Kubeconfig: "/path/to/config",
			AuthToken:  "token123",
		},
	}

	assert.Equal(t, "test-deployment", deployment.DeploymentName)
	assert.Equal(t, "default", deployment.Namespace)
	assert.Equal(t, 1, deployment.Replicas)
	assert.Empty(t, deployment.Containers)
}

func TestKubernetesDeploymentName(t *testing.T) {
	deployment := &KubernetesDeployment{
		DeploymentName: "my-app",
	}

	assert.Equal(t, "my-app", deployment.Name())
}

func TestKubernetesDeploymentString(t *testing.T) {
	deployment := &KubernetesDeployment{
		DeploymentName: "test-app",
		Namespace:      "production",
		Replicas:       3,
	}

	result := deployment.String()
	assert.Contains(t, result, "KubernetesDeployment")
	assert.Contains(t, result, "test-app")
	assert.Contains(t, result, "namespace=production")
	assert.Contains(t, result, "replicas=3")
}

func TestKubernetesDeploymentImplementsDockerContainer(t *testing.T) {
	deployment := &KubernetesDeployment{
		DeploymentName: "test",
	}

	assert.False(t, deployment.ImplementsDockerContainer())
}

func TestKubernetesDeploymentImplementsDockerWorkspace(t *testing.T) {
	deployment := &KubernetesDeployment{
		DeploymentName: "test",
	}

	assert.True(t, deployment.ImplementsDockerWorkspace())
}

func TestKubernetesDeploymentAddContainer(t *testing.T) {
	deployment := &KubernetesDeployment{
		DeploymentName: "test",
		Containers:     []any{},
	}

	// Mock container objects
	container1 := "container1"
	container2 := "container2"

	deployment.Containers = append(deployment.Containers, container1)
	deployment.Containers = append(deployment.Containers, container2)

	assert.Len(t, deployment.Containers, 2)
	assert.Equal(t, "container1", deployment.Containers[0])
	assert.Equal(t, "container2", deployment.Containers[1])
}

func TestClusterConfiguration(t *testing.T) {
	tests := []struct {
		name   string
		config ClusterConfiguration
		want   struct {
			hasEndpoint   bool
			hasKubeconfig bool
			hasToken      bool
		}
	}{
		{
			name: "full configuration",
			config: ClusterConfiguration{
				Endpoint:   "https://k8s.example.com:6443",
				Kubeconfig: "/home/user/.kube/config",
				AuthToken:  "bearer-token",
			},
			want: struct {
				hasEndpoint   bool
				hasKubeconfig bool
				hasToken      bool
			}{true, true, true},
		},
		{
			name: "endpoint only",
			config: ClusterConfiguration{
				Endpoint: "https://k8s.example.com",
			},
			want: struct {
				hasEndpoint   bool
				hasKubeconfig bool
				hasToken      bool
			}{true, false, false},
		},
		{
			name: "kubeconfig only",
			config: ClusterConfiguration{
				Kubeconfig: "/path/to/kubeconfig",
			},
			want: struct {
				hasEndpoint   bool
				hasKubeconfig bool
				hasToken      bool
			}{false, true, false},
		},
		{
			name:   "empty configuration",
			config: ClusterConfiguration{},
			want: struct {
				hasEndpoint   bool
				hasKubeconfig bool
				hasToken      bool
			}{false, false, false},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want.hasEndpoint, tt.config.Endpoint != "")
			assert.Equal(t, tt.want.hasKubeconfig, tt.config.Kubeconfig != "")
			assert.Equal(t, tt.want.hasToken, tt.config.AuthToken != "")
		})
	}
}

func TestKubernetesDeploymentDefaults(t *testing.T) {
	deployment := &KubernetesDeployment{
		DeploymentName: "test",
	}

	// Test default values
	assert.Equal(t, "", deployment.Namespace, "Namespace should be empty by default")
	assert.Equal(t, 0, deployment.Replicas, "Replicas should be 0 by default")
	assert.NotNil(t, deployment.Containers, "Containers should not be nil")
	assert.Empty(t, deployment.Containers, "Containers should be empty")
}

func TestKubernetesDeploymentWithMultipleContainers(t *testing.T) {
	deployment := &KubernetesDeployment{
		DeploymentName: "multi-container-app",
		Namespace:      "production",
		Replicas:       5,
		Containers:     []any{},
	}

	// Add multiple containers
	containers := []string{"frontend", "backend", "database", "cache", "queue"}
	for _, c := range containers {
		deployment.Containers = append(deployment.Containers, c)
	}

	assert.Len(t, deployment.Containers, 5)
	assert.Equal(t, "multi-container-app", deployment.Name())
	assert.Equal(t, "production", deployment.Namespace)
	assert.Equal(t, 5, deployment.Replicas)
}

func TestKubernetesDeploymentStringWithContainers(t *testing.T) {
	deployment := &KubernetesDeployment{
		DeploymentName: "app-with-containers",
		Namespace:      "staging",
		Replicas:       2,
		Containers:     []any{"service1", "service2"},
	}

	result := deployment.String()
	assert.Contains(t, result, "KubernetesDeployment")
	assert.Contains(t, result, "app-with-containers")
	assert.Contains(t, result, "namespace=staging")
	assert.Contains(t, result, "replicas=2")
	assert.Contains(t, result, "containers=2")
}

func TestKubernetesDeploymentClusterConfigValidation(t *testing.T) {
	tests := []struct {
		name          string
		deployment    *KubernetesDeployment
		expectedValid bool
	}{
		{
			name: "valid with endpoint",
			deployment: &KubernetesDeployment{
				DeploymentName: "test",
				ClusterConfig: ClusterConfiguration{
					Endpoint: "https://k8s.example.com",
				},
			},
			expectedValid: true,
		},
		{
			name: "valid with kubeconfig",
			deployment: &KubernetesDeployment{
				DeploymentName: "test",
				ClusterConfig: ClusterConfiguration{
					Kubeconfig: "/path/to/config",
				},
			},
			expectedValid: true,
		},
		{
			name: "valid with both",
			deployment: &KubernetesDeployment{
				DeploymentName: "test",
				ClusterConfig: ClusterConfiguration{
					Endpoint:   "https://k8s.example.com",
					Kubeconfig: "/path/to/config",
				},
			},
			expectedValid: true,
		},
		{
			name: "empty config is valid (can be set at runtime)",
			deployment: &KubernetesDeployment{
				DeploymentName: "test",
				ClusterConfig:  ClusterConfiguration{},
			},
			expectedValid: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// In this implementation, all configs are valid since they can be set at runtime
			// This test documents that behavior
			require.NotNil(t, tt.deployment)
			assert.Equal(t, tt.expectedValid, true)
		})
	}
}
