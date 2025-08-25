package kubernetesgen

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ManifestBuilder builds Kubernetes YAML manifests for deployment
type ManifestBuilder struct {
	deploymentName string
	namespace      string
	replicas       int32
	outputDir      string

	containers     map[string]*ContainerInfo
	envVars        map[string]map[string]string // container -> env var -> value
	ports          map[string][]PortInfo        // container -> ports
	passthroughEnv map[string][]string          // container -> env var names to passthrough
}

// ContainerInfo holds information about a container
type ContainerInfo struct {
	Name    string
	Image   string
	IsLocal bool
	Ports   []PortInfo
	EnvVars map[string]string
}

// PortInfo holds information about a port
type PortInfo struct {
	Name          string
	Port          int
	ContainerPort int
	Protocol      string
}

// NewManifestBuilder creates a new ManifestBuilder
func NewManifestBuilder(deploymentName, namespace string, replicas int32, outputDir string) *ManifestBuilder {
	if namespace == "" {
		namespace = "default"
	}
	if replicas <= 0 {
		replicas = 1
	}
	return &ManifestBuilder{
		deploymentName: deploymentName,
		namespace:      namespace,
		replicas:       replicas,
		outputDir:      outputDir,
		containers:     make(map[string]*ContainerInfo),
		envVars:        make(map[string]map[string]string),
		ports:          make(map[string][]PortInfo),
		passthroughEnv: make(map[string][]string),
	}
}

// AddContainer adds a container to the deployment
func (m *ManifestBuilder) AddContainer(name, image string, isLocal bool) error {
	if _, exists := m.containers[name]; exists {
		return fmt.Errorf("container %s already exists", name)
	}
	m.containers[name] = &ContainerInfo{
		Name:    name,
		Image:   image,
		IsLocal: isLocal,
		EnvVars: make(map[string]string),
	}
	m.envVars[name] = make(map[string]string)
	return nil
}

// AddEnvVar adds an environment variable to a container
func (m *ManifestBuilder) AddEnvVar(containerName, key, value string) error {
	if _, exists := m.containers[containerName]; !exists {
		return fmt.Errorf("container %s does not exist", containerName)
	}
	if m.envVars[containerName] == nil {
		m.envVars[containerName] = make(map[string]string)
	}
	m.envVars[containerName][key] = value
	return nil
}

// PassthroughEnvVar marks an environment variable to be passed through from the host
func (m *ManifestBuilder) PassthroughEnvVar(containerName, key string, optional bool) error {
	if _, exists := m.containers[containerName]; !exists {
		return fmt.Errorf("container %s does not exist", containerName)
	}
	if m.passthroughEnv[containerName] == nil {
		m.passthroughEnv[containerName] = []string{}
	}
	m.passthroughEnv[containerName] = append(m.passthroughEnv[containerName], key)
	return nil
}

// ExposePort exposes a port for a container and creates a service
func (m *ManifestBuilder) ExposePort(containerName string, port int, portName string) error {
	if _, exists := m.containers[containerName]; !exists {
		return fmt.Errorf("container %s does not exist", containerName)
	}
	if m.ports[containerName] == nil {
		m.ports[containerName] = []PortInfo{}
	}
	m.ports[containerName] = append(m.ports[containerName], PortInfo{
		Name:          portName,
		Port:          port,
		ContainerPort: port,
		Protocol:      "TCP",
	})
	return nil
}

// Generate creates all Kubernetes manifest files
func (m *ManifestBuilder) Generate() error {
	// Generate deployment manifest
	if err := m.generateDeployment(); err != nil {
		return fmt.Errorf("failed to generate deployment: %w", err)
	}

	// Generate service manifests
	if err := m.generateServices(); err != nil {
		return fmt.Errorf("failed to generate services: %w", err)
	}

	// Generate ConfigMap if there are environment variables
	if err := m.generateConfigMap(); err != nil {
		return fmt.Errorf("failed to generate configmap: %w", err)
	}

	// Generate apply script
	if err := m.generateApplyScript(); err != nil {
		return fmt.Errorf("failed to generate apply script: %w", err)
	}

	// Generate README
	if err := m.generateReadme(); err != nil {
		return fmt.Errorf("failed to generate README: %w", err)
	}

	return nil
}

