// Package rabbitmq provides a plugin to generate and include a rabbitmq instance in a Blueprint application.
//
// The package provides a built-in rabbitmq container that provides the server-side implementation
// and a go-client for connecting to the client.
//
// The applications must use a backend.Queue (runtime/core/backend) as the interface in the workflow.
package rabbitmq

import (
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/blueprint"
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/coreplugins/address"
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/coreplugins/pointer"
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/ir"
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/wiring"
)

// Container generate the IRNodes for a mysql server docker container that uses the latest mysql/mysql image
// and the clients needed by the generated application to communicate with the server.
func Container(spec wiring.WiringSpec, name string, queue_name string) string {
	// The nodes that we are defining
	ctrName := name + ".ctr"
	clientName := name + ".client"
	addrName := name + ".addr"

	// Define the rabbitmq container
	spec.Define(ctrName, &RabbitmqContainer{}, func(ns wiring.Namespace) (ir.IRNode, error) {
		ctr, err := newRabbitmqContainer(ctrName)
		if err != nil {
			return nil, err
		}

		err = address.Bind[*RabbitmqContainer](ns, addrName, ctr, &ctr.BindAddr)
		return ctr, err
	})

	// Create a pointer to the rabbitmq container
	ptr := pointer.CreatePointer[*RabbitmqGoClient](spec, name, ctrName)

	// Define the address that points to the Rabbitmq container
	address.Define[*RabbitmqContainer](spec, addrName, ctrName)
	ptr.AddAddrModifier(spec, addrName)

	// Define the Rabbitmq client and add it to the client side of the pointer
	clientNext := ptr.AddSrcModifier(spec, clientName)
	spec.Define(clientName, &RabbitmqGoClient{}, func(ns wiring.Namespace) (ir.IRNode, error) {
		addr, err := address.Dial[*RabbitmqContainer](ns, clientNext)
		if err != nil {
			return nil, blueprint.Errorf("%s expected %s to be an address but encountered %s", clientName, clientNext, err)
		}

		queue_val := &ir.IRValue{Value: queue_name}

		return newRabbitmqGoClient(clientName, addr.Dial, queue_val)
	})

	return name
}
