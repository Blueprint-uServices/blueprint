package replication

import (
	"fmt"

	"github.com/blueprint-uservices/blueprint/blueprint/pkg/wiring"
	"github.com/blueprint-uservices/blueprint/plugins/loadbalancer"
	"github.com/blueprint-uservices/blueprint/plugins/workflow"
)

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
