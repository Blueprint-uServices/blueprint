// Package main provides an example wiring specification for deploying
// a multi-service application to Kubernetes using the Blueprint Kubernetes plugin.
package main

import (
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/wiring"
	"github.com/blueprint-uservices/blueprint/plugins/golang"
	"github.com/blueprint-uservices/blueprint/plugins/grpc"
	"github.com/blueprint-uservices/blueprint/plugins/healthchecker"
	"github.com/blueprint-uservices/blueprint/plugins/http"
	"github.com/blueprint-uservices/blueprint/plugins/jaeger"
	"github.com/blueprint-uservices/blueprint/plugins/kubernetes"
	"github.com/blueprint-uservices/blueprint/plugins/mongodb"
	"github.com/blueprint-uservices/blueprint/plugins/redis"
)

// WireSpec defines the wiring specification for deploying a sample
// e-commerce application to Kubernetes.
func WireSpec(spec wiring.WiringSpec) {
	// Define the services

	// Frontend service - serves the web UI
	frontend := golang.Service("frontend",
		golang.WithServicePort(3000),
		golang.WithEnvironment(map[string]string{
			"NODE_ENV": "production",
			"API_URL":  "http://api-gateway:8080",
		}),
	)

	// API Gateway - routes requests to backend services
	apiGateway := golang.Service("api-gateway",
		golang.WithServicePort(8080),
	)

	// User Service - handles user authentication and profiles
	userService := golang.Service("user-service",
		golang.WithServicePort(8081),
	)

	// Product Service - manages product catalog
	productService := golang.Service("product-service",
		golang.WithServicePort(8082),
	)

	// Order Service - processes orders
	orderService := golang.Service("order-service",
		golang.WithServicePort(8083),
	)

	// Cart Service - manages shopping carts
	cartService := golang.Service("cart-service",
		golang.WithServicePort(8084),
	)

	// Payment Service - handles payment processing
	paymentService := golang.Service("payment-service",
		golang.WithServicePort(8085),
		golang.WithEnvironment(map[string]string{
			"STRIPE_API_KEY": "${STRIPE_API_KEY}",
			"PAYMENT_MODE":   "sandbox",
		}),
	)

	// Notification Service - sends emails and notifications
	notificationService := golang.Service("notification-service",
		golang.WithServicePort(8086),
		golang.WithEnvironment(map[string]string{
			"SMTP_HOST":  "smtp.gmail.com",
			"SMTP_PORT":  "587",
			"EMAIL_FROM": "noreply@example.com",
		}),
	)

	// Add databases

	// MongoDB for product catalog and orders
	mongoDb := mongodb.Container("mongodb")
	mongodb.Connect(productService, mongoDb, "products_db")
	mongodb.Connect(orderService, mongoDb, "orders_db")

	// Redis for session storage and caching
	redisCache := redis.Container("redis-cache")
	redis.Connect(apiGateway, redisCache)
	redis.Connect(cartService, redisCache)
	redis.Connect(userService, redisCache)

	// Add HTTP endpoints
	http.Expose(frontend, "frontend")
	http.Expose(apiGateway, "api")

	// Add gRPC communication between services
	grpc.Deploy(userService)
	grpc.Deploy(productService)
	grpc.Deploy(orderService)
	grpc.Deploy(cartService)
	grpc.Deploy(paymentService)
	grpc.Deploy(notificationService)

	// Connect API Gateway to backend services
	userClient := grpc.Client(apiGateway, userService)
	productClient := grpc.Client(apiGateway, productService)
	orderClient := grpc.Client(apiGateway, orderService)
	cartClient := grpc.Client(apiGateway, cartService)
	paymentClient := grpc.Client(apiGateway, paymentService)
	notificationClient := grpc.Client(apiGateway, notificationService)

	// Add health checking
	healthchecker.AddHealthCheck(frontend)
	healthchecker.AddHealthCheck(apiGateway)
	healthchecker.AddHealthCheck(userService)
	healthchecker.AddHealthCheck(productService)
	healthchecker.AddHealthCheck(orderService)
	healthchecker.AddHealthCheck(cartService)
	healthchecker.AddHealthCheck(paymentService)
	healthchecker.AddHealthCheck(notificationService)

	// Add distributed tracing with Jaeger
	jaegerCollector := jaeger.Collector("jaeger")
	jaeger.Instrument(frontend, jaegerCollector)
	jaeger.Instrument(apiGateway, jaegerCollector)
	jaeger.Instrument(userService, jaegerCollector)
	jaeger.Instrument(productService, jaegerCollector)
	jaeger.Instrument(orderService, jaegerCollector)
	jaeger.Instrument(cartService, jaegerCollector)
	jaeger.Instrument(paymentService, jaegerCollector)
	jaeger.Instrument(notificationService, jaegerCollector)

	// Create Kubernetes deployment
	k8sDeployment := kubernetes.NewDeployment("ecommerce-app")

	// Add all services to the deployment
	kubernetes.AddContainerToDeployment(k8sDeployment, frontend)
	kubernetes.AddContainerToDeployment(k8sDeployment, apiGateway)
	kubernetes.AddContainerToDeployment(k8sDeployment, userService)
	kubernetes.AddContainerToDeployment(k8sDeployment, productService)
	kubernetes.AddContainerToDeployment(k8sDeployment, orderService)
	kubernetes.AddContainerToDeployment(k8sDeployment, cartService)
	kubernetes.AddContainerToDeployment(k8sDeployment, paymentService)
	kubernetes.AddContainerToDeployment(k8sDeployment, notificationService)
	kubernetes.AddContainerToDeployment(k8sDeployment, mongoDb)
	kubernetes.AddContainerToDeployment(k8sDeployment, redisCache)
	kubernetes.AddContainerToDeployment(k8sDeployment, jaegerCollector)

	// Configure the Kubernetes deployment
	kubernetes.SetNamespace(k8sDeployment, "ecommerce")
	kubernetes.SetReplicas(k8sDeployment, 2) // 2 replicas for each service

	// Cluster configuration can be provided at runtime
	// For now, we'll leave it empty to be configured during deployment
	kubernetes.ConfigureCluster(k8sDeployment, "", "", "")

	// Add the deployment to the wiring spec
	spec.AddNode(k8sDeployment)
}

