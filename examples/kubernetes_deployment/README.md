# Kubernetes Deployment Example

This example demonstrates how to use the Blueprint Kubernetes plugin to deploy a multi-service application to a Kubernetes cluster.

## Overview

The example includes three different wiring specifications:

1. **WireSpec** - A complete e-commerce application with multiple services
2. **WireSpecWithNamespaces** - The same application organized using Blueprint namespaces
3. **SimpleExample** - A minimal "Hello World" service deployment

## Architecture

### E-Commerce Application (WireSpec)

The example e-commerce application consists of:

**Frontend Services:**
- `frontend` - Web UI served on port 3000
- `api-gateway` - API Gateway on port 8080 routing requests to backend services

**Backend Services:**
- `user-service` - User authentication and profiles (port 8081)
- `product-service` - Product catalog management (port 8082)
- `order-service` - Order processing (port 8083)
- `cart-service` - Shopping cart management (port 8084)
- `payment-service` - Payment processing with Stripe integration (port 8085)
- `notification-service` - Email and notification handling (port 8086)

**Data Stores:**
- `mongodb` - Document database for products and orders
- `redis-cache` - In-memory cache for sessions and data

**Observability:**
- `jaeger` - Distributed tracing for all services
- Health checking endpoints for all services

## Prerequisites

1. A running Kubernetes cluster
2. `kubectl` configured to access your cluster
3. Blueprint framework installed
4. Go 1.19 or later

## Building the Application

### Option 1: Simple Example

```bash
# Compile the simple example
blueprint compile \
  -w examples/kubernetes_deployment/wiring/main.go \
  -f SimpleExample \
  -o output/simple
```

### Option 2: Full E-Commerce Application

```bash
# Compile the full application
blueprint compile \
  -w examples/kubernetes_deployment/wiring/main.go \
  -f WireSpec \
  -o output/ecommerce
```

### Option 3: Namespaced Organization

```bash
# Compile with namespace organization
blueprint compile \
  -w examples/kubernetes_deployment/wiring/main.go \
  -f WireSpecWithNamespaces \
  -o output/ecommerce-namespaced
```

## Deployment

After compilation, navigate to the output directory and deploy:

### 1. Review Generated Manifests

```bash
cd output/ecommerce/kubernetes/
ls -la

# You should see:
# - deployment.yaml      # Kubernetes Deployment
# - services.yaml       # Service definitions
# - configmap.yaml      # Environment variables
# - namespace.yaml      # Namespace definition
# - deploy.sh          # Deployment script
# - README.md          # Deployment instructions
```

### 2. Configure Cluster Access

```bash
# Ensure kubectl is configured
kubectl config current-context

# Verify cluster access
kubectl cluster-info
```

### 3. Deploy to Kubernetes

```bash
# Using the generated script
./deploy.sh

# Or manually
kubectl apply -f namespace.yaml
kubectl apply -f configmap.yaml
kubectl apply -f services.yaml
kubectl apply -f deployment.yaml
```

### 4. Verify Deployment

```bash
# Check deployment status
kubectl get deployments -n ecommerce

# Check all pods are running
kubectl get pods -n ecommerce

# Check services
kubectl get services -n ecommerce
```

## Accessing the Application

### Port Forwarding (Development)

```bash
# Forward frontend service
kubectl port-forward -n ecommerce service/frontend-service 3000:3000

# Forward API gateway
kubectl port-forward -n ecommerce service/api-gateway-service 8080:8080

# Access Jaeger UI
kubectl port-forward -n ecommerce service/jaeger-service 16686:16686
```

Then access:
- Frontend: http://localhost:3000
- API Gateway: http://localhost:8080
- Jaeger UI: http://localhost:16686

### Production Access

For production, you would typically:

1. Configure an Ingress controller
2. Set up LoadBalancer services
3. Configure DNS entries

Example Ingress configuration:

```yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: ecommerce-ingress
  namespace: ecommerce
spec:
  rules:
  - host: app.example.com
    http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: frontend-service
            port:
              number: 3000
  - host: api.example.com
    http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: api-gateway-service
            port:
              number: 8080
```

## Configuration

### Environment Variables

The application uses environment variables for configuration. These are stored in a ConfigMap and can be modified:

