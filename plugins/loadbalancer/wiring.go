package loadbalancer

import (
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/coreplugins/pointer"
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/ir"
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/wiring"
)

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
