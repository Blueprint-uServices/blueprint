package kubernetes

import (
	"context"
	"testing"

	"github.com/blueprint-uservices/blueprint/blueprint/pkg/coreplugins/pointer"
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/ir"
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/wiring"
	"github.com/blueprint-uservices/blueprint/plugins/clientpool"
	"github.com/blueprint-uservices/blueprint/plugins/docker"
	"github.com/blueprint-uservices/blueprint/plugins/golang"
	"github.com/blueprint-uservices/blueprint/plugins/grpc"
	"github.com/blueprint-uservices/blueprint/plugins/http"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestIntegrationWithGolangService tests integration with golang.Service nodes
func TestIntegrationWithGolangService(t *testing.T) {
	helpers := NewTestHelpers(t)
	spec := helpers.CreateTestWiringSpec("GolangServiceIntegration")

	// Create a Golang service
	service := golang.Service(spec, "myservice")
	golang.Deploy(spec, service)

	// Create Kubernetes deployment and add the service
	deployment := NewDeployment(spec, "k8s-deployment")
	AddContainerToDeployment(spec, deployment, service)
	ConfigureCluster(spec, deployment, "https://k8s.example.com", "/path/to/kubeconfig", "")

	// Build IR
	app, err := buildTestApp(t, spec, deployment)
	require.NoError(t, err)

	// Find the deployment node
	deploymentNode := findNodeOfType[*KubernetesDeployment](app)
	require.NotNil(t, deploymentNode)

	// Verify the service was added as a container
	assert.Len(t, deploymentNode.Containers, 1)
	container, ok := deploymentNode.Containers[0].(*docker.Container)
	require.True(t, ok)
	assert.Equal(t, "myservice", container.Name())
}

// TestIntegrationWithHTTPServer tests integration with HTTP server
func TestIntegrationWithHTTPServer(t *testing.T) {
	helpers := NewTestHelpers(t)
	spec := helpers.CreateTestWiringSpec("HTTPServerIntegration")

	// Create a service with HTTP server
	service := golang.Service(spec, "api-service")
	httpServer := http.HTTPServer(spec, service, "8080")
	golang.Deploy(spec, httpServer)

	// Create Kubernetes deployment
	deployment := NewDeployment(spec, "api-deployment")
	AddContainerToDeployment(spec, deployment, httpServer)

	// Build IR
	app, err := buildTestApp(t, spec, deployment)
	require.NoError(t, err)

	// Find nodes
	deploymentNode := findNodeOfType[*KubernetesDeployment](app)
	require.NotNil(t, deploymentNode)

	// Verify HTTP port was exposed
	container, ok := deploymentNode.Containers[0].(*docker.Container)
	require.True(t, ok)
	assert.Contains(t, container.Ports, "http")
	assert.Equal(t, 8080, container.Ports["http"].Port)
}

// TestIntegrationWithGRPCServer tests integration with gRPC server
func TestIntegrationWithGRPCServer(t *testing.T) {
	helpers := NewTestHelpers(t)
	spec := helpers.CreateTestWiringSpec("GRPCServerIntegration")

	// Create a service with gRPC server
	service := golang.Service(spec, "grpc-service")
	grpcServer := grpc.GRPCServer(spec, service, "9090")
	golang.Deploy(spec, grpcServer)

	// Create Kubernetes deployment
	deployment := NewDeployment(spec, "grpc-deployment")
	AddContainerToDeployment(spec, deployment, grpcServer)

	// Build IR
	app, err := buildTestApp(t, spec, deployment)
	require.NoError(t, err)

	// Find nodes
	deploymentNode := findNodeOfType[*KubernetesDeployment](app)
	require.NotNil(t, deploymentNode)

	// Verify gRPC port was exposed
	container, ok := deploymentNode.Containers[0].(*docker.Container)
	require.True(t, ok)
	assert.Contains(t, container.Ports, "grpc")
	assert.Equal(t, 9090, container.Ports["grpc"].Port)
}

// TestIntegrationWithMultipleServices tests deployment with multiple interconnected services
func TestIntegrationWithMultipleServices(t *testing.T) {
	helpers := NewTestHelpers(t)
	spec := helpers.CreateTestWiringSpec("MultiServiceIntegration")

	// Create frontend service
	frontend := golang.Service(spec, "frontend")
	frontendHTTP := http.HTTPServer(spec, frontend, "3000")

	// Create backend service
	backend := golang.Service(spec, "backend")
	backendGRPC := grpc.GRPCServer(spec, backend, "9090")

	// Create client connection from frontend to backend
	backendClient := grpc.GRPCClient(spec, frontendHTTP, backendGRPC)

	// Deploy services
	golang.Deploy(spec, frontendHTTP)
	golang.Deploy(spec, backendGRPC)

	// Create Kubernetes deployment with both services
	deployment := NewDeployment(spec, "multi-service-deployment")
	AddContainerToDeployment(spec, deployment, frontendHTTP)
	AddContainerToDeployment(spec, deployment, backendGRPC)
	SetNamespace(spec, deployment, "production")

	// Build IR
	app, err := buildTestApp(t, spec, deployment)
	require.NoError(t, err)

	// Find deployment
	deploymentNode := findNodeOfType[*KubernetesDeployment](app)
	require.NotNil(t, deploymentNode)

	// Verify both services are in the deployment
	assert.Len(t, deploymentNode.Containers, 2)

	// Generate artifacts to test service discovery
	ctx := context.Background()
	mockCtx := newMockBuildContext("/tmp/test")
	err = deploymentNode.GenerateArtifacts(ctx, mockCtx)
	require.NoError(t, err)

	// Check that environment variables for service discovery were added
	frontendContainer := deploymentNode.Containers[0].(*docker.Container)
	assert.Contains(t, frontendContainer.EnvironmentVariables, "BACKEND_ADDR")
}

// TestIntegrationWithClientPool tests integration with client pool
func TestIntegrationWithClientPool(t *testing.T) {
	helpers := NewTestHelpers(t)
	spec := helpers.CreateTestWiringSpec("ClientPoolIntegration")

	// Create services
	service := golang.Service(spec, "pooled-service")
	httpServer := http.HTTPServer(spec, service, "8080")

	// Add client pool
	pooledServer := clientpool.Pool(spec, httpServer, 10)
	golang.Deploy(spec, pooledServer)

	// Create Kubernetes deployment
	deployment := NewDeployment(spec, "pooled-deployment")
	AddContainerToDeployment(spec, deployment, pooledServer)

	// Build IR
	app, err := buildTestApp(t, spec, deployment)
	require.NoError(t, err)

	// Verify deployment was created correctly
	deploymentNode := findNodeOfType[*KubernetesDeployment](app)
	require.NotNil(t, deploymentNode)
	assert.Len(t, deploymentNode.Containers, 1)
}

// TestIntegrationWithNamespaces tests integration with Blueprint namespaces
func TestIntegrationWithNamespaces(t *testing.T) {
	helpers := NewTestHelpers(t)
	spec := helpers.CreateTestWiringSpec("NamespaceIntegration")

	// Create services in different namespaces
	spec.AddNamespace("frontend", "frontend-ns", nil)
	spec.AddNamespace("backend", "backend-ns", nil)

	// Create frontend service
	frontendService := golang.Service(spec, "frontend-service")
	spec.PlaceInNamespace("frontend", frontendService)

	// Create backend service
	backendService := golang.Service(spec, "backend-service")
	spec.PlaceInNamespace("backend", backendService)

	// Deploy services
	golang.Deploy(spec, frontendService)
	golang.Deploy(spec, backendService)

	// Build IR
	_, err := spec.BuildIR()
	require.NoError(t, err)
}

// TestIntegrationComplexMicroserviceDeployment tests a complex microservice deployment scenario
func TestIntegrationComplexMicroserviceDeployment(t *testing.T) {
	helpers := NewTestHelpers(t)
	spec := helpers.CreateTestWiringSpec("ComplexMicroserviceDeployment")

	// Create API Gateway
	apiGateway := golang.Service(spec, "api-gateway")
	apiGatewayHTTP := http.HTTPServer(spec, apiGateway, "80")

	// Create User Service
	userService := golang.Service(spec, "user-service")
	userServiceGRPC := grpc.GRPCServer(spec, userService, "9001")

	// Create Order Service
	orderService := golang.Service(spec, "order-service")
	orderServiceGRPC := grpc.GRPCServer(spec, orderService, "9002")

	// Create Payment Service
	paymentService := golang.Service(spec, "payment-service")
	paymentServiceGRPC := grpc.GRPCServer(spec, paymentService, "9003")

	// Create connections
	userClient := grpc.GRPCClient(spec, apiGatewayHTTP, userServiceGRPC)
	orderClient := grpc.GRPCClient(spec, apiGatewayHTTP, orderServiceGRPC)
	paymentClient := grpc.GRPCClient(spec, orderServiceGRPC, paymentServiceGRPC)

	// Deploy all services
	golang.Deploy(spec, apiGatewayHTTP)
	golang.Deploy(spec, userServiceGRPC)
	golang.Deploy(spec, orderServiceGRPC)
	golang.Deploy(spec, paymentServiceGRPC)

	// Create Kubernetes deployments
	apiDeployment := NewDeployment(spec, "api-gateway-deployment")
	AddContainerToDeployment(spec, apiDeployment, apiGatewayHTTP)
	SetReplicas(spec, apiDeployment, 3)

	servicesDeployment := NewDeployment(spec, "services-deployment")
	AddContainerToDeployment(spec, servicesDeployment, userServiceGRPC)
	AddContainerToDeployment(spec, servicesDeployment, orderServiceGRPC)
	AddContainerToDeployment(spec, servicesDeployment, paymentServiceGRPC)
	SetReplicas(spec, servicesDeployment, 2)

	// Configure cluster for both deployments
	ConfigureCluster(spec, apiDeployment, "https://prod.k8s.example.com", "/etc/kubeconfig", "")
	ConfigureCluster(spec, servicesDeployment, "https://prod.k8s.example.com", "/etc/kubeconfig", "")

	// Build IR
	app, err := buildTestApp(t, spec, apiDeployment, servicesDeployment)
	require.NoError(t, err)

	// Find deployments
	deployments := findAllNodesOfType[*KubernetesDeployment](app)
	assert.Len(t, deployments, 2)

	// Generate artifacts for services deployment
	ctx := context.Background()
	mockCtx := newMockBuildContext("/tmp/test")

	for _, deployment := range deployments {
		if deployment.DeploymentName == "services-deployment" {
			err = deployment.GenerateArtifacts(ctx, mockCtx)
			require.NoError(t, err)

			// Verify all services were included
			assert.Len(t, deployment.Containers, 3)

			// Check that service discovery environment variables were set
			for _, container := range deployment.Containers {
				c := container.(*docker.Container)
				// Each service should know about the others
				if c.Name() == "order-service" {
					assert.Contains(t, c.EnvironmentVariables, "PAYMENT_SERVICE_ADDR")
				}
			}
		}
	}
}

// TestIntegrationWithPointerNodes tests handling of pointer nodes
func TestIntegrationWithPointerNodes(t *testing.T) {
	helpers := NewTestHelpers(t)
	spec := helpers.CreateTestWiringSpec("PointerNodeIntegration")

	// Create a service
	service := golang.Service(spec, "pointed-service")
	httpServer := http.HTTPServer(spec, service, "8080")

	// Create a pointer to the service
	servicePtr := pointer.CreatePointer(spec, httpServer)

	// Deploy the pointed service
	golang.Deploy(spec, servicePtr)

	// Create deployment with the pointer
	deployment := NewDeployment(spec, "pointer-deployment")
	AddContainerToDeployment(spec, deployment, servicePtr)

	// Build IR
	app, err := buildTestApp(t, spec, deployment)
	require.NoError(t, err)

	// Verify the deployment correctly resolved the pointer
	deploymentNode := findNodeOfType[*KubernetesDeployment](app)
	require.NotNil(t, deploymentNode)
	assert.Len(t, deploymentNode.Containers, 1)
}

// TestIntegrationArtifactGeneration tests end-to-end artifact generation
func TestIntegrationArtifactGeneration(t *testing.T) {
	helpers := NewTestHelpers(t)
	spec := helpers.CreateTestWiringSpec("ArtifactGenerationIntegration")

	// Create a complete microservice setup
	deployment, containers := helpers.CreateMicroserviceSetup()

	// Add containers to wiring spec
	for _, container := range containers {
		spec.AddNode(container.Name(), container)
	}
	spec.AddNode(deployment.Name(), deployment)

	// Generate artifacts
	ctx := helpers.GenerateTestArtifacts(deployment, "/tmp/integration-test")

	// Verify all expected files were generated
	helpers.AssertManifestCount(ctx, 7) // 1 deployment + 4 services + 1 configmap + 1 for deployment without services

	// Verify deployment manifest
	deploymentYAML := helpers.AssertFileGenerated(ctx,
		helpers.GetManifestPath("/tmp/integration-test", "microservice-app-deployment.yaml"))
	helpers.ValidateDeploymentManifest(deploymentYAML, "microservice-app", 3)

	// Verify service manifests
	frontendService := helpers.AssertFileGenerated(ctx,
		helpers.GetManifestPath("/tmp/integration-test", "frontend-service.yaml"))
	helpers.ValidateServiceManifest(frontendService, "frontend", 3000)

	apiService := helpers.AssertFileGenerated(ctx,
		helpers.GetManifestPath("/tmp/integration-test", "api-service.yaml"))
	helpers.ValidateServiceManifest(apiService, "api", 8080)

	// Verify scripts
	helpers.AssertScriptContains(ctx,
		helpers.GetScriptPath("/tmp/integration-test", "deploy.sh"),
		[]string{
			"#!/bin/bash",
			"kubectl apply",
			"KUBECONFIG=/etc/kubeconfig",
		})

	// Verify README
	readme := helpers.AssertFileGenerated(ctx,
		helpers.GetScriptPath("/tmp/integration-test", "README.md"))
	assert.Contains(t, readme, "Kubernetes Deployment")
	assert.Contains(t, readme, "microservice-app")
}

// Helper functions for integration tests

func buildTestApp(t *testing.T, spec wiring.WiringSpec, nodes ...ir.IRNode) (*ir.ApplicationNode, error) {
	// Add nodes to spec if not already added
	for _, node := range nodes {
		if !spec.HasNode(node.Name()) {
			spec.AddNode(node.Name(), node)
		}
	}

	app, err := spec.BuildIR()
	if err != nil {
		return nil, err
	}

	return app.(*ir.ApplicationNode), nil
}

func findNodeOfType[T ir.IRNode](app *ir.ApplicationNode) T {
	var result T
	ir.VisitNodes(app, func(node ir.IRNode) error {
		if n, ok := node.(T); ok {
			result = n
			return ir.StopVisiting
		}
		return nil
	})
	return result
}

func findAllNodesOfType[T ir.IRNode](app *ir.ApplicationNode) []T {
	var results []T
	ir.VisitNodes(app, func(node ir.IRNode) error {
		if n, ok := node.(T); ok {
			results = append(results, n)
		}
		return nil
	})
	return results
}

// TestIntegrationWithExistingDockerPlugin tests that our plugin works well with existing docker plugin
func TestIntegrationWithExistingDockerPlugin(t *testing.T) {
	helpers := NewTestHelpers(t)
	spec := helpers.CreateTestWiringSpec("DockerPluginIntegration")

	// Create a service and deploy it with docker
	service := golang.Service(spec, "docker-service")
	dockerContainer := docker.Container(spec, service)
	docker.Deploy(spec, dockerContainer)

	// Create Kubernetes deployment and add the docker container
	deployment := NewDeployment(spec, "k8s-docker-deployment")
	AddContainerToDeployment(spec, deployment, dockerContainer)

	// Build IR
	app, err := buildTestApp(t, spec, deployment)
	require.NoError(t, err)

	// Verify the deployment includes the docker container
	deploymentNode := findNodeOfType[*KubernetesDeployment](app)
	require.NotNil(t, deploymentNode)
	assert.Len(t, deploymentNode.Containers, 1)

	container := deploymentNode.Containers[0].(*docker.Container)
	assert.Equal(t, "docker-service", container.Name())
}

// TestIntegrationErrorHandling tests error handling in integration scenarios
func TestIntegrationErrorHandling(t *testing.T) {
	helpers := NewTestHelpers(t)

	t.Run("InvalidContainerType", func(t *testing.T) {
		spec := helpers.CreateTestWiringSpec("InvalidContainer")

		// Create deployment and try to add non-container node
		deployment := NewDeployment(spec, "bad-deployment")

		// This should handle gracefully when building IR
		spec.AddNode("not-a-container", "string-value")
		AddContainerToDeployment(spec, deployment, "not-a-container")

		_, err := buildTestApp(t, spec, deployment)
		// The error might occur during IR building or artifact generation
		// depending on Blueprint's validation timing
		if err == nil {
			// If no error during build, check artifact generation
			ctx := context.Background()
			mockCtx := newMockBuildContext("/tmp/test")
			err = deployment.GenerateArtifacts(ctx, mockCtx)
		}

		// We expect an error at some point
		assert.Error(t, err)
	})
}