// generateDeployment creates the Kubernetes Deployment manifest
func (m *ManifestBuilder) generateDeployment() error {
	var yaml strings.Builder

	yaml.WriteString("apiVersion: apps/v1\n")
	yaml.WriteString("kind: Deployment\n")
	yaml.WriteString("metadata:\n")
	yaml.WriteString(fmt.Sprintf("  name: %s\n", m.deploymentName))
	yaml.WriteString(fmt.Sprintf("  namespace: %s\n", m.namespace))
	yaml.WriteString("spec:\n")
	yaml.WriteString(fmt.Sprintf("  replicas: %d\n", m.replicas))
	yaml.WriteString("  selector:\n")
	yaml.WriteString("    matchLabels:\n")
	yaml.WriteString(fmt.Sprintf("      app: %s\n", m.deploymentName))
	yaml.WriteString("  template:\n")
	yaml.WriteString("    metadata:\n")
	yaml.WriteString("      labels:\n")
	yaml.WriteString(fmt.Sprintf("        app: %s\n", m.deploymentName))
	yaml.WriteString("    spec:\n")
	yaml.WriteString("      containers:\n")

	// Add each container to the deployment
	for _, container := range m.containers {
		yaml.WriteString(fmt.Sprintf("      - name: %s\n", container.Name))

		// Use the image name or local build reference
		if container.IsLocal {
			// For local images, we'll need to specify a registry
			// This should be configured at runtime
			yaml.WriteString(fmt.Sprintf("        image: ${REGISTRY}/%s:latest\n", container.Name))
		} else {
			yaml.WriteString(fmt.Sprintf("        image: %s\n", container.Image))
		}

		// Add ports if any
		if ports, exists := m.ports[container.Name]; exists && len(ports) > 0 {
			yaml.WriteString("        ports:\n")
			for _, port := range ports {
				yaml.WriteString(fmt.Sprintf("        - containerPort: %d\n", port.ContainerPort))
				yaml.WriteString(fmt.Sprintf("          name: %s\n", port.Name))
				yaml.WriteString(fmt.Sprintf("          protocol: %s\n", port.Protocol))
			}
		}

		// Add environment variables
		envVars := m.envVars[container.Name]
		passthroughVars := m.passthroughEnv[container.Name]

		if len(envVars) > 0 || len(passthroughVars) > 0 {
			yaml.WriteString("        env:\n")

			// Add direct environment variables
			for key, value := range envVars {
				yaml.WriteString(fmt.Sprintf("        - name: %s\n", key))
				yaml.WriteString(fmt.Sprintf("          value: \"%s\"\n", value))
			}

			// Add passthrough environment variables from ConfigMap
			for _, key := range passthroughVars {
				yaml.WriteString(fmt.Sprintf("        - name: %s\n", key))
				yaml.WriteString("          valueFrom:\n")
				yaml.WriteString("            configMapKeyRef:\n")
				yaml.WriteString(fmt.Sprintf("              name: %s-config\n", m.deploymentName))
				yaml.WriteString(fmt.Sprintf("              key: %s\n", key))
				yaml.WriteString("              optional: true\n")
			}
		}
	}

	// Write to file
	deploymentFile := filepath.Join(m.outputDir, "deployment.yaml")
	return os.WriteFile(deploymentFile, []byte(yaml.String()), 0644)
}

// generateServices creates Kubernetes Service manifests for containers with exposed ports
func (m *ManifestBuilder) generateServices() error {
	var yaml strings.Builder

	for containerName, ports := range m.ports {
		if len(ports) == 0 {
			continue
		}

		// Each container with ports gets its own service
		yaml.WriteString("---\n")
		yaml.WriteString("apiVersion: v1\n")
		yaml.WriteString("kind: Service\n")
		yaml.WriteString("metadata:\n")
		yaml.WriteString(fmt.Sprintf("  name: %s\n", containerName))
		yaml.WriteString(fmt.Sprintf("  namespace: %s\n", m.namespace))
		yaml.WriteString("spec:\n")
		yaml.WriteString("  selector:\n")
		yaml.WriteString(fmt.Sprintf("    app: %s\n", m.deploymentName))
		yaml.WriteString("  ports:\n")

		for _, port := range ports {
			yaml.WriteString(fmt.Sprintf("  - port: %d\n", port.Port))
			yaml.WriteString(fmt.Sprintf("    targetPort: %d\n", port.ContainerPort))
			yaml.WriteString(fmt.Sprintf("    protocol: %s\n", port.Protocol))
			yaml.WriteString(fmt.Sprintf("    name: %s\n", port.Name))
		}

		yaml.WriteString("  type: ClusterIP\n")
	}

	if yaml.Len() > 0 {
		servicesFile := filepath.Join(m.outputDir, "services.yaml")
		return os.WriteFile(servicesFile, []byte(yaml.String()), 0644)
	}

	return nil
}

