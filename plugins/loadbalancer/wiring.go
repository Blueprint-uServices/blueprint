// Package loadbalancer creates a loadbalancer instance in front of the service instance replicas.
//
// # Wiring Spec Usage
//
// Example: To add a load balancer in front of 3 instances of PaymentService:
//
// balancer := CreateLoadBalancer[payment.PaymentService](spec, "payment_service", []string{"payment_service1", "payment_service2", "payment_service3"})
//
// # Generated Artifacts
//
// Container for a load balancer instance that acts as the entrypoint for the replica instances.
package loadbalancer

import (
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/coreplugins/pointer"
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/ir"
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/wiring"
)

// [CreateLoadBalancer] is used by wiring specs to add a load balancer instance in front of replica instances for a specific service group.
//
// Creates a load balancer instance.
// All the service instances must be previously defined as part of the wiring spec.
//
// Type parameter [ServiceType] is used to specify the type of the service. It can be the name of an interface or an implementing struct. [ServiceType] must be a valid workflow service.
// [ServiceType] should be the same as the type to define the workflow instances.
//
// Returns the name of the load balancer instance created and added to the wiring spec.
func CreateLoadBalancer[ServiceType any](spec wiring.WiringSpec, serviceGroupName string, serviceNames []string) string {
	balancerName := serviceGroupName + "_lb"
	handlerName := balancerName + ".handler"
	clientName := balancerName + ".client"

	spec.Define(handlerName, &dynamicLBHandler{}, func(ns wiring.Namespace) (ir.IRNode, error) {
		handler := &dynamicLBHandler{}
		if err := initDynamicLBNode[ServiceType](&handler.dynamicLBNode, balancerName); err != nil {
			return nil, err
		}

		args := make([]ir.IRNode, len(serviceNames))
		for i, name := range serviceNames {
			if err := ns.Get(name, &args[i]); err != nil {
				return nil, err
			}
		}
		if err := handler.Init(args); err != nil {
			return nil, err
		}

		return handler, nil
	})

	ptr := pointer.CreatePointer[*dynamicLBNode](spec, balancerName, handlerName)

	clientNext := ptr.AddSrcModifier(spec, clientName)
	spec.Define(clientName, &dynamicLBClient{}, func(ns wiring.Namespace) (ir.IRNode, error) {
		client := &dynamicLBClient{}
		if err := initDynamicLBNode[ServiceType](&client.dynamicLBNode, balancerName); err != nil {
			return nil, err
		}
		return client, ns.Get(clientNext, &client.Wrapped)
	})

	return balancerName
}
