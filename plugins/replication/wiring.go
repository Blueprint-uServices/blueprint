// Package replication creates replicas for the services defined in the application's workflow spec.
//
// # Wiring Spec Usage
//
// Example: To replicate a PaymentService with 5 replicas that depends on payment_cache and payment_db:
//
// replicas, balancer := Replicate[payment.PaymentService](spec, "payment_service", 5, "payment_cache", "payment_db")
//
// # Generated Artifacts
//
// Container for each replica instance and a load balancer that acts as the entrypoint for the replica group.
package replication

import (
	"fmt"

	"github.com/blueprint-uservices/blueprint/blueprint/pkg/wiring"
	"github.com/blueprint-uservices/blueprint/plugins/loadbalancer"
	"github.com/blueprint-uservices/blueprint/plugins/workflow"
)

// [Replicate] is used by wiring specs to replicate services from the workflow spec.
//
// Creates replicas for the given service and creates a load balancer in front of the replicas.
// The Replicate method automatically defines the services into the wiring spec.
//
// Type parameter [ServiceType] is used to specify the type of the service. It can be the name of an interface
// or an implementing struct. [ServiceType] must be a valid workflow service.
//
// `serviceName` is a unique name for the replicated service
//
// `numReplicas` is the total number of replicas that need to be deployed for this service.
//
// `serviceArgs` must correspond to the arguments of the service's constructor within the workflow spec.
// These args will be used for each replica.
//
// Returns the name of the all the replica instances and the name of the load balancer in front of the replicas.
func Replicate[ServiceType any](spec wiring.WiringSpec, serviceName string, numReplicas int, serviceArgs ...string) ([]string, string) {
	services := []string{}
	// Define the services in the workflow
	for i := 0; i < numReplicas; i++ {
		sname := fmt.Sprintf("%s_%d", serviceName, i)
		sname = workflow.Service[ServiceType](spec, sname, serviceArgs...)
		services = append(services, sname)
	}

	// Add a load balancer
	balancer := loadbalancer.CreateLoadBalancer[ServiceType](spec, serviceName, services)
	return services, balancer
}