// generateConfigMap creates a ConfigMap for environment variables
func (m *ManifestBuilder) generateConfigMap() error {
	// Check if we need a ConfigMap
	hasPassthrough := false
	for _, vars := range m.passthroughEnv {
		if len(vars) > 0 {
			hasPassthrough = true
			break
		}
	}

	if !hasPassthrough {
		return nil
	}

	var yaml strings.Builder

	yaml.WriteString("apiVersion: v1\n")
	yaml.WriteString("kind: ConfigMap\n")
	yaml.WriteString("metadata:\n")
	yaml.WriteString(fmt.Sprintf("  name: %s-config\n", m.deploymentName))
	yaml.WriteString(fmt.Sprintf("  namespace: %s\n", m.namespace))
	yaml.WriteString("data:\n")

	// Collect all unique passthrough variables
	uniqueVars := make(map[string]bool)
	for _, vars := range m.passthroughEnv {
		for _, v := range vars {
			uniqueVars[v] = true
		}
	}

	// Add placeholders for each variable
	for varName := range uniqueVars {
		yaml.WriteString(fmt.Sprintf("  %s: \"${%s}\"\n", varName, varName))
	}

	configMapFile := filepath.Join(m.outputDir, "configmap.yaml")
	return os.WriteFile(configMapFile, []byte(yaml.String()), 0644)
}

// generateApplyScript creates a script to apply all manifests
func (m *ManifestBuilder) generateApplyScript() error {
	var script strings.Builder

	script.WriteString("#!/bin/bash\n\n")
	script.WriteString("# Script to deploy the Kubernetes manifests\n\n")

	script.WriteString("# Check if kubectl is installed\n")
	script.WriteString("if ! command -v kubectl &> /dev/null; then\n")
	script.WriteString("    echo \"kubectl is not installed. Please install kubectl first.\"\n")
	script.WriteString("    exit 1\n")
	script.WriteString("fi\n\n")

	script.WriteString("# Create namespace if it doesn't exist\n")
	script.WriteString(fmt.Sprintf("kubectl create namespace %s --dry-run=client -o yaml | kubectl apply -f -\n\n", m.namespace))

	script.WriteString("# Apply manifests\n")
	script.WriteString("echo \"Applying Kubernetes manifests...\"\n\n")

	// Check if files exist and apply them
	script.WriteString("if [ -f configmap.yaml ]; then\n")
	script.WriteString("    echo \"Applying ConfigMap...\"\n")
	script.WriteString("    kubectl apply -f configmap.yaml\n")
	script.WriteString("fi\n\n")

	script.WriteString("if [ -f services.yaml ]; then\n")
	script.WriteString("    echo \"Applying Services...\"\n")
	script.WriteString("    kubectl apply -f services.yaml\n")
	script.WriteString("fi\n\n")

	script.WriteString("echo \"Applying Deployment...\"\n")
	script.WriteString("kubectl apply -f deployment.yaml\n\n")

	script.WriteString("echo \"Deployment complete!\"\n")
	script.WriteString(fmt.Sprintf("echo \"To check status: kubectl get pods -n %s\"\n", m.namespace))
	script.WriteString(fmt.Sprintf("echo \"To view logs: kubectl logs -n %s -l app=%s\"\n", m.namespace, m.deploymentName))

	scriptFile := filepath.Join(m.outputDir, "apply.sh")
	if err := os.WriteFile(scriptFile, []byte(script.String()), 0755); err != nil {
		return err
	}

	// Also create a Windows batch file
	var batch strings.Builder
	batch.WriteString("@echo off\n\n")
	batch.WriteString("REM Script to deploy the Kubernetes manifests\n\n")
	batch.WriteString("REM Check if kubectl is installed\n")
	batch.WriteString("where kubectl >nul 2>nul\n")
	batch.WriteString("if %ERRORLEVEL% NEQ 0 (\n")
	batch.WriteString("    echo kubectl is not installed. Please install kubectl first.\n")
	batch.WriteString("    exit /b 1\n")
	batch.WriteString(")\n\n")

	batch.WriteString("REM Create namespace if it doesn't exist\n")
	batch.WriteString(fmt.Sprintf("kubectl create namespace %s --dry-run=client -o yaml | kubectl apply -f -\n\n", m.namespace))

	batch.WriteString("REM Apply manifests\n")
	batch.WriteString("echo Applying Kubernetes manifests...\n\n")

	batch.WriteString("if exist configmap.yaml (\n")
	batch.WriteString("    echo Applying ConfigMap...\n")
	batch.WriteString("    kubectl apply -f configmap.yaml\n")
	batch.WriteString(")\n\n")

	batch.WriteString("if exist services.yaml (\n")
	batch.WriteString("    echo Applying Services...\n")
	batch.WriteString("    kubectl apply -f services.yaml\n")
	batch.WriteString(")\n\n")

	batch.WriteString("echo Applying Deployment...\n")
	batch.WriteString("kubectl apply -f deployment.yaml\n\n")

	batch.WriteString("echo Deployment complete!\n")
	batch.WriteString(fmt.Sprintf("echo To check status: kubectl get pods -n %s\n", m.namespace))
	batch.WriteString(fmt.Sprintf("echo To view logs: kubectl logs -n %s -l app=%s\n", m.namespace, m.deploymentName))

	batchFile := filepath.Join(m.outputDir, "apply.bat")
	return os.WriteFile(batchFile, []byte(batch.String()), 0644)
}

