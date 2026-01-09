package loadbalancer

import (
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/ir"
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/wiring"
)

// Creates a client-side load-balancer for multiple instances of a service. The list of services must be provided as an argument at compile-time when using this plugin.
func Create(spec wiring.WiringSpec, serviceGroupName string, services []string) string {
	loadbalancer_name := serviceGroupName + ".lb"
	spec.Define(loadbalancer_name, &LoadBalancerClient{}, func(namespace wiring.Namespace) (ir.IRNode, error) {
		var arg_nodes []ir.IRNode
		for _, arg_name := range services {
			var arg ir.IRNode
			if err := namespace.Get(arg_name, &arg); err != nil {
				return nil, err
			}
			arg_nodes = append(arg_nodes, arg)
		}

		return newLoadBalancerClient(serviceGroupName, arg_nodes)
	})

	dstName := loadbalancer_name + ".dst"
	spec.Alias(dstName, loadbalancer_name)

	return loadbalancer_name
}