// WireSpecWithNamespaces demonstrates using namespace handlers
// to organize the deployment.
func WireSpecWithNamespaces(spec wiring.WiringSpec) {
	// Define a Kubernetes namespace for the application
	k8s := spec.Define("kubernetes", "ecommerce-app", func(ns wiring.Namespace) {
		// The namespace will handle docker.Container nodes
		// and deploy them to Kubernetes
	})

	// Define services in separate namespaces for organization

	// Frontend namespace
	web := spec.Define("web", "frontend", func(ns wiring.Namespace) {
		frontend := golang.Service("frontend",
			golang.WithServicePort(3000),
		)
		http.Expose(frontend, "frontend")
		healthchecker.AddHealthCheck(frontend)
		ns.Export(frontend, "frontend")
	})

	// API namespace
	api := spec.Define("api", "gateway", func(ns wiring.Namespace) {
		gateway := golang.Service("api-gateway",
			golang.WithServicePort(8080),
		)
		http.Expose(gateway, "api")
		healthchecker.AddHealthCheck(gateway)
		ns.Export(gateway, "gateway")
	})

	// Services namespace
	services := spec.Define("services", "backend", func(ns wiring.Namespace) {
		// Define all backend services
		user := golang.Service("user-service", golang.WithServicePort(8081))
		product := golang.Service("product-service", golang.WithServicePort(8082))
		order := golang.Service("order-service", golang.WithServicePort(8083))
		cart := golang.Service("cart-service", golang.WithServicePort(8084))
		payment := golang.Service("payment-service", golang.WithServicePort(8085))
		notification := golang.Service("notification-service", golang.WithServicePort(8086))

		// Deploy as gRPC services
		grpc.Deploy(user)
		grpc.Deploy(product)
		grpc.Deploy(order)
		grpc.Deploy(cart)
		grpc.Deploy(payment)
		grpc.Deploy(notification)

		// Add health checks
		healthchecker.AddHealthCheck(user)
		healthchecker.AddHealthCheck(product)
		healthchecker.AddHealthCheck(order)
		healthchecker.AddHealthCheck(cart)
		healthchecker.AddHealthCheck(payment)
		healthchecker.AddHealthCheck(notification)

		// Export services
		ns.Export(user, "user")
		ns.Export(product, "product")
		ns.Export(order, "order")
		ns.Export(cart, "cart")
		ns.Export(payment, "payment")
		ns.Export(notification, "notification")
	})

	// Data namespace
	data := spec.Define("data", "storage", func(ns wiring.Namespace) {
		mongo := mongodb.Container("mongodb")
		redis := redis.Container("redis-cache")

		ns.Export(mongo, "mongodb")
		ns.Export(redis, "redis")
	})

	// Observability namespace
	observability := spec.Define("observability", "monitoring", func(ns wiring.Namespace) {
		jaeger := jaeger.Collector("jaeger")
		ns.Export(jaeger, "jaeger")
	})

	// Import services into the API gateway namespace
	api.Import(services, "user", "userService")
	api.Import(services, "product", "productService")
	api.Import(services, "order", "orderService")
	api.Import(services, "cart", "cartService")
	api.Import(services, "payment", "paymentService")
	api.Import(services, "notification", "notificationService")

	// Import data services
	services.Import(data, "mongodb", "database")
	services.Import(data, "redis", "cache")
	api.Import(data, "redis", "sessionStore")

	// Import observability
	web.Import(observability, "jaeger", "tracer")
	api.Import(observability, "jaeger", "tracer")
	services.Import(observability, "jaeger", "tracer")

	// Place all components in the Kubernetes namespace
	k8s.Place(web.IR())
	k8s.Place(api.IR())
	k8s.Place(services.IR())
	k8s.Place(data.IR())
	k8s.Place(observability.IR())

	// Configure Kubernetes deployment
	deployment := k8s.IR()
	kubernetes.SetNamespace(deployment, "ecommerce")
	kubernetes.SetReplicas(deployment, 3)

	// Cluster config to be provided at runtime
	kubernetes.ConfigureCluster(deployment, "", "", "")
}

// SimpleExample provides a minimal example of deploying a single service
func SimpleExample(spec wiring.WiringSpec) {
	// Create a simple HTTP service
	helloService := golang.Service("hello-world",
		golang.WithServicePort(8080),
		golang.WithEnvironment(map[string]string{
			"MESSAGE": "Hello from Kubernetes!",
		}),
	)

	// Expose HTTP endpoint
	http.Expose(helloService, "hello")

	// Add health check
	healthchecker.AddHealthCheck(helloService)

	// Create Kubernetes deployment
	deployment := kubernetes.NewDeployment("hello-app")
	kubernetes.AddContainerToDeployment(deployment, helloService)

	// Configure deployment
	kubernetes.SetNamespace(deployment, "default")
	kubernetes.SetReplicas(deployment, 1)

	// Add to spec
	spec.AddNode(deployment)
}

func main() {
	// This main function is required for the wiring spec to compile
	// The actual wiring function (WireSpec, WireSpecWithNamespaces, or SimpleExample)
	// will be selected when running the Blueprint compiler
}
