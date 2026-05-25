# Kubernetes Plugin

The Kubernetes plugin enables deployment of Blueprint applications to pre-existing Kubernetes clusters. It generates Kubernetes manifests (Deployments, Services, ConfigMaps) and deployment scripts from Blueprint IR nodes.

## Overview

This plugin provides functionality to:
- Deploy containers to Kubernetes clusters
- Generate Kubernetes manifests (Deployments, Services, ConfigMaps)
- Handle service discovery through Kubernetes Services
- Manage environment variables via ConfigMaps
- Support runtime cluster configuration
- Generate cross-platform deployment scripts

## Features

### Container Deployment
- Transforms docker.Container nodes into Kubernetes Deployments
- Supports custom replica counts
- Handles container images and resource requirements

### Service Discovery
- Automatically creates Kubernetes Services for containers with exposed ports
- Enables inter-service communication through cluster DNS
- Maps container ports to service ports

### Configuration Management
- Environment variables stored in ConfigMaps
- Runtime cluster configuration (endpoint, credentials)
- Namespace isolation support

### Deployment Artifacts
- Kubernetes YAML manifests
- Bash scripts for Linux/macOS deployment
- Batch scripts for Windows deployment
- Comprehensive deployment documentation

## Usage

### Basic Deployment

```go
package main

import (
    "github.com/blueprint-uservices/blueprint/plugins/kubernetes"
    "github.com/blueprint-uservices/blueprint/plugins/golang"
)

func main() {
    // Define your service
    backend := golang.Service("backend")
    
    // Create a Kubernetes deployment
    deployment := kubernetes.NewDeployment("my-app")
    
    // Add the service container to the deployment
    kubernetes.AddContainerToDeployment(deployment, backend)
    
    // Configure cluster (runtime configuration)
    kubernetes.ConfigureCluster(deployment, 
        "https://k8s.example.com:6443",  // API endpoint
        "/path/to/kubeconfig",            // Kubeconfig path
        "my-auth-token")                  // Optional auth token
    
    // Set namespace
    kubernetes.SetNamespace(deployment, "production")
    
    // Set replica count
    kubernetes.SetReplicas(deployment, 3)
}
```

### Multi-Service Application

```go
func DeployMicroservices(spec wiring.WiringSpec) {
    // Define services
    frontend := golang.Service("frontend")
    backend := golang.Service("backend")
    database := golang.Service("database")
    
    // Create deployment
    deployment := kubernetes.NewDeployment("microservices")
    
    // Add all services
    kubernetes.AddContainerToDeployment(deployment, frontend)
    kubernetes.AddContainerToDeployment(deployment, backend)
    kubernetes.AddContainerToDeployment(deployment, database)
    
    // Configure deployment
    kubernetes.SetNamespace(deployment, "microservices")
    kubernetes.SetReplicas(deployment, 2)
    
    // Runtime cluster config will be provided during deployment
    kubernetes.ConfigureCluster(deployment, "", "", "")
}
```

### With Environment Variables

```go
func DeployWithConfig(spec wiring.WiringSpec) {
    service := golang.Service("my-service")
    
    // Set environment variables on the service
    service.SetEnvironmentVariable("DATABASE_URL", "postgres://...")
    service.SetEnvironmentVariable("API_KEY", "secret-key")
    
    // Deploy to Kubernetes
    deployment := kubernetes.NewDeployment("configured-app")
    kubernetes.AddContainerToDeployment(deployment, service)
    
    // Environment variables will be stored in a ConfigMap
}
```

## Wiring Specification

The plugin can be used in wiring specifications with namespace handlers:

```go
func WireApplication(spec wiring.WiringSpec) {
    // Define a Kubernetes deployment namespace
    k8s := spec.Define("kubernetes", "my-app", func(ns wiring.Namespace) {
        // The namespace handler accepts docker.Container nodes
        // and deploys them to Kubernetes
    })
    
    // Place services in the Kubernetes namespace
    frontend := golang.Service("frontend")
    backend := golang.Service("backend")
    
    k8s.Place(frontend)
    k8s.Place(backend)
    
    // Configure the deployment
    kubernetes.SetNamespace(k8s.IR(), "production")
    kubernetes.SetReplicas(k8s.IR(), 3)
}
```