```bash
# Edit the ConfigMap
kubectl edit configmap ecommerce-app-config -n ecommerce
```

Key environment variables:
- `STRIPE_API_KEY` - Payment processor API key
- `SMTP_HOST`, `SMTP_PORT` - Email server configuration
- `NODE_ENV` - Application environment (development/production)

### Scaling

Adjust the number of replicas:

```bash
# Scale the deployment
kubectl scale deployment ecommerce-app -n ecommerce --replicas=5

# Or edit the deployment
kubectl edit deployment ecommerce-app -n ecommerce
```

## Monitoring

### View Logs

```bash
# View logs for a specific pod
kubectl logs -n ecommerce <pod-name>

# Follow logs
kubectl logs -f -n ecommerce <pod-name>

# View logs for all pods with a label
kubectl logs -n ecommerce -l app=ecommerce-app
```

### Health Checks

All services expose health check endpoints:

```bash
# Check health of a service
kubectl exec -n ecommerce <pod-name> -- curl http://localhost:8080/health
```

### Distributed Tracing

Access Jaeger UI to view traces:

```bash
kubectl port-forward -n ecommerce service/jaeger-service 16686:16686
```

Open http://localhost:16686 in your browser.

## Troubleshooting

### Pods Not Starting

```bash
# Describe the pod for details
kubectl describe pod <pod-name> -n ecommerce

# Check events
kubectl get events -n ecommerce --sort-by='.lastTimestamp'
```

### Service Discovery Issues

```bash
# Test DNS resolution from within a pod
kubectl exec -n ecommerce <pod-name> -- nslookup user-service

# Test service connectivity
kubectl exec -n ecommerce <pod-name> -- curl http://user-service:8081/health
```

### Database Connection Issues

```bash
# Check MongoDB pod
kubectl logs -n ecommerce -l app=mongodb

# Check Redis pod
kubectl logs -n ecommerce -l app=redis-cache
```

## Cleanup

To remove the deployed application:

```bash
# Delete all resources in the namespace
kubectl delete namespace ecommerce

# Or delete individual resources
kubectl delete -f deployment.yaml
kubectl delete -f services.yaml
kubectl delete -f configmap.yaml
kubectl delete -f namespace.yaml
```

## Customization

### Adding New Services

1. Edit the wiring specification to add new service definitions
2. Configure service dependencies and connections
3. Recompile and redeploy

### Modifying Deployment Parameters

Edit the wiring specification to change:
- Namespace: `kubernetes.SetNamespace(deployment, "custom-namespace")`
- Replicas: `kubernetes.SetReplicas(deployment, 5)`
- Cluster config: `kubernetes.ConfigureCluster(deployment, endpoint, kubeconfig, token)`

## Advanced Features

### Using with CI/CD

The generated artifacts can be integrated into CI/CD pipelines:

```yaml
# Example GitHub Actions workflow
name: Deploy to Kubernetes
on:
  push:
    branches: [main]
jobs:
  deploy:
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v2
    - name: Compile Blueprint
      run: |
        blueprint compile \
          -w examples/kubernetes_deployment/wiring/main.go \
          -f WireSpec \
          -o output
    - name: Deploy to Kubernetes
      run: |
        kubectl apply -f output/kubernetes/namespace.yaml
        kubectl apply -f output/kubernetes/configmap.yaml
        kubectl apply -f output/kubernetes/services.yaml
        kubectl apply -f output/kubernetes/deployment.yaml
```

### Multi-Environment Deployments

Use different wiring functions for different environments:

```go
// development.go
func WireSpecDev(spec wiring.WiringSpec) {
    // Development configuration
    kubernetes.SetReplicas(deployment, 1)
    kubernetes.SetNamespace(deployment, "dev")
}

// production.go
func WireSpecProd(spec wiring.WiringSpec) {
    // Production configuration
    kubernetes.SetReplicas(deployment, 5)
    kubernetes.SetNamespace(deployment, "prod")
}
```

## Related Examples

- `examples/sockshop` - Another microservices example
- `examples/dsb_hotel` - Hotel reservation system
- `examples/train_ticket` - Train ticket booking system

## Support

For issues or questions about this example:
1. Check the main Blueprint documentation
2. Review the Kubernetes plugin README at `plugins/kubernetes/README.md`
3. Open an issue on the Blueprint GitHub repository
