package address

import (
	"fmt"
	"strings"

	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/blueprint"
)

/*
This file contains some helper methods for working with address configs.

In general, server IRnodes will need to bind to a port, but for most
IRNodes this port is not pre-defined and can be assigned at deployment
time or even at runtime.

Addresses and ports are thus usually passed into nodes at runtime as
arguments.

For a few nodes -- primarily namespace nodes -- the nodes might need to
explicitly expose ports of servers running within the namespace.

Likewise, some namespace nodes might want to assign ports to the servers
running within the namespace.  In this case, each server needs its own
unique port.
*/

/*
A helper method for use by namespace nodes.

nodes -- IRnodes that exist within the namespace and/or the namespace receives as arguments

This method searches containedNodes and argNodes for server bind addresses
and assigns ports to any addresses that haven't yet been assigned.

Returns an error if multiple nodes have pre-assigned themselves conflicting ports
*/
func AssignPorts(nodes []blueprint.IRNode) error {
	// Extract the BindConfig nodes
	addrs := blueprint.Filter[*BindConfig](nodes)

	ports := make(map[uint16]*BindConfig)

	// Save any pre-assigned ports
	for _, addr := range addrs {
		if addr.Port != 0 {
			if other, conflict := ports[addr.Port]; conflict {
				return blueprint.Errorf("%v and %v both pre-assigned to port %v", other.Name(), addr.Name(), addr.Port)
			}
			ports[addr.Port] = addr
		}
	}

	// Assign preferred ports first
	for _, addr := range addrs {
		if addr.Port == 0 && addr.PreferredPort != 0 {
			if _, conflict := ports[addr.PreferredPort]; !conflict {
				addr.Port = addr.PreferredPort
				addr.Hostname = "0.0.0.0"
				ports[addr.Port] = addr
			}
		}
	}

	// Assign remaining ports
	for _, addr := range addrs {
		if addr.Port == 0 {
			candidatePort := addr.PreferredPort
			if candidatePort == 0 {
				candidatePort = 2000
			}
			for {
				if _, alreadyAssigned := ports[candidatePort]; !alreadyAssigned {
					addr.Port = candidatePort
					addr.Hostname = "0.0.0.0"
					ports[addr.Port] = addr
					break
				}
				candidatePort++
			}

		}
	}

	// Update preferred ports
	for _, addr := range addrs {
		addr.PreferredPort = addr.Port
	}

	for _, conf := range ports {
		fmt.Printf("assigned %v to port %v\n", conf.Name(), conf.Port)
	}
	return nil
}

/*
Returns an error if any ports haven't been allocated
*/
func CheckPorts(nodes []blueprint.IRNode) error {
	var missing []string
	for _, addr := range blueprint.Filter[*BindConfig](nodes) {
		if addr.Port == 0 {
			missing = append(missing, addr.Name())
		}
	}
	if len(missing) > 0 {
		return fmt.Errorf("unassigned bind addresses %v", strings.Join(missing, ", "))
	}
	return nil
}

/*
If a namespace translates addresses, then it will need to reset the assigned
ports before returning to the parent namespace
*/
func ResetPorts(nodes []blueprint.IRNode) {
	for _, addr := range blueprint.Filter[*BindConfig](nodes) {
		addr.Port = 0
		addr.Hostname = ""
	}
}