## Generated Artifacts

When compiled, the plugin generates the following artifacts in the output directory:

### Kubernetes Manifests (`kubernetes/`)
- `deployment.yaml` - Kubernetes Deployment resource
- `services.yaml` - Kubernetes Service resources for exposed ports
- `configmap.yaml` - ConfigMap for environment variables
- `namespace.yaml` - Namespace definition (if specified)

### Deployment Scripts
- `deploy.sh` - Bash script for Linux/macOS deployment
- `deploy.bat` - Batch script for Windows deployment

### Documentation
- `README.md` - Deployment instructions and service information

## Deployment Process

### 1. Generate Artifacts
```bash
# Compile your wiring specification
blueprint compile -w wiring/main.go -o output/
```

### 2. Configure Cluster Access
```bash
# Set your kubeconfig
export KUBECONFIG=/path/to/kubeconfig

# Or use kubectl config
kubectl config use-context my-cluster
```

### 3. Deploy to Kubernetes
```bash
# Navigate to output directory
cd output/kubernetes/

# Deploy using the generated script
./deploy.sh

# Or apply manually
kubectl apply -f namespace.yaml
kubectl apply -f configmap.yaml
kubectl apply -f services.yaml
kubectl apply -f deployment.yaml
```

### 4. Verify Deployment
```bash
# Check deployment status
kubectl get deployments -n <namespace>

# Check pods
kubectl get pods -n <namespace>

# Check services
kubectl get services -n <namespace>
```

## Configuration Options

### Namespace Configuration
- `SetNamespace(deployment, namespace)` - Set Kubernetes namespace
- Default: "default"

### Replica Configuration
- `SetReplicas(deployment, count)` - Set number of replicas
- Default: 1

### Cluster Configuration
- `ConfigureCluster(deployment, endpoint, kubeconfig, token)`
- Can be configured at compile time or runtime
- Supports multiple authentication methods

## Service Discovery

Services within the same deployment can discover each other using Kubernetes DNS:

- Service URL format: `http://<service-name>:<port>`
- Full DNS name: `<service-name>.<namespace>.svc.cluster.local`

Example:
```go
// In your application code
backendURL := "http://backend-service:8080"
```

## Environment Variables

Environment variables are automatically collected from container nodes and stored in a ConfigMap:

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: my-app-config
data:
  DATABASE_URL: "postgres://..."
  API_KEY: "secret-key"
```

## Limitations

- Requires a pre-existing Kubernetes cluster
- Does not provision cloud resources
- Persistent volumes must be manually configured
- Secrets should be managed separately for production use

## Integration with Other Plugins

The Kubernetes plugin works seamlessly with other Blueprint plugins:

- **Docker Plugin**: Accepts docker.Container nodes
- **Golang Plugin**: Deploy Go services
- **HTTP/gRPC Plugins**: Expose services with proper ports
- **Jaeger/Zipkin**: Deploy tracing infrastructure

## Examples

See the `examples/` directory for complete examples:
- Simple web application deployment
- Microservices with service mesh
- Database-backed application
- Distributed tracing setup

## Troubleshooting

### Common Issues

1. **Authentication Failed**
   - Verify kubeconfig path
   - Check cluster endpoint URL
   - Ensure credentials are valid

2. **Services Not Connecting**
   - Check service names match
   - Verify namespace configuration
   - Ensure ports are correctly exposed

3. **Pods Not Starting**
   - Check image availability
   - Review resource limits
   - Examine pod logs: `kubectl logs <pod-name> -n <namespace>`

### Debug Commands

```bash
# Describe deployment
kubectl describe deployment <name> -n <namespace>

# Get events
kubectl get events -n <namespace>

# View logs
kubectl logs -f <pod-name> -n <namespace>

# Port forward for testing
kubectl port-forward <pod-name> 8080:8080 -n <namespace>
```

## Contributing

Contributions are welcome! Please see the main Blueprint contributing guidelines.

## License

See the Blueprint project license.
