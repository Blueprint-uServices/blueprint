package loadbalancer

import (
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/blueprint"
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/core/pointer"
)

// Creates a client-side load-balancer for multiple instances of a service. The list of services must be provided as an argument at compile-time when using this plugin.
func Create(wiring blueprint.WiringSpec, services []string, serviceType string) string {
	loadbalancer_name := serviceType + ".lb"
	wiring.Define(loadbalancer_name, &LoadBalancerClient{}, func(namespace blueprint.Namespace) (blueprint.IRNode, error) {
		var arg_nodes []blueprint.IRNode
		for _, arg_name := range services {
			var arg blueprint.IRNode
			if err := namespace.Get(arg_name, &arg); err != nil {
				return nil, err
			}
			arg_nodes = append(arg_nodes, arg)
		}

		return newLoadBalancerClient(serviceType, arg_nodes)
	})

	dstName := loadbalancer_name + ".dst"
	wiring.Alias(dstName, loadbalancer_name)
	pointer.RequireUniqueness(wiring, dstName, &blueprint.ApplicationNode{})

	pointer.CreatePointer(wiring, loadbalancer_name, &LoadBalancerClient{}, dstName)
	return loadbalancer_name
}