// generateReadme creates a README file with deployment instructions
func (m *ManifestBuilder) generateReadme() error {
	var readme strings.Builder

	readme.WriteString(fmt.Sprintf("# Kubernetes Deployment: %s\n\n", m.deploymentName))
	readme.WriteString("This directory contains Kubernetes manifests for deploying the application.\n\n")

	readme.WriteString("## Prerequisites\n\n")
	readme.WriteString("1. Access to a Kubernetes cluster\n")
	readme.WriteString("2. `kubectl` configured to connect to your cluster\n")
	readme.WriteString("3. Required environment variables set (see below)\n\n")

	readme.WriteString("## Files\n\n")
	readme.WriteString("- `deployment.yaml` - Kubernetes Deployment resource\n")
	readme.WriteString("- `services.yaml` - Kubernetes Service resources for networking\n")
	readme.WriteString("- `configmap.yaml` - ConfigMap for environment variables\n")
	readme.WriteString("- `apply.sh` / `apply.bat` - Scripts to deploy all resources\n\n")

	readme.WriteString("## Deployment\n\n")
	readme.WriteString("### Using the provided script:\n\n")
	readme.WriteString("```bash\n")
	readme.WriteString("# Linux/Mac\n")
	readme.WriteString("./apply.sh\n\n")
	readme.WriteString("# Windows\n")
	readme.WriteString("apply.bat\n")
	readme.WriteString("```\n\n")

	readme.WriteString("### Manual deployment:\n\n")
	readme.WriteString("```bash\n")
	readme.WriteString(fmt.Sprintf("# Create namespace\n"))
	readme.WriteString(fmt.Sprintf("kubectl create namespace %s\n\n", m.namespace))
	readme.WriteString("# Apply manifests\n")
	readme.WriteString("kubectl apply -f configmap.yaml\n")
	readme.WriteString("kubectl apply -f services.yaml\n")
	readme.WriteString("kubectl apply -f deployment.yaml\n")
	readme.WriteString("```\n\n")

	readme.WriteString("## Configuration\n\n")
	readme.WriteString(fmt.Sprintf("- **Namespace:** %s\n", m.namespace))
	readme.WriteString(fmt.Sprintf("- **Replicas:** %d\n", m.replicas))
	readme.WriteString(fmt.Sprintf("- **Deployment Name:** %s\n\n", m.deploymentName))

	// List required environment variables
	uniqueVars := make(map[string]bool)
	for _, vars := range m.passthroughEnv {
		for _, v := range vars {
			uniqueVars[v] = true
		}
	}

	if len(uniqueVars) > 0 {
		readme.WriteString("## Required Environment Variables\n\n")
		readme.WriteString("Set these environment variables before deploying:\n\n")
		for varName := range uniqueVars {
			readme.WriteString(fmt.Sprintf("- `%s`\n", varName))
		}
		readme.WriteString("\n")
	}

	readme.WriteString("## Monitoring\n\n")
	readme.WriteString("```bash\n")
	readme.WriteString(fmt.Sprintf("# Check pod status\n"))
	readme.WriteString(fmt.Sprintf("kubectl get pods -n %s\n\n", m.namespace))
	readme.WriteString(fmt.Sprintf("# View logs\n"))
	readme.WriteString(fmt.Sprintf("kubectl logs -n %s -l app=%s\n\n", m.namespace, m.deploymentName))
	readme.WriteString(fmt.Sprintf("# Describe deployment\n"))
	readme.WriteString(fmt.Sprintf("kubectl describe deployment -n %s %s\n", m.namespace, m.deploymentName))
	readme.WriteString("```\n\n")

	readme.WriteString("## Cleanup\n\n")
	readme.WriteString("```bash\n")
	readme.WriteString(fmt.Sprintf("kubectl delete -f deployment.yaml\n"))
	readme.WriteString(fmt.Sprintf("kubectl delete -f services.yaml\n"))
	readme.WriteString(fmt.Sprintf("kubectl delete -f configmap.yaml\n"))
	readme.WriteString("```\n")

	readmeFile := filepath.Join(m.outputDir, "README.md")
	return os.WriteFile(readmeFile, []byte(readme.String()), 0644)
}
